//go:build ignore

// This program generates the ent code for the kitchensink example.
// Run: go run generate.go

package main

import (
	"log"

	"entgo.io/ent/entc"
	"entgo.io/ent/entc/gen"
	"github.com/smintz/entmcp"
)

func main() {
	ext, err := entmcp.NewExtension(
		entmcp.WithToolPrefix(""),
		entmcp.WithDefaultOperations(entmcp.OpAll),
		entmcp.WithMaxPageSize(100),
		entmcp.WithDefaultPageSize(25),
	)
	if err != nil {
		log.Fatal(err)
	}
	if err := entc.Generate("./schema", &gen.Config{
		Target:  "./ent",
		Package: "github.com/smintz/entmcp/_examples/kitchensink/ent",
	}, entc.Extensions(ext)); err != nil {
		log.Fatalf("running ent codegen: %v", err)
	}
}
