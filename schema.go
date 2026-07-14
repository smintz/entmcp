// Copyright 2025 The entmcp Authors. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package entmcp

import (
	"fmt"
	"sort"
	"strings"
	"unicode"

	"entgo.io/ent/entc/gen"
	"entgo.io/ent/schema/field"
)

// JSONSchema represents a JSON Schema object (subset needed for MCP tool schemas).
type JSONSchema struct {
	Type             string                 `json:"type,omitempty"`
	Format           string                 `json:"format,omitempty"`
	Enum             []string               `json:"enum,omitempty"`
	Properties       map[string]*JSONSchema `json:"properties,omitempty"`
	Required         []string               `json:"required,omitempty"`
	Items            *JSONSchema            `json:"items,omitempty"`
	ContentEncoding  string                 `json:"contentEncoding,omitempty"`
	Description      string                 `json:"description,omitempty"`
	Example          any                    `json:"example,omitempty"`
	AdditionalProps  *bool                  `json:"additionalProperties,omitempty"`
	Ref              string                 `json:"$ref,omitempty"`
	Defs             map[string]*JSONSchema `json:"$defs,omitempty"`
}

// FieldSpec describes a single field for code generation.
type FieldSpec struct {
	// GoName is the Go struct field name (PascalCase).
	GoName string
	// JSONName is the JSON field name (snake_case).
	JSONName string
	// GoType is the Go type string for the field.
	GoType string
	// GoTypeNillable is true if the field is a pointer type.
	GoTypeNillable bool
	// Required is true if the field is required in create input.
	Required bool
	// Immutable is true if the field cannot be updated.
	Immutable bool
	// Sensitive is true if the field should never appear in output.
	Sensitive bool
	// ReadOnly is true if the field is only in responses.
	ReadOnly bool
	// WriteOnly is true if the field is only in input.
	WriteOnly bool
	// Description is the field description.
	Description string
	// Schema is the JSON Schema for this field.
	Schema *JSONSchema
	// FilterOps is the list of filter operations for this field.
	FilterOps []string
	// Sortable indicates this field can be used in order_by.
	Sortable bool
	// IsEnum indicates this is an enum field.
	IsEnum bool
	// EnumValues are the possible enum values.
	EnumValues []string
	// FieldType is the underlying Ent field type.
	FieldType field.Type
}

// EdgeSpec describes a single edge for code generation.
type EdgeSpec struct {
	// GoName is the Go field name.
	GoName string
	// JSONName is the JSON field name.
	JSONName string
	// TypeName is the related entity type name.
	TypeName string
	// Unique is true for O2O and M2O edges.
	Unique bool
	// EagerLoad controls whether this edge is embedded in get/list responses.
	EagerLoad bool
	// Skip excludes this edge from MCP surface.
	Skip bool
}

// ToolSpec is the full view model for one entity type, used for code generation.
type ToolSpec struct {
	// TypeName is the Go type name (e.g. "User").
	TypeName string
	// PackageName is the ent subpackage name (e.g. "user").
	PackageName string
	// IDField is the ID field spec.
	IDField FieldSpec
	// Fields contains all non-ID fields.
	Fields []FieldSpec
	// Edges contains all edges.
	Edges []EdgeSpec
	// Ops is the set of enabled operations.
	Ops Operations
	// BaseToolName is the snake_case entity name used in tool names.
	BaseToolName string
	// CreateToolName is the create tool name.
	CreateToolName string
	// GetToolName is the get tool name.
	GetToolName string
	// ListToolName is the list tool name.
	ListToolName string
	// UpdateToolName is the update tool name.
	UpdateToolName string
	// DeleteToolName is the delete tool name.
	DeleteToolName string
	// Description is the entity description.
	Description string
}

// buildToolSpecs constructs ToolSpec view models from the gen.Graph.
func buildToolSpecs(g *gen.Graph, cfg *Config) ([]ToolSpec, error) {
	// Collect all tool names to detect collisions.
	toolNames := make(map[string]string) // tool name -> type name

	var specs []ToolSpec
	// Sort nodes for deterministic output.
	nodes := make([]*gen.Type, len(g.Nodes))
	copy(nodes, g.Nodes)
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].Name < nodes[j].Name
	})

	for _, t := range nodes {
		typeAnt, err := annotationFromType(t)
		if err != nil {
			return nil, fmt.Errorf("entmcp: type %q: %w", t.Name, err)
		}
		if typeAnt.Skip {
			continue
		}

		ops := effectiveOps(cfg, typeAnt)
		base := toolName(t.Name, typeAnt)

		// Build tool names and check for collisions.
		toolNamesForType := map[string]string{
			createToolName(cfg.ToolPrefix, base): "create",
			getToolName(cfg.ToolPrefix, base):    "get",
			listToolName(cfg.ToolPrefix, base):   "list",
			updateToolName(cfg.ToolPrefix, base): "update",
			deleteToolName(cfg.ToolPrefix, base): "delete",
		}
		for tn, op := range toolNamesForType {
			if existing, ok := toolNames[tn]; ok {
				return nil, fmt.Errorf("entmcp: tool name collision: %q (from %s op on %s) conflicts with existing tool from %s", tn, op, t.Name, existing)
			}
			toolNames[tn] = t.Name
		}

		// Build field specs.
		idField, err := buildFieldSpec(t.ID, cfg)
		if err != nil {
			return nil, fmt.Errorf("entmcp: type %q ID: %w", t.Name, err)
		}
		idField.Required = true

		var fields []FieldSpec
		// Sort fields for deterministic output.
		sortedFields := make([]*gen.Field, len(t.Fields))
		copy(sortedFields, t.Fields)
		sort.Slice(sortedFields, func(i, j int) bool {
			return sortedFields[i].Name < sortedFields[j].Name
		})

		for _, f := range sortedFields {
			fAnt, err := annotationFromField(f)
			if err != nil {
				return nil, fmt.Errorf("entmcp: type %q field %q: %w", t.Name, f.Name, err)
			}
			if fAnt.Skip {
				continue
			}
			fs, err := buildFieldSpec(f, cfg)
			if err != nil {
				// Skip unsupported field types with a warning.
				continue
			}
			// Apply field-level annotation overrides.
			if fAnt.ReadOnly {
				fs.ReadOnly = true
			}
			if fAnt.WriteOnly || f.Sensitive() {
				fs.WriteOnly = true
				fs.Sensitive = f.Sensitive()
				// Sensitive/WriteOnly fields must not appear in filter or sort contexts.
				fs.FilterOps = nil
				fs.Sortable = false
			}
			if fAnt.Description != "" {
				fs.Description = fAnt.Description
			}
			if fAnt.Filterable != nil {
				if !*fAnt.Filterable {
					fs.FilterOps = nil
				}
			}
			if fAnt.Sortable != nil {
				fs.Sortable = *fAnt.Sortable
			}
			if fAnt.Example != nil {
				fs.Schema.Example = fAnt.Example
			}
			fields = append(fields, fs)
		}

		// Build edge specs.
		var edges []EdgeSpec
		sortedEdges := make([]*gen.Edge, len(t.Edges))
		copy(sortedEdges, t.Edges)
		sort.Slice(sortedEdges, func(i, j int) bool {
			return sortedEdges[i].Name < sortedEdges[j].Name
		})
		for _, e := range sortedEdges {
			eAnt, err := annotationFromEdge(e)
			if err != nil {
				return nil, fmt.Errorf("entmcp: type %q edge %q: %w", t.Name, e.Name, err)
			}
			es := EdgeSpec{
				GoName:   pascal(e.Name),
				JSONName: toSnakeCase(e.Name),
				TypeName: e.Type.Name,
				Unique:   e.Unique,
				Skip:     eAnt.Skip,
			}
			// Determine eager load setting.
			if eAnt.EagerLoad != nil {
				es.EagerLoad = *eAnt.EagerLoad
			} else if cfg.DefaultEagerLoad && e.Unique {
				es.EagerLoad = true
			}
			edges = append(edges, es)
		}

		// Get entity description from schema comment annotation.
		desc := typeAnt.Description

		spec := ToolSpec{
			TypeName:       t.Name,
			PackageName:    t.PackageDir(),
			IDField:        idField,
			Fields:         fields,
			Edges:          edges,
			Ops:            ops,
			BaseToolName:   base,
			CreateToolName: createToolName(cfg.ToolPrefix, base),
			GetToolName:    getToolName(cfg.ToolPrefix, base),
			ListToolName:   listToolName(cfg.ToolPrefix, base),
			UpdateToolName: updateToolName(cfg.ToolPrefix, base),
			DeleteToolName: deleteToolName(cfg.ToolPrefix, base),
			Description:    desc,
		}
		specs = append(specs, spec)
	}
	return specs, nil
}

// buildFieldSpec constructs a FieldSpec from a gen.Field.
func buildFieldSpec(f *gen.Field, cfg *Config) (FieldSpec, error) {
	schema, err := fieldJSONSchema(f)
	if err != nil {
		return FieldSpec{}, err
	}

	desc := f.Comment()
	if desc != "" && schema.Description == "" {
		schema.Description = desc
	}

	fs := FieldSpec{
		GoName:         pascal(f.Name),
		JSONName:       f.Name,
		GoType:         goType(f),
		GoTypeNillable: f.Nillable || f.Optional,
		Required:       !f.Optional && !f.Default && !f.Nillable,
		Immutable:      f.Immutable,
		Sensitive:      f.Sensitive(),
		Description:    desc,
		Schema:         schema,
		FieldType:      f.Type.Type,
	}

	// Determine filter operations.
	fs.FilterOps = filterOps(f)

	// Determine sortability.
	fs.Sortable = isSortable(f)

	// Enum values.
	if f.Type.Type == field.TypeEnum {
		fs.IsEnum = true
		for _, e := range f.Enums {
			fs.EnumValues = append(fs.EnumValues, e.Value)
		}
	}

	return fs, nil
}

// fieldJSONSchema returns the JSON Schema for a gen.Field.
func fieldJSONSchema(f *gen.Field) (*JSONSchema, error) {
	t := f.Type.Type
	switch t {
	case field.TypeString:
		return &JSONSchema{Type: "string"}, nil
	case field.TypeInt, field.TypeInt8, field.TypeInt16, field.TypeInt32, field.TypeInt64,
		field.TypeUint, field.TypeUint8, field.TypeUint16, field.TypeUint32, field.TypeUint64:
		return &JSONSchema{Type: "integer"}, nil
	case field.TypeFloat32, field.TypeFloat64:
		return &JSONSchema{Type: "number"}, nil
	case field.TypeBool:
		return &JSONSchema{Type: "boolean"}, nil
	case field.TypeTime:
		return &JSONSchema{Type: "string", Format: "date-time"}, nil
	case field.TypeEnum:
		vals := make([]string, len(f.Enums))
		for i, e := range f.Enums {
			vals[i] = e.Value
		}
		return &JSONSchema{Type: "string", Enum: vals}, nil
	case field.TypeUUID:
		return &JSONSchema{Type: "string", Format: "uuid"}, nil
	case field.TypeBytes:
		return &JSONSchema{Type: "string", ContentEncoding: "base64"}, nil
	case field.TypeJSON:
		// Check if it's a slice type.
		if f.Type.Nillable || isSliceType(f) {
			return &JSONSchema{Type: "array"}, nil
		}
		return &JSONSchema{Type: "object"}, nil
	default:
		return nil, fmt.Errorf("unsupported field type: %v", t)
	}
}

// isSliceType reports whether the field's Go type is a slice.
func isSliceType(f *gen.Field) bool {
	ident := f.Type.Ident
	return len(ident) > 0 && ident[0] == '['
}

// goType returns the Go type string for a field.
func goType(f *gen.Field) string {
	t := f.Type.Type
	var base string
	switch t {
	case field.TypeString:
		base = "string"
	case field.TypeInt:
		base = "int"
	case field.TypeInt8:
		base = "int8"
	case field.TypeInt16:
		base = "int16"
	case field.TypeInt32:
		base = "int32"
	case field.TypeInt64:
		base = "int64"
	case field.TypeUint:
		base = "uint"
	case field.TypeUint8:
		base = "uint8"
	case field.TypeUint16:
		base = "uint16"
	case field.TypeUint32:
		base = "uint32"
	case field.TypeUint64:
		base = "uint64"
	case field.TypeFloat32:
		base = "float32"
	case field.TypeFloat64:
		base = "float64"
	case field.TypeBool:
		base = "bool"
	case field.TypeTime:
		base = "time.Time"
	case field.TypeEnum:
		// Use the ent-generated enum type if available.
		if f.Type.Ident != "" {
			base = f.Type.Ident
		} else {
			base = "string"
		}
	case field.TypeUUID:
		base = "uuid.UUID"
	case field.TypeBytes:
		base = "[]byte"
	case field.TypeJSON:
		if f.Type.Ident != "" {
			base = f.Type.Ident
		} else {
			base = "json.RawMessage"
		}
	default:
		base = "any"
	}
	if f.Nillable {
		return "*" + base
	}
	return base
}

// filterOps returns the filter operations applicable to this field type.
func filterOps(f *gen.Field) []string {
	t := f.Type.Type
	var ops []string
	switch t {
	case field.TypeInt, field.TypeInt8, field.TypeInt16, field.TypeInt32, field.TypeInt64,
		field.TypeUint, field.TypeUint8, field.TypeUint16, field.TypeUint32, field.TypeUint64,
		field.TypeFloat32, field.TypeFloat64, field.TypeTime:
		ops = []string{"eq", "neq", "gt", "gte", "lt", "lte", "in", "not_in"}
		if f.Optional || f.Nillable {
			ops = append(ops, "is_nil")
		}
	case field.TypeString:
		ops = []string{"contains", "eq", "equal_fold", "has_prefix", "has_suffix", "in", "neq", "not_in"}
		if f.Optional || f.Nillable {
			ops = append(ops, "is_nil")
		}
	case field.TypeBool:
		ops = []string{"eq"}
	case field.TypeEnum:
		ops = []string{"eq", "in", "neq", "not_in"}
	case field.TypeUUID:
		ops = []string{"eq", "in", "neq", "not_in"}
		if f.Optional || f.Nillable {
			ops = append(ops, "is_nil")
		}
	default:
		// Other types not filterable by default.
		return nil
	}
	sort.Strings(ops)
	return ops
}

// isSortable reports whether a field is sortable by default.
func isSortable(f *gen.Field) bool {
	t := f.Type.Type
	switch t {
	case field.TypeString, field.TypeInt, field.TypeInt8, field.TypeInt16, field.TypeInt32, field.TypeInt64,
		field.TypeUint, field.TypeUint8, field.TypeUint16, field.TypeUint32, field.TypeUint64,
		field.TypeFloat32, field.TypeFloat64, field.TypeTime, field.TypeEnum:
		return true
	default:
		return false
	}
}

// pascal converts a snake_case or camelCase name to PascalCase.
func pascal(s string) string {
	if s == "" {
		return ""
	}
	words := strings.Split(s, "_")
	var b strings.Builder
	for _, w := range words {
		if w == "" {
			continue
		}
		r := []rune(w)
		r[0] = unicode.ToUpper(r[0])
		b.WriteString(string(r))
	}
	return b.String()
}
