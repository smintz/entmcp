// Copyright 2025 The entmcp Authors. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package entmcp

import (
	"encoding/json"
	"fmt"

	"entgo.io/ent/entc/gen"
	entschema "entgo.io/ent/schema"
)

// Annotation is the entmcp schema annotation. It can be applied at the schema,
// field, and edge level. Per-field and per-edge settings take precedence over
// the schema-level setting, which in turn takes precedence over the global Config.
type Annotation struct {
	// Skip excludes the entity/field/edge entirely from the MCP surface.
	Skip bool `json:"Skip,omitempty"`
	// Operations overrides the enabled operations for this entity.
	Operations *Operations `json:"Operations,omitempty"`
	// ReadOnly marks a field as accepted in responses but rejected in create/update input.
	ReadOnly bool `json:"ReadOnly,omitempty"`
	// WriteOnly marks a field as accepted in input but omitted from responses.
	WriteOnly bool `json:"WriteOnly,omitempty"`
	// Description overrides the tool/field description.
	Description string `json:"Description,omitempty"`
	// ToolName overrides the base tool name for this entity (else snake_case of type name).
	ToolName string `json:"ToolName,omitempty"`
	// Filterable opts the field in or out of filter generation.
	Filterable *bool `json:"Filterable,omitempty"`
	// Sortable opts the field in or out of sort generation.
	Sortable *bool `json:"Sortable,omitempty"`
	// EagerLoad opts an edge in or out of eager loading in get/list responses.
	EagerLoad *bool `json:"EagerLoad,omitempty"`
	// Example is a JSON Schema example value for this field.
	Example any `json:"Example,omitempty"`
}

// Name implements the schema.Annotation interface.
func (Annotation) Name() string { return "EntMCP" }

// Merge implements the schema.Merger interface for combining annotations.
func (a Annotation) Merge(other entschema.Annotation) entschema.Annotation {
	b, ok := other.(Annotation)
	if !ok {
		return a
	}
	// Fields in b take precedence over a.
	if b.Skip {
		a.Skip = true
	}
	if b.Operations != nil {
		a.Operations = b.Operations
	}
	if b.ReadOnly {
		a.ReadOnly = true
	}
	if b.WriteOnly {
		a.WriteOnly = true
	}
	if b.Description != "" {
		a.Description = b.Description
	}
	if b.ToolName != "" {
		a.ToolName = b.ToolName
	}
	if b.Filterable != nil {
		a.Filterable = b.Filterable
	}
	if b.Sortable != nil {
		a.Sortable = b.Sortable
	}
	if b.EagerLoad != nil {
		a.EagerLoad = b.EagerLoad
	}
	if b.Example != nil {
		a.Example = b.Example
	}
	return a
}

// annotationFromType decodes the EntMCP annotation from a gen.Type.
// Returns a zero-value Annotation if none is set.
func annotationFromType(t *gen.Type) (Annotation, error) {
	return decodeAnnotation(t.Annotations)
}

// annotationFromField decodes the EntMCP annotation from a gen.Field.
func annotationFromField(f *gen.Field) (Annotation, error) {
	return decodeAnnotation(f.Annotations)
}

// annotationFromEdge decodes the EntMCP annotation from a gen.Edge.
func annotationFromEdge(e *gen.Edge) (Annotation, error) {
	return decodeAnnotation(e.Annotations)
}

func decodeAnnotation(annotations gen.Annotations) (Annotation, error) {
	v, ok := annotations["EntMCP"]
	if !ok {
		return Annotation{}, nil
	}
	b, err := json.Marshal(v)
	if err != nil {
		return Annotation{}, fmt.Errorf("entmcp: marshal annotation: %w", err)
	}
	var a Annotation
	if err := json.Unmarshal(b, &a); err != nil {
		return Annotation{}, fmt.Errorf("entmcp: unmarshal annotation: %w", err)
	}
	return a, nil
}

// effectiveOps returns the effective operations for a type, merging global config
// with schema-level annotation.
func effectiveOps(cfg *Config, typeAnt Annotation) Operations {
	ops := cfg.DefaultOperations
	if typeAnt.Operations != nil {
		ops = *typeAnt.Operations
	}
	if cfg.ReadOnly {
		ops &^= (OpCreate | OpUpdate | OpDelete)
	}
	return ops
}

// Helper constructors for common annotation patterns.

// Skip returns an Annotation that skips the entity/field/edge.
func Skip() Annotation { return Annotation{Skip: true} }

// OperationsAnnotation returns an Annotation that sets the operations bitmask.
func OperationsAnnotation(ops Operations) Annotation { return Annotation{Operations: &ops} }

// ReadOnlyAnnotation returns an Annotation that marks the field as read-only.
func ReadOnlyAnnotation() Annotation { return Annotation{ReadOnly: true} }

// WriteOnlyAnnotation returns an Annotation that marks the field as write-only.
func WriteOnlyAnnotation() Annotation { return Annotation{WriteOnly: true} }

// DescriptionAnnotation returns an Annotation with a custom description.
func DescriptionAnnotation(desc string) Annotation { return Annotation{Description: desc} }

// FilterableAnnotation returns an Annotation that opts the field in or out of filter generation.
func FilterableAnnotation(filterable bool) Annotation { return Annotation{Filterable: &filterable} }

// SortableAnnotation returns an Annotation that opts the field in or out of sort generation.
func SortableAnnotation(sortable bool) Annotation { return Annotation{Sortable: &sortable} }

// EagerLoadAnnotation returns an Annotation that opts the edge in or out of eager loading.
func EagerLoadAnnotation(eagerLoad bool) Annotation { return Annotation{EagerLoad: &eagerLoad} }

var _ entschema.Annotation = Annotation{}
var _ entschema.Merger = Annotation{}
