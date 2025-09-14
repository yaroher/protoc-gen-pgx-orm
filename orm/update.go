package orm

import "strings"

// ---------------------------------------------------------------------------
// UPDATE builder -------------------------------------------------------------
// ---------------------------------------------------------------------------

type UpdateQuery[F fieldAlias] struct {
	baseQuery[F]
	setAssigns   []ValueSetter[F]
	whereClauses []Clause[F]
}

func (q *UpdateQuery[F]) mustOrmQuery() {}

func (q *UpdateQuery[F]) Build() (string, []any) {
	i := 1
	sb := sbPool.Get().(*strings.Builder)
	sb.Reset()
	sb.Grow(128 + len(q.setAssigns)*24)

	args := make([]any, 0, len(q.setAssigns)+len(q.whereClauses)*2)
	q.build(sb, q.tableAlias(), &i, &args)
	sql := sb.String() // копия в новую строку
	sbPool.Put(sb)
	return sql, args
}

func (q *UpdateQuery[F]) build(buf *strings.Builder, ta string, paramIndex *int, args *[]any) {
	buf.WriteString("UPDATE ")
	buf.WriteString(ta)
	buf.WriteString(" SET ")
	for i, asg := range q.setAssigns {
		if i > 0 {
			buf.WriteString(", ")
		}
		asg.build(buf, ta, paramIndex, args)
	}

	if len(q.whereClauses) > 0 {
		buf.WriteString(" WHERE ")
		for i, cl := range q.whereClauses {
			if i > 0 {
				buf.WriteString(" AND ")
			}
			cl.build(buf, ta, paramIndex, args)
		}
	}

	q.buildReturning(buf)

	buf.WriteByte(';')
}

func (q *UpdateQuery[F]) Where(clause ...Clause[F]) *UpdateQuery[F] {
	q.whereClauses = append(q.whereClauses, clause...)
	return q
}
func (q *UpdateQuery[F]) Set(assign ...ValueSetter[F]) *UpdateQuery[F] {
	q.setAssigns = append(q.setAssigns, assign...)
	return q
}
func (q *UpdateQuery[F]) ReturningAll() *UpdateQuery[F] {
	q.usingFields = q.allFields
	return q
}
func (q *UpdateQuery[F]) Returning(fields ...F) *UpdateQuery[F] {
	q.usingFields = fields
	return q
}
