# entmcp

[![CI](https://github.com/smintz/entmcp/actions/workflows/ci.yml/badge.svg)](https://github.com/smintz/entmcp/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/smintz/entmcp.svg)](https://pkg.go.dev/github.com/smintz/entmcp)

**entmcp** is a [Go code-generation extension](https://entgo.io/docs/code-gen#external-templates) for [entgo.io/ent](https://entgo.io) that generates a fully functional [Model Context Protocol (MCP)](https://modelcontextprotocol.io) server from your Ent schema.

Given an Ent schema, entmcp generates:
- **MCP tool definitions** for each entity — create / get / list (with filtering, sorting, pagination) / update / delete
- **Tool handler implementations** that execute against a `*ent.Client`, honouring Ent hooks, interceptors, and privacy policies
- **A server assembly function** `NewMCPServer(client, opts...)` that registers all generated tools; you choose the transport (stdio, HTTP)

## Compatibility

| entmcp | entgo.io/ent | MCP Go SDK | Go  |
|--------|-------------|------------|-----|
| v0.x   | v0.14.x     | v1.x       | 1.24+ |

The generated server targets the v1.x MCP Go SDK API
(`mcp.NewServer(&mcp.Implementation{...}, nil)`).

## Installation

```bash
go get github.com/smintz/entmcp
```

## Quickstart

### 1. Create a code-generator file

In your `ent/` directory, create `entc.go`:

```go
//go:build ignore

package main

import (
	"log"

	"entgo.io/ent/entc"
	"entgo.io/ent/entc/gen"
	"github.com/smintz/entmcp"
)

func main() {
	ext, err := entmcp.NewExtension(
		entmcp.WithToolPrefix("myapp"),             // tools: myapp_create_user, ...
		entmcp.WithDefaultOperations(entmcp.OpAll), // or OpRead|OpList for read-only
		entmcp.WithMaxPageSize(100),
		entmcp.WithDefaultPageSize(25),
	)
	if err != nil {
		log.Fatal(err)
	}
	if err := entc.Generate("./schema", &gen.Config{}, entc.Extensions(ext)); err != nil {
		log.Fatalf("running ent codegen: %v", err)
	}
}
```

### 2. Generate code

```bash
go run ent/entc.go
```

This generates `ent/entmcp/` containing:
- `entmcp.go` — `NewMCPServer` and `ServerOption` types
- `helpers.go` — shared utility functions
- `{entity}_tools.go` — per-entity tool definitions and handlers

### 3. Wire the server (stdio transport)

```go
package main

import (
	"context"
	"log"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"yourmodule/ent"
	"yourmodule/ent/entmcp"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	client, err := ent.Open("sqlite3", "file:app.db?cache=shared&_fk=1")
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// Run auto-migration.
	if err := client.Schema.Create(context.Background()); err != nil {
		log.Fatal(err)
	}

	srv := entmcp.NewMCPServer(client,
		entmcp.WithServerInfo("myapp", "v1"),
	)

	if err := srv.Run(context.Background(), mcp.NewStdioTransport()); err != nil {
		log.Fatal(err)
	}
}
```

## Configuration options

| Option | Default | Description |
|--------|---------|-------------|
| `WithToolPrefix(string)` | `""` | Prefix for all tool names (e.g. `"myapp"` → `myapp_create_user`) |
| `WithDefaultOperations(Operations)` | `OpAll` | Enable only specific operations globally |
| `WithReadOnly(bool)` | `false` | Disable all write operations (create/update/delete) |
| `WithMaxPageSize(int)` | `100` | Maximum page size for list operations |
| `WithDefaultPageSize(int)` | `25` | Default page size for list operations |
| `WithResources(bool)` | `false` | Also generate MCP resources (`ent://user/{id}`) |
| `WithEdgeTools(bool)` | `true` | Generate edge traversal tools (`list_user_posts`) |
| `WithSpecPath(string)` | `""` | Write a JSON manifest of all tools to this path |
| `WithDefaultEagerLoad(bool)` | `false` | Embed unique edges in responses by default |

## Per-entity annotations

Use annotations in your schema to customise MCP behaviour:

```go
func (User) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entmcp.Operations(entmcp.OpRead | entmcp.OpList), // override ops for this entity
		entmcp.DescriptionAnnotation("A registered user"),
	}
}

func (User) Fields() []ent.Field {
	return []ent.Field{
		field.String("api_key").
			Sensitive(). // automatically WriteOnly — never in responses
			Optional(),
		field.Time("created_at").
			Immutable(). // excluded from update input
			Default(time.Now),
	}
}
```

Available annotations:

| Constructor | Effect |
|-------------|--------|
| `entmcp.Skip()` | Exclude entity/field/edge from MCP surface |
| `entmcp.Operations(ops)` | Override enabled operations for this entity |
| `entmcp.ReadOnlyAnnotation()` | Field only appears in responses (not in create/update input) |
| `entmcp.WriteOnlyAnnotation()` | Field only in input; omitted from responses |
| `entmcp.DescriptionAnnotation("...")` | Override description in tool schema |
| `entmcp.FilterableAnnotation(false)` | Opt this field out of filter generation |
| `entmcp.SortableAnnotation(false)` | Opt this field out of sort generation |
| `entmcp.EagerLoadAnnotation(true)` | Embed this edge in get/list responses |

## Generated tools

For an entity named `User` (with empty prefix):

| Tool | Description | Hints |
|------|-------------|-------|
| `create_user` | Insert a new User | `readOnly:false` |
| `get_user` | Fetch by ID | `readOnly:true` |
| `list_users` | Filter + sort + paginate | `readOnly:true` |
| `update_user` | Partial update by ID | `readOnly:false, idempotent:true` |
| `delete_user` | Delete by ID | `readOnly:false, idempotent:true, destructive:true` |

### list_* input contract

```jsonc
{
  "filter":   { "name": "Alice", "age": 30 },
  "order_by": [{ "field": "created_at", "direction": "desc" }],
  "page":     { "size": 25, "number": 1 }
}
```

## Runtime server options

```go
entmcp.NewMCPServer(client,
	entmcp.WithServerInfo("myapp", "v1.2.3"),
	entmcp.WithLogger(slog.Default()),
	entmcp.WithToolFilter(func(name string) bool {
		return name != "delete_user" // disable at runtime
	}),
	entmcp.WithContextFunc(func(ctx context.Context) context.Context {
		return privacy.NewContext(ctx, viewer) // inject privacy viewer
	}),
)
```

## Examples

See [`_examples/kitchensink/`](./_examples/kitchensink/) for a complete schema + codegen setup.

## License

MIT
