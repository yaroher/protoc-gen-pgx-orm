package orm

import "strings"

// ---------------------------------------------------------------------------
// DELETE builder -------------------------------------------------------------
// ---------------------------------------------------------------------------

type DeleteQuery[F fieldAlias] struct {
	baseQuery[F]
	whereClauses []Clause[F]
}

func (q *DeleteQuery[F]) mustOrmQuery() {}

func (q *DeleteQuery[F]) Build() (string, []any) {
	idx := 1
	sb := sbPool.Get().(*strings.Builder)
	sb.Reset()
	sb.Grow(96)

	args := make([]any, 0, len(q.whereClauses)*2)
	q.build(sb, q.tableAlias(), &idx, &args)
	sql := sb.String() // копия в новую строку
	sbPool.Put(sb)
	return sql, args
}

func (q *DeleteQuery[F]) build(sb *strings.Builder, ta string, paramIndex *int, args *[]any) {
	sb.WriteString("DELETE FROM ")
	sb.WriteString(ta)

	if len(q.whereClauses) > 0 {
		sb.WriteString(" WHERE ")
		for i, cl := range q.whereClauses {
			if i > 0 {
				sb.WriteString(" AND ")
			}
			cl.build(sb, ta, paramIndex, args)
		}
	}
	q.buildReturning(sb)
	sb.WriteByte(';')
}

func (q *DeleteQuery[F]) Where(clause ...Clause[F]) *DeleteQuery[F] {
	q.whereClauses = append(q.whereClauses, clause...)
	return q
}
func (q *DeleteQuery[F]) Returning(fields ...F) *DeleteQuery[F] {
	q.usingFields = fields
	return q
}
func (q *DeleteQuery[F]) ReturningAll() *DeleteQuery[F] {
	q.usingFields = q.allFields
	return q
}
