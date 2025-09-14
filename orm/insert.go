package orm

import (
	"strconv"
	"strings"
)

// ---------------------------------------------------------------------------
// INSERT builder -------------------------------------------------------------
// ---------------------------------------------------------------------------

type InsertQuery[F fieldAlias] struct {
	baseQuery[F]

	columns []F
	values  []any // single‑row insert for simplicity; extendable to [][]any

	// ON CONFLICT handling
	conflictTarget []F // columns in UNIQUE/PK to target; empty → global
	doNothing      bool
	updateAssigns  []F // SET … = … when DO UPDATE used
}

func (q *InsertQuery[F]) mustOrmQuery() {}

func (q *InsertQuery[F]) Build() (string, []any) {
	idx := 1
	sb := sbPool.Get().(*strings.Builder)
	sb.Reset()

	sb.Grow(128 + len(q.columns)*16)

	args := make([]any, 0, len(q.values)+len(q.updateAssigns))
	q.build(sb, q.tableAlias(), &idx, &args)
	sql := sb.String() // копия в новую строку
	sbPool.Put(sb)
	return sql, args
}

//goland:noinspection t
func (q *InsertQuery[F]) build(buf *strings.Builder, ta string, paramIndex *int, args *[]any) {
	// INSERT INTO tbl (c1,c2) VALUES ($1,$2)
	buf.WriteString("INSERT INTO ")
	buf.WriteString(ta)
	if len(q.columns) > 0 {
		buf.WriteString(" (")
		for i, c := range q.columns {
			if i > 0 {
				buf.WriteString(", ")
			}
			buf.WriteString(c.String())
		}
		buf.WriteByte(')')
	}

	// VALUES
	buf.WriteString(" VALUES (")
	for i, v := range q.values {
		if i > 0 {
			buf.WriteString(", ")
		}
		buf.WriteByte('$')
		buf.WriteString(strconv.Itoa(*paramIndex))
		*paramIndex++
		vals := *args
		vals = append(vals, v)
		*args = vals
	}
	buf.WriteByte(')')

	// ON CONFLICT
	if q.doNothing || len(q.updateAssigns) > 0 {
		buf.WriteString(" ON CONFLICT")
		if len(q.conflictTarget) > 0 {
			buf.WriteString(" (")
			for i, c := range q.conflictTarget {
				if i > 0 {
					buf.WriteString(", ")
				}
				buf.WriteString(c.String())
			}
			buf.WriteByte(')')
		}
		if q.doNothing {
			buf.WriteString(" DO NOTHING")
		} else {
			buf.WriteString(" DO UPDATE SET ")
			for i, asg := range q.updateAssigns {
				if i > 0 {
					buf.WriteString(", ")
				}
				buf.WriteString(asg.String() + "=EXCLUDED." + asg.String())
			}
		}
	}

	// RETURNING
	q.buildReturning(buf)
	buf.WriteByte(';')
}

func (q *InsertQuery[F]) Columns(columns ...F) *InsertQuery[F] {
	q.columns = columns
	return q
}
func (q *InsertQuery[F]) Values(values ...any) *InsertQuery[F] {
	q.values = values
	return q
}
func (q *InsertQuery[F]) From(setters ...ValueSetter[F]) *InsertQuery[F] {
	cols := make([]F, 0, len(setters))
	vals := make([]any, 0, len(setters))
	for _, setter := range setters {
		cols = append(cols, setter.Column())
		vals = append(vals, setter.Value())
	}
	q.Columns(cols...)
	q.Values(vals...)
	return q
}
func (q *InsertQuery[F]) OnConflict(columns ...F) *InsertQuery[F] {
	q.conflictTarget = columns
	return q
}
func (q *InsertQuery[F]) DoNothing() *InsertQuery[F] {
	q.doNothing = true
	return q
}
func (q *InsertQuery[F]) DoUpdate(assign ...F) *InsertQuery[F] {
	q.updateAssigns = assign
	return q
}
func (q *InsertQuery[F]) Returning(fields ...F) *InsertQuery[F] {
	q.usingFields = fields
	return q
}
func (q *InsertQuery[F]) ReturningAll() *InsertQuery[F] {
	q.usingFields = q.allFields
	return q
}
