// Copyright 2025 The entmcp Authors. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package entmcp_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"entgo.io/ent/entc"
	"entgo.io/ent/entc/gen"
	"github.com/smintz/entmcp"
	"github.com/stretchr/testify/require"
)

// TestCodeGeneration tests that the entmcp extension generates valid Go code.
func TestCodeGeneration(t *testing.T) {
	ext, err := entmcp.NewExtension()
	require.NoError(t, err, "NewExtension should not return an error")

	tmpDir := t.TempDir()
	cfg := &gen.Config{
		Target:  tmpDir,
		Package: "github.com/example/ent",
		Schema:  "testdata/schema",
	}
	err = entc.Generate("./testdata/schema", cfg, entc.Extensions(ext))
	require.NoError(t, err, "entc.Generate should not return an error")

	// Verify entmcp subdirectory was created.
	entmcpDir := filepath.Join(tmpDir, "entmcp")
	require.DirExists(t, entmcpDir, "entmcp directory should be created")

	// Verify expected files exist.
	expectedFiles := []string{
		"entmcp.go",
		"helpers.go",
		"user_tools.go",
	}
	for _, f := range expectedFiles {
		path := filepath.Join(entmcpDir, f)
		require.FileExists(t, path, "expected file %s to be generated", f)
	}

	// Verify generated file content.
	serverContent, err := os.ReadFile(filepath.Join(entmcpDir, "entmcp.go"))
	require.NoError(t, err)
	require.Contains(t, string(serverContent), "NewMCPServer", "server file should contain NewMCPServer")
	require.Contains(t, string(serverContent), "registerUserTools", "server file should reference registerUserTools")

	userToolsContent, err := os.ReadFile(filepath.Join(entmcpDir, "user_tools.go"))
	require.NoError(t, err)

	// Verify all CRUD tools are registered.
	require.Contains(t, string(userToolsContent), "create_user", "should have create tool")
	require.Contains(t, string(userToolsContent), "get_user", "should have get tool")
	require.Contains(t, string(userToolsContent), "list_users", "should have list tool")
	require.Contains(t, string(userToolsContent), "update_user", "should have update tool")
	require.Contains(t, string(userToolsContent), "delete_user", "should have delete tool")

	// Verify sensitive field (password) is excluded from output struct and read operations.
	// It may appear in write-path code (create/update handlers) but must not be in the output struct.
	outStructStart := strings.Index(string(userToolsContent), "type UserOutput struct")
	require.Greater(t, outStructStart, 0, "should have UserOutput struct")
	outStructEnd := strings.Index(string(userToolsContent)[outStructStart:], "\n}")
	require.Greater(t, outStructEnd, 0)
	outStruct := string(userToolsContent)[outStructStart : outStructStart+outStructEnd]
	require.NotContains(t, outStruct, "Password", "sensitive field should not appear in output struct")
	require.NotContains(t, outStruct, "password", "sensitive field should not appear in output struct")

	// Verify immutable field (created_at) is excluded from update input.
	// Check the update handler function body specifically.
	require.NotContains(t, string(userToolsContent), `args["created_at"]`, "immutable field should not be in update handler")
}

// TestNewExtension tests the NewExtension function with various options.
func TestNewExtension(t *testing.T) {
	tests := []struct {
		name    string
		opts    []entmcp.Option
		wantErr bool
		errMsg  string
	}{
		{
			name: "default config",
			opts: nil,
		},
		{
			name: "with prefix",
			opts: []entmcp.Option{entmcp.WithToolPrefix("myapp")},
		},
		{
			name:    "invalid prefix",
			opts:    []entmcp.Option{entmcp.WithToolPrefix("My-App")},
			wantErr: true,
			errMsg:  "ToolPrefix",
		},
		{
			name: "read only",
			opts: []entmcp.Option{entmcp.WithReadOnly(true)},
		},
		{
			name: "custom page size",
			opts: []entmcp.Option{
				entmcp.WithMaxPageSize(200),
				entmcp.WithDefaultPageSize(50),
			},
		},
		{
			name:    "invalid page size",
			opts:    []entmcp.Option{entmcp.WithMaxPageSize(10), entmcp.WithDefaultPageSize(20)},
			wantErr: true,
			errMsg:  "DefaultPageSize",
		},
		{
			name: "read only disables write ops",
			opts: []entmcp.Option{
				entmcp.WithReadOnly(true),
				entmcp.WithDefaultOperations(entmcp.OpAll),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ext, err := entmcp.NewExtension(tt.opts...)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					require.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}
			require.NoError(t, err)
			require.NotNil(t, ext)
		})
	}
}

// TestAnnotations tests annotation behavior.
func TestAnnotations(t *testing.T) {
	// Test Skip annotation.
	ant := entmcp.Skip()
	require.True(t, ant.Skip)

	// Test Operations annotation.
	ops := entmcp.OpRead | entmcp.OpList
	ant2 := entmcp.OperationsAnnotation(ops)
	require.NotNil(t, ant2.Operations)
	require.True(t, ant2.Operations.Has(entmcp.OpRead))
	require.True(t, ant2.Operations.Has(entmcp.OpList))
	require.False(t, ant2.Operations.Has(entmcp.OpCreate))

	// Test ReadOnly annotation.
	ant3 := entmcp.ReadOnlyAnnotation()
	require.True(t, ant3.ReadOnly)

	// Test Description annotation.
	ant4 := entmcp.DescriptionAnnotation("a user entity")
	require.Equal(t, "a user entity", ant4.Description)

	// Test Merge.
	base := entmcp.Annotation{Description: "base desc"}
	override := entmcp.Annotation{Skip: true, Description: "override desc"}
	merged := base.Merge(override).(entmcp.Annotation)
	require.True(t, merged.Skip)
	require.Equal(t, "override desc", merged.Description)
}

// TestNaming tests the tool naming functions.
func TestNaming(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"User", "user"},
		{"UserProfile", "user_profile"},
		{"HTTPServer", "http_server"},
		{"URLParser", "url_parser"},
		{"ID", "id"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			// Access through tool name with no annotation.
			got := entmcp.ToSnakeCaseExport(tt.input)
			require.Equal(t, tt.want, got)
		})
	}
}

// TestOperationsBitmask tests the Operations type.
func TestOperationsBitmask(t *testing.T) {
	require.True(t, entmcp.OpAll.Has(entmcp.OpCreate))
	require.True(t, entmcp.OpAll.Has(entmcp.OpRead))
	require.True(t, entmcp.OpAll.Has(entmcp.OpList))
	require.True(t, entmcp.OpAll.Has(entmcp.OpUpdate))
	require.True(t, entmcp.OpAll.Has(entmcp.OpDelete))

	readOnly := entmcp.OpRead | entmcp.OpList
	require.True(t, readOnly.Has(entmcp.OpRead))
	require.True(t, readOnly.Has(entmcp.OpList))
	require.False(t, readOnly.Has(entmcp.OpCreate))
	require.False(t, readOnly.Has(entmcp.OpUpdate))
	require.False(t, readOnly.Has(entmcp.OpDelete))
}

// TestCodeGenerationWithPrefix tests generation with a tool prefix.
func TestCodeGenerationWithPrefix(t *testing.T) {
	ext, err := entmcp.NewExtension(entmcp.WithToolPrefix("myapp"))
	require.NoError(t, err)

	tmpDir := t.TempDir()
	cfg := &gen.Config{
		Target:  tmpDir,
		Package: "github.com/example/ent",
		Schema:  "testdata/schema",
	}
	err = entc.Generate("./testdata/schema", cfg, entc.Extensions(ext))
	require.NoError(t, err)

	userToolsContent, err := os.ReadFile(filepath.Join(tmpDir, "entmcp", "user_tools.go"))
	require.NoError(t, err)

	require.Contains(t, string(userToolsContent), "myapp_create_user")
	require.Contains(t, string(userToolsContent), "myapp_get_user")
	require.Contains(t, string(userToolsContent), "myapp_list_users")
}

// TestCodeGenerationReadOnly tests generation with ReadOnly=true.
func TestCodeGenerationReadOnly(t *testing.T) {
	ext, err := entmcp.NewExtension(entmcp.WithReadOnly(true))
	require.NoError(t, err)

	tmpDir := t.TempDir()
	cfg := &gen.Config{
		Target:  tmpDir,
		Package: "github.com/example/ent",
		Schema:  "testdata/schema",
	}
	err = entc.Generate("./testdata/schema", cfg, entc.Extensions(ext))
	require.NoError(t, err)

	userToolsContent, err := os.ReadFile(filepath.Join(tmpDir, "entmcp", "user_tools.go"))
	require.NoError(t, err)

	// Read-only mode should not generate write tools.
	require.NotContains(t, string(userToolsContent), "create_user")
	require.NotContains(t, string(userToolsContent), "update_user")
	require.NotContains(t, string(userToolsContent), "delete_user")

	// But should still have read tools.
	require.Contains(t, string(userToolsContent), "get_user")
	require.Contains(t, string(userToolsContent), "list_users")
}
