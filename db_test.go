package gormgenerics

import (
	"context"
	"fmt"
	"gorm.io/gorm/clause"
	"testing"
)

func TestNewDataBase(t *testing.T) {
	base, err := NewDataBase("postgresql", "postgres://postgres:starwiz.cn@192.168.3.247:31301/pacific?sslmode=disable",
		WithTablePrefix("pacific_"),
	)
	if err != nil {
		return
	}
	var data []map[string]any
	base.DB(context.Background()).Clauses(clause.Join{
		Table: clause.Table{Name: "pacific_casbin_rules"},
		ON: clause.Where{
			Exprs: []clause.Expression{clause.Eq{Column: clause.Column{Table: "pacific_projects", Name: "uuid"},
				Value: clause.Column{Table: "pacific_casbin_rules", Name: "v2"}}},
		},
	}).Find(&data)
	fmt.Println(data)
}
