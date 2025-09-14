package orm

import (
	"strings"
	"testing"
)

type testTable struct{ name string }

func (ta testTable) mustOrmTableAlias() {}

func (ta testTable) String() string { return ta.name }

type testField string

func (f testField) mustFieldAlias() {
}
func (f testField) IsCount() bool  { return false }
func (f testField) String() string { return string(f) }

func buildSQL(cl Clause[testField]) (string, []any) {
	var buf strings.Builder
	paramIndex := 1
	var args []any
	cl.build(&buf, "users", &paramIndex, &args)
	return buf.String(), args
}

func TestAllClauses(t *testing.T) {
	tests := []struct {
		name            string
		cl              Clause[testField]
		wantSQLContains string
	}{
		{"Simple Eq", &FieldClause[testField]{Field: "alias", Operator: "=", Right: &ParamExprClause[testField]{Value: "John"}}, "users.alias = $1"},
		{"IN Slice", &FieldClause[testField]{Field: "id", Operator: "IN", Right: &SliceExprClause[testField]{[]int{1, 2, 3}}}, "users.id IN ($1, $2, $3)"},
		{"LIKE", &FieldClause[testField]{Field: "email", Operator: "LIKE", Right: &ParamExprClause[testField]{Value: "%@gmail.com"}}, "users.email LIKE $1"},
		{"NOT", &FieldClause[testField]{Field: "email", Operator: "LIKE", Right: &ParamExprClause[testField]{Value: "%@test.com"}, Negate: true}, "NOT (users.email LIKE $1)"},
		//{"EXISTS SubQuery", ExistsClause[testTable,testField]{SubQuery: SubQueryExprClause[testTable,testField]{
		//	Query: mockSubQuery{"SELECT id FROM roles", nil}}}, "EXISTS (SELECT id FROM roles)"},
		//{"= ANY SubQuery", FieldClause[testTable,testField]{table fieldAlias: "role_id", Operator: "= ANY", Right: SubQueryExprClause[testTable,testField]{Query: mockSubQuery{"SELECT id FROM roles", nil}}}, "users.role_id = ANY (SELECT id FROM roles)"},
		{"EXISTS Raw", &ExistsClause[testField]{SubQuery: &RawExprClause[testField]{SQL: "SELECT 1 FROM dual", Args: nil}}, "EXISTS SELECT 1 FROM dual"},
	}

	for _, tt := range tests {
		sql, _ := buildSQL(tt.cl)
		if !strings.Contains(sql, tt.wantSQLContains) {
			t.Errorf("%s SQL mismatch:\nwant contain: %s\ngot : %s", tt.name, tt.wantSQLContains, sql)
		}
	}
}

func TestLogicalNested(t *testing.T) {
	cl := &AndClause[testField]{
		Clauses: []Clause[testField]{
			&FieldClause[testField]{Field: "age", Operator: ">", Right: &ParamExprClause[testField]{Value: 18}},
			&OrClause[testField]{
				Clauses: []Clause[testField]{
					&FieldClause[testField]{Field: "status", Operator: "=", Right: &ParamExprClause[testField]{Value: "active"}},
					&FieldClause[testField]{Field: "status", Operator: "=", Right: &ParamExprClause[testField]{Value: "pending"}},
				},
			},
		},
	}
	sql, args := buildSQL(cl)
	if !strings.Contains(sql, "AND") || !strings.Contains(sql, "OR") {
		t.Fatalf("logical fail: %s", sql)
	}
	if len(args) != 3 || args[0] != 18 {
		t.Fatalf("args mismatch")
	}
}

func TestRawExpr(t *testing.T) {
	raw := &FieldClause[testField]{
		Field:    "username",
		Operator: "=",
		Right:    &RawExprClause[testField]{SQL: "LOWER(?)", Args: []any{"John"}},
	}
	sql, args := buildSQL(raw)
	if !strings.Contains(sql, "LOWER($1)") {
		t.Fatalf("raw fail: %s", sql)
	}
	if args[0] != "John" {
		t.Fatalf("args mismatch")
	}
}
