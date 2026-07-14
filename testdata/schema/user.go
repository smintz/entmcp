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
//
// These fields intentionally exercise the codegen edge cases that must produce
// code compiling against real ent output:
//   - external_id: Go-initialism naming (must become ExternalID, not ExternalId).
//   - score:       int64 filter predicate (must cast to int64, not int).
//   - avatar:      []byte setter (value, no SetNillable variant).
//   - nickname:    optional/nillable scalar (must use SetNillableNickname).
func (User) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			NotEmpty().
			Comment("The user's full name"),
		field.String("email").
			Unique().
			Comment("The user's email address"),
		field.String("external_id").
			Optional().
			Comment("External identifier (initialism naming)"),
		field.Int("age").
			Optional().
			Comment("The user's age"),
		field.Int64("score").
			Optional().
			Comment("A wide integer, exercises int64 filter predicates"),
		field.Bytes("avatar").
			Optional().
			Comment("Raw avatar bytes, exercises []byte setters"),
		field.String("nickname").
			Optional().
			Nillable().
			Comment("Optional nickname, exercises SetNillable setters"),
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
