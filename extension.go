// Copyright 2025 The entmcp Authors. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package entmcp

import (
	"encoding/json"
	"fmt"

	"entgo.io/ent/entc"
	"entgo.io/ent/entc/gen"
)

// Extension is the entc.Extension implementation for entmcp.
// It generates a Model Context Protocol (MCP) server from an Ent schema.
type Extension struct {
	entc.DefaultExtension
	config *Config
}

// NewExtension creates a new entmcp Extension with the given options.
func NewExtension(opts ...Option) (*Extension, error) {
	cfg := &Config{}
	for _, opt := range opts {
		opt(cfg)
	}
	cfg.defaults()
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return &Extension{config: cfg}, nil
}

// Annotations returns the global annotations contributed by the extension.
func (e *Extension) Annotations() []entc.Annotation {
	return []entc.Annotation{e.config}
}

// Hooks returns the code generation hooks.
func (e *Extension) Hooks() []gen.Hook {
	return []gen.Hook{
		generateMCPHook(e.config),
	}
}

// Templates returns nil; this extension uses hooks for code generation.
func (e *Extension) Templates() []*gen.Template {
	return nil
}

// Name implements the schema.Annotation interface for the Config annotation.
func (c *Config) Name() string { return "EntMCPConfig" }

// annotationFromGraph retrieves the EntMCPConfig annotation from the graph.
func annotationFromGraph(g *gen.Graph) (*Config, error) {
	v, ok := g.Config.Annotations["EntMCPConfig"]
	if !ok {
		return nil, fmt.Errorf("entmcp: EntMCPConfig annotation not found in graph")
	}
	b, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("entmcp: marshal config annotation: %w", err)
	}
	cfg := &Config{}
	if err := json.Unmarshal(b, cfg); err != nil {
		return nil, fmt.Errorf("entmcp: unmarshal config annotation: %w", err)
	}
	cfg.defaults()
	return cfg, nil
}
