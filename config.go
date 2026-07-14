// Copyright 2025 The entmcp Authors. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package entmcp

import (
	"fmt"
	"regexp"
)

// Operations is a bitmask of allowed CRUD operations.
type Operations uint8

const (
	// OpCreate enables the create_<entity> tool.
	OpCreate Operations = 1 << iota
	// OpRead enables the get_<entity> tool.
	OpRead
	// OpList enables the list_<entities> tool.
	OpList
	// OpUpdate enables the update_<entity> tool.
	OpUpdate
	// OpDelete enables the delete_<entity> tool.
	OpDelete
	// OpAll enables all operations.
	OpAll Operations = OpCreate | OpRead | OpList | OpUpdate | OpDelete
)

// Has reports whether op includes the given operation.
func (o Operations) Has(op Operations) bool { return o&op != 0 }

// Config holds the configuration for the entmcp extension.
type Config struct {
	// ToolPrefix is prepended to all tool names. Must match ^[a-z][a-z0-9_]*$ or be empty.
	ToolPrefix string
	// DefaultOperations is the bitmask of enabled operations for all entities.
	// Defaults to OpAll.
	DefaultOperations Operations
	// ReadOnly disables Create/Update/Delete everywhere when true.
	ReadOnly bool
	// MaxPageSize is the maximum allowed page size. Defaults to 100.
	MaxPageSize int
	// DefaultPageSize is the default page size. Defaults to 25.
	DefaultPageSize int
	// EnableResources causes MCP resources (ent://entity/{id}) to be generated.
	EnableResources bool
	// EnableEdgeTools causes list_<entity>_<edge> traversal tools to be generated.
	// Defaults to true.
	EnableEdgeTools bool
	// SpecPath is the optional path to write a JSON manifest of all generated tools.
	SpecPath string
	// DefaultEagerLoad controls whether unique edges are embedded in responses by default.
	DefaultEagerLoad bool
}

var toolPrefixRegexp = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

// validate returns an error if the config is invalid.
func (c *Config) validate() error {
	if c.ToolPrefix != "" && !toolPrefixRegexp.MatchString(c.ToolPrefix) {
		return fmt.Errorf("entmcp: ToolPrefix %q must match ^[a-z][a-z0-9_]*$", c.ToolPrefix)
	}
	if c.DefaultPageSize > c.MaxPageSize {
		return fmt.Errorf("entmcp: DefaultPageSize (%d) must be <= MaxPageSize (%d)", c.DefaultPageSize, c.MaxPageSize)
	}
	return nil
}

// defaults fills in zero values with sensible defaults.
func (c *Config) defaults() {
	if c.DefaultOperations == 0 {
		c.DefaultOperations = OpAll
	}
	if c.MaxPageSize == 0 {
		c.MaxPageSize = 100
	}
	if c.DefaultPageSize == 0 {
		c.DefaultPageSize = 25
	}
}

// Option is a functional option for configuring the entmcp extension.
type Option func(*Config)

// WithToolPrefix sets the prefix for all generated tool names.
func WithToolPrefix(prefix string) Option {
	return func(c *Config) { c.ToolPrefix = prefix }
}

// WithDefaultOperations sets the default set of enabled operations.
func WithDefaultOperations(ops Operations) Option {
	return func(c *Config) { c.DefaultOperations = ops }
}

// WithReadOnly disables all write operations (Create/Update/Delete).
func WithReadOnly(ro bool) Option {
	return func(c *Config) { c.ReadOnly = ro }
}

// WithMaxPageSize sets the maximum allowed page size.
func WithMaxPageSize(n int) Option {
	return func(c *Config) { c.MaxPageSize = n }
}

// WithDefaultPageSize sets the default page size.
func WithDefaultPageSize(n int) Option {
	return func(c *Config) { c.DefaultPageSize = n }
}

// WithResources enables MCP resource generation.
func WithResources(enable bool) Option {
	return func(c *Config) { c.EnableResources = enable }
}

// WithEdgeTools enables or disables edge traversal tools.
func WithEdgeTools(enable bool) Option {
	return func(c *Config) { c.EnableEdgeTools = enable }
}

// WithSpecPath sets the path to write the JSON tool manifest.
func WithSpecPath(path string) Option {
	return func(c *Config) { c.SpecPath = path }
}

// WithDefaultEagerLoad controls whether unique edges are eagerly loaded by default.
func WithDefaultEagerLoad(enable bool) Option {
	return func(c *Config) { c.DefaultEagerLoad = enable }
}
