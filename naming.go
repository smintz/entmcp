// Copyright 2025 The entmcp Authors. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package entmcp

import (
	"strings"
	"unicode"
)

// toSnakeCase converts a PascalCase or camelCase name to snake_case.
// Examples: "UserProfile" -> "user_profile", "HTTPServer" -> "http_server".
func toSnakeCase(s string) string {
	if s == "" {
		return ""
	}
	var b strings.Builder
	runes := []rune(s)
	for i, r := range runes {
		if unicode.IsUpper(r) {
			if i > 0 {
				// Don't add underscore if previous char was also uppercase
				// and next char (if any) is also uppercase or end of string
				// (e.g. "HTTP" should stay together until next lowercase).
				prev := runes[i-1]
				if unicode.IsLower(prev) || unicode.IsDigit(prev) {
					b.WriteRune('_')
				} else if i+1 < len(runes) && unicode.IsLower(runes[i+1]) {
					b.WriteRune('_')
				}
			}
			b.WriteRune(unicode.ToLower(r))
		} else {
			b.WriteRune(r)
		}
	}
	return strings.TrimPrefix(b.String(), "_")
}

// plural returns a simple plural form of the given name.
// It applies basic English pluralization rules.
func plural(s string) string {
	if s == "" {
		return ""
	}
	lower := strings.ToLower(s)
	// Irregular plurals
	irregulars := map[string]string{
		"person": "people",
		"child":  "children",
		"mouse":  "mice",
		"goose":  "geese",
		"tooth":  "teeth",
		"foot":   "feet",
		"ox":     "oxen",
		"man":    "men",
		"woman":  "women",
	}
	if p, ok := irregulars[lower]; ok {
		return p
	}
	// Standard rules
	switch {
	case strings.HasSuffix(lower, "s") || strings.HasSuffix(lower, "x") ||
		strings.HasSuffix(lower, "z") || strings.HasSuffix(lower, "ch") ||
		strings.HasSuffix(lower, "sh"):
		return s + "es"
	case strings.HasSuffix(lower, "y") && len(s) > 1 && !isVowel(rune(lower[len(lower)-2])):
		return s[:len(s)-1] + "ies"
	default:
		return s + "s"
	}
}

func isVowel(r rune) bool {
	return strings.ContainsRune("aeiou", r)
}

// toolName returns the base tool name for an entity type.
// If the annotation has a ToolName set, that is used; otherwise snake_case of typeName.
func toolName(typeName string, ant Annotation) string {
	if ant.ToolName != "" {
		return ant.ToolName
	}
	return toSnakeCase(typeName)
}

// prefixed prepends the tool prefix (if any) to a tool name.
func prefixed(prefix, name string) string {
	if prefix == "" {
		return name
	}
	return prefix + "_" + name
}

// createToolName returns the "create_<entity>" tool name.
func createToolName(prefix, entityBase string) string {
	return prefixed(prefix, "create_"+entityBase)
}

// getToolName returns the "get_<entity>" tool name.
func getToolName(prefix, entityBase string) string {
	return prefixed(prefix, "get_"+entityBase)
}

// listToolName returns the "list_<entities>" tool name.
func listToolName(prefix, entityBase string) string {
	return prefixed(prefix, "list_"+plural(entityBase))
}

// updateToolName returns the "update_<entity>" tool name.
func updateToolName(prefix, entityBase string) string {
	return prefixed(prefix, "update_"+entityBase)
}

// deleteToolName returns the "delete_<entity>" tool name.
func deleteToolName(prefix, entityBase string) string {
	return prefixed(prefix, "delete_"+entityBase)
}

// edgeListToolName returns the "list_<entity>_<edge>" tool name.
func edgeListToolName(prefix, entityBase, edgeName string) string {
	return prefixed(prefix, "list_"+entityBase+"_"+toSnakeCase(edgeName))
}

// ToSnakeCaseExport is an exported version of toSnakeCase for testing.
func ToSnakeCaseExport(s string) string { return toSnakeCase(s) }
