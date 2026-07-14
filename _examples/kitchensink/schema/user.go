// Copyright 2025 The entmcp Authors. All rights reserved.
// Use of this source code is governed by an MIT-style
// license that can be found in the LICENSE file.

package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

// User holds the schema definition for the User entity.
type User struct {
	ent.Schema
}

// Fields of the User.
func (User) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			NotEmpty().
			Comment("The user's full name"),
		field.String("email").
			Unique().
			Comment("The user's email address"),
		field.Int("age").
			Optional().
			Comment("The user's age"),
		field.String("password").
			Sensitive().
			Optional(),
		field.Time("created_at").
			Default(time.Now).
			Immutable().
			Comment("When the user was created"),
		field.Bool("active").
			Default(true).
			Comment("Whether the user is active"),
	}
}

// Edges of the User.
func (User) Edges() []ent.Edge {
	return nil
}
