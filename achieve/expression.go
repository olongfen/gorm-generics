package achieve

import "gorm.io/gorm/clause"

// ILike ilike
type ILike clause.Eq

// Build builder sql
func (like ILike) Build(builder clause.Builder) {
	builder.WriteQuoted(like.Column)
	_, err := builder.WriteString(" ILIKE ")
	if err != nil {
		panic(err)
	}
	builder.AddVar(builder, like.Value)
}

// NegationBuild builder sql
func (like ILike) NegationBuild(builder clause.Builder) {
	builder.WriteQuoted(like.Column)
	_, err := builder.WriteString(" NOT ILIKE ")
	if err != nil {
		panic(err)
	}
	builder.AddVar(builder, like.Value)
}
