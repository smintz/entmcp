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

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))

	// NOTE: After running generate.go, import paths will be:
	//   "<module>/ent"
	//   "<module>/ent/entmcp"
	//
	// For this example we demonstrate the server wiring pattern.
	// The actual generated code would be imported like:
	//
	//   client, err := ent.Open("sqlite3", "file:ent?mode=memory&cache=shared&_fk=1")
	//   ...
	//   srv := entmcp.NewMCPServer(client,
	//       entmcp.WithServerInfo("kitchensink", "v1"),
	//       entmcp.WithLogger(logger),
	//   )
	//   if err := srv.Run(ctx, mcp.NewStdioTransport()); err != nil {
	//       log.Fatal(err)
	//   }

	_ = ctx
	_ = logger
	_ = mcp.NewServer

	log.Println("Run 'go run generate.go' in the kitchensink directory first to generate ent+entmcp code.")
	log.Println("Then update the import paths in this file to match the generated module.")
	os.Exit(0)
}
