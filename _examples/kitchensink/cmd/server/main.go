// Copyright 2025 The entmcp Authors. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

// This example requires generating code first:
//
//	cd _examples/kitchensink && go run generate.go
//
// Then run:
//
//	go run _examples/kitchensink/cmd/server/main.go

package main

import (
	"context"
	"log"
	"log/slog"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/smintz/entmcp/_examples/kitchensink/ent"
	"github.com/smintz/entmcp/_examples/kitchensink/ent/entmcp"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))

	client, err := ent.Open("sqlite3", "file:kitchensink.db?cache=shared&_fk=1")
	if err != nil {
		log.Fatalf("opening ent client: %v", err)
	}
	defer client.Close()

	// Run auto-migration to create the schema.
	if err := client.Schema.Create(ctx); err != nil {
		log.Fatalf("running schema migration: %v", err)
	}

	srv := entmcp.NewMCPServer(client,
		entmcp.WithServerInfo("kitchensink", "v1"),
		entmcp.WithLogger(logger),
	)

	if err := srv.Run(ctx, mcp.NewStdioTransport()); err != nil {
		log.Fatal(err)
	}
}
