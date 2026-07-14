// Copyright 2025 The entmcp Authors. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package entmcp_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"entgo.io/ent/entc"
	"entgo.io/ent/entc/gen"
	"github.com/smintz/entmcp"
	"github.com/stretchr/testify/require"

	// Pin the MCP SDK in go.mod/go.sum: the generated code imports it, but no
	// hand-written source in this module does, so `go mod tidy` would otherwise
	// drop it and break TestGeneratedCodeCompiles.
	_ "github.com/modelcontextprotocol/go-sdk/mcp"
)

// TestGeneratedCodeCompiles is the regression guard that the rest of the suite
// lacked: it does not merely check that generation succeeds, it compiles the
// generated output with `go build`. Every historical codegen bug (missing
// imports, wrong initialism casing, pointer/value setter mismatches, int64
// predicate casts, wrong ent import) manifested only at compile time and so
// slipped past generation-only tests.
//
// The output is generated into a real, module-resolvable package path so the
// build can resolve ent, the MCP SDK, and the generated ent package.
func TestGeneratedCodeCompiles(t *testing.T) {
	cwd, err := os.Getwd()
	require.NoError(t, err)

	const pkgDir = "testcompile"
	genDir := filepath.Join(cwd, pkgDir)
	require.NoError(t, os.RemoveAll(genDir))
	t.Cleanup(func() { _ = os.RemoveAll(genDir) })

	ext, err := entmcp.NewExtension(
		entmcp.WithToolPrefix("test"),
		entmcp.WithDefaultOperations(entmcp.OpAll),
	)
	require.NoError(t, err)

	cfg := &gen.Config{
		Target:  genDir,
		Package: "github.com/smintz/entmcp/" + pkgDir,
		Schema:  "testdata/schema",
	}
	err = entc.Generate("./testdata/schema", cfg, entc.Extensions(ext))
	require.NoError(t, err, "entc.Generate should not error")

	// The actual guard: the generated package must build.
	build := exec.Command("go", "build", "./"+pkgDir+"/...")
	build.Dir = cwd
	out, err := build.CombinedOutput()
	require.NoErrorf(t, err, "generated code must compile; go build output:\n%s", out)

	// Guard against the %!s(MISSING) format-string bug: the generator must not
	// leave unresolved verbs baked into emitted error messages.
	helpers, err := os.ReadFile(filepath.Join(genDir, "entmcp", "helpers.go"))
	require.NoError(t, err)
	require.NotContains(t, string(helpers), "%!s(MISSING)",
		"mapEntError must not contain unescaped format verbs")

	// Spot-check that ent's initialism-aware names are used (external_id ->
	// ExternalID, not ExternalId).
	userTools, err := os.ReadFile(filepath.Join(genDir, "entmcp", "user_tools.go"))
	require.NoError(t, err)
	src := string(userTools)
	require.Contains(t, src, "ExternalID", "must use ent's initialism-aware Go name")
	require.False(t, strings.Contains(src, "ExternalId("),
		"must not emit naive-PascalCase ExternalId")
}
