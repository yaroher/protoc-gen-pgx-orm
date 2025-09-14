package orm

import (
	"reflect"
	"strconv"
	"strings"
)

type Clause[F fieldAlias] interface {
	sqlBuilder
	mustClauseAlias(F)
}

// renumberPlaceholders — renumber $placeholders when embedding sub‑queries
func renumberPlaceholders(sql string, paramIndex *int) string {
	var b strings.Builder
	for i := 0; i < len(sql); {
		if sql[i] == '$' {
			j := i + 1
			for j < len(sql) && sql[j] >= '0' && sql[j] <= '9' {
				j++
			}
			b.WriteByte('$')
			b.WriteString(strconv.Itoa(*paramIndex))
			*paramIndex++
			i = j
		} else {
			b.WriteByte(sql[i])
			i++
		}
	}
	return b.String()
}

type ParamExprClause[F fieldAlias] struct {
	Value     any
	LikeWrapp bool
}

func (e *ParamExprClause[F]) mustClauseAlias(F) {}
func (e *ParamExprClause[F]) build(buf *strings.Builder, ta string, paramIndex *int, args *[]any) {
	if e.LikeWrapp {
		buf.WriteString("'%' || ")
	}
	buf.WriteByte('$')
	buf.WriteString(strconv.Itoa(*paramIndex))
	*paramIndex++
	if e.LikeWrapp {
		buf.WriteString("::text || '%'")
	}
	*args = append(*args, e.Value)
}

type SliceExprClause[F fieldAlias] struct {
	Values any // slice
}

func (e *SliceExprClause[F]) mustClauseAlias(F) {}
func (e *SliceExprClause[F]) build(buf *strings.Builder, ta string, paramIndex *int, args *[]any) {
	rv := reflect.ValueOf(e.Values)
	if rv.Kind() != reflect.Slice {
		panic("SliceExprClause expects slice")
	}
	buf.WriteByte('(')
	buf.WriteByte('$')
	buf.WriteString(strconv.Itoa(*paramIndex))
	*paramIndex++
	*args = append(*args, e.Values)
	buf.WriteByte(')')
}

type RawExprClause[F fieldAlias] struct {
	SQL  string
	Args []any
}

func (e *RawExprClause[F]) mustClauseAlias(F) {}
func (e *RawExprClause[F]) build(buf *strings.Builder, ta string, paramIndex *int, args *[]any) {
	parts := strings.Split(e.SQL, "?")
	for i, part := range parts {
		if i > 0 {
			buf.WriteByte('$')
			buf.WriteString(strconv.Itoa(*paramIndex))
			*paramIndex++
		}
		buf.WriteString(part)
	}
	*args = append(*args, e.Args...)
}

type SubQueryExprClause[F fieldAlias] struct {
	Query ormQuery
}

func (e *SubQueryExprClause[F]) mustClauseAlias(F) {}
func (e *SubQueryExprClause[F]) build(buf *strings.Builder, ta string, paramIndex *int, args *[]any) {
	//sql, subArgs := e.Query.Build()
	//sql = renumberPlaceholders(sql, paramIndex)
	//buf.WriteByte('(')
	//buf.WriteString(sql)
	//buf.WriteByte(')')
	//*args = append(*args, subArgs...)
	buf.WriteByte('(')
	e.Query.build(buf, e.Query.tableAlias(), paramIndex, args)
	buf.WriteByte(')')

}

type FieldClause[F fieldAlias] struct {
	Field    F
	Operator string
	Right    sqlBuilder
	Negate   bool
}

func (c *FieldClause[F]) mustClauseAlias(F) {}
func (c *FieldClause[F]) build(buf *strings.Builder, ta string, paramIndex *int, args *[]any) {
	if c.Negate {
		buf.WriteString("NOT (")
	}
	buf.WriteString(ta)
	buf.WriteByte('.')
	buf.WriteString(c.Field.String())
	buf.WriteByte(' ')
	buf.WriteString(c.Operator)
	buf.WriteByte(' ')
	c.Right.build(buf, ta, paramIndex, args)
	if c.Negate {
		buf.WriteByte(')')
	}
}

type AndClause[F fieldAlias] struct {
	Clauses []Clause[F]
}

func (c *AndClause[F]) mustClauseAlias(F) {}
func (c *AndClause[F]) build(buf *strings.Builder, ta string, paramIndex *int, args *[]any) {
	buf.WriteByte('(')
	for i, cl := range c.Clauses {
		if i > 0 {
			buf.WriteString(" AND ")
		}
		cl.build(buf, ta, paramIndex, args)
	}
	buf.WriteByte(')')
}

type OrClause[F fieldAlias] struct {
	Clauses []Clause[F]
}

func (c *OrClause[F]) mustClauseAlias(F) {}
func (c *OrClause[F]) build(buf *strings.Builder, ta string, paramIndex *int, args *[]any) {
	buf.WriteByte('(')
	for i, cl := range c.Clauses {
		if i > 0 {
			buf.WriteString(" OR ")
		}
		cl.build(buf, ta, paramIndex, args)
	}
	buf.WriteByte(')')
}

type NotClause[F fieldAlias] struct {
	Inner Clause[F]
}

func (c *NotClause[F]) mustClauseAlias(F) {}
func (c *NotClause[F]) build(buf *strings.Builder, ta string, paramIndex *int, args *[]any) {
	buf.WriteString("NOT (")
	c.Inner.build(buf, ta, paramIndex, args)
	buf.WriteByte(')')
}

type ExistsClause[F fieldAlias] struct {
	SubQuery Clause[F]
	Negate   bool
}

func (c *ExistsClause[F]) mustClauseAlias(F) {}
func (c *ExistsClause[F]) build(buf *strings.Builder, ta string, paramIndex *int, args *[]any) {
	if c.Negate {
		buf.WriteString("NOT ")
	}
	buf.WriteString("EXISTS ")
	c.SubQuery.build(buf, ta, paramIndex, args)
}
