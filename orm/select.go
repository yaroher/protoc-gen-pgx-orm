package orm

import (
	"strconv"
	"strings"
)

type SelectQuery[F fieldAlias] struct {
	baseQuery[F]
	distinct     bool
	whereClauses []Clause[F]
	groupBy      []F
	orderByASC   []F
	orderByDESC  []F
	limit        int
	offset       int
	forUpdate    bool
}

func (q *SelectQuery[F]) Build() (string, []any) {
	i := 1
	// берём буфер из пула
	sb := sbPool.Get().(*strings.Builder)
	sb.Reset()

	// heuristics: 128 базовый + ~32 на каждый where/order/group + ~16 на поле
	approxCap := 128 + (len(q.whereClauses)+len(q.orderByASC)+len(q.orderByDESC)+len(q.groupBy))*32 + len(q.usingFields)*16
	sb.Grow(approxCap)

	args := make([]any, 0, len(q.whereClauses)*2) // простой грубый estimate
	q.build(sb, q.tableAlias(), &i, &args)
	sql := sb.String() // копия в новую строку
	sbPool.Put(sb)
	return sql, args
}

//goland:noinspection t
func (q *SelectQuery[F]) build(buf *strings.Builder, ta string, paramIndex *int, args *[]any) {
	inSubquery := buf.Len() > 0 || *paramIndex > 1 || len(*args) > 0
	// ---------- SELECT ----------
	buf.WriteString("SELECT ")
	if q.distinct {
		buf.WriteString("DISTINCT ")
	}
	if len(q.usingFields) > 0 {
		for i, f := range q.usingFields {
			if f.IsCount() {
				buf.WriteString("COUNT(")
				buf.WriteString(ta)
				buf.WriteByte('.')
				buf.WriteString(f.String())
				buf.WriteString(")")
				continue
			}
			if i > 0 {
				buf.WriteString(", ")
			}
			buf.WriteString(ta)
			buf.WriteByte('.')
			buf.WriteString(f.String())
		}
	} else {
		buf.WriteByte('1')
	}

	// ---------- FROM ----------
	buf.WriteString(" FROM ")
	buf.WriteString(ta)
	buf.WriteString(" AS ")
	buf.WriteString(ta)

	// ---------- WHERE ----------
	if len(q.whereClauses) > 0 {
		buf.WriteString(" WHERE ")
		for i, clause := range q.whereClauses {
			if i > 0 {
				buf.WriteString(" AND ")
			}
			clause.build(buf, ta, paramIndex, args)
		}
	}

	// ---------- GROUP BY ----------
	if len(q.groupBy) > 0 {
		buf.WriteString(" GROUP BY ")
		for i, g := range q.groupBy {
			if i > 0 {
				buf.WriteString(", ")
			}
			buf.WriteString(ta)
			buf.WriteByte('.')
			buf.WriteString(g.String())
		}
	}

	// ---------- ORDER BY ----------
	if len(q.orderByASC) > 0 || len(q.orderByDESC) > 0 {
		buf.WriteString(" ORDER BY ")
		first := true
		for _, f := range q.orderByASC {
			if !first {
				buf.WriteString(", ")
			}
			buf.WriteString(ta)
			buf.WriteByte('.')
			buf.WriteString(f.String())
			buf.WriteString(" ASC")
			first = false
		}
		for _, f := range q.orderByDESC {
			if !first {
				buf.WriteString(", ")
			}
			buf.WriteString(ta)
			buf.WriteByte('.')
			buf.WriteString(f.String())
			buf.WriteString(" DESC")
			first = false
		}
	}

	// ---------- LIMIT / OFFSET ----------
	if q.limit > 0 {
		buf.WriteString(" LIMIT ")
		buf.WriteString(strconv.Itoa(q.limit))
	}
	if q.offset > 0 {
		buf.WriteString(" OFFSET ")
		buf.WriteString(strconv.Itoa(q.offset))
	}

	// ---------- FOR UPDATE ----------
	if q.forUpdate {
		buf.WriteString(" FOR UPDATE")
	}
	if !inSubquery {
		buf.WriteByte(';')
	}
}
func (q *SelectQuery[F]) Fields(
	fields ...F,
) *SelectQuery[F] {
	q.usingFields = fields
	return q
}
func (q *SelectQuery[F]) Distinct() *SelectQuery[F] {
	q.distinct = true
	return q
}
func (q *SelectQuery[F]) Alias(
	alias string,
) *SelectQuery[F] {
	q.ta = alias
	return q
}
func (q *SelectQuery[F]) Where(clause ...Clause[F]) *SelectQuery[F] {
	q.whereClauses = append(q.whereClauses, clause...)
	return q
}
func (q *SelectQuery[F]) GroupBy(fields ...F) *SelectQuery[F] {
	q.groupBy = append(q.groupBy, fields...)
	return q
}
func (q *SelectQuery[F]) OrderByASC(fields ...F) *SelectQuery[F] {
	q.orderByASC = append(q.orderByASC, fields...)
	return q
}
func (q *SelectQuery[F]) OrderByDESC(fields ...F) *SelectQuery[F] {
	q.orderByDESC = append(q.orderByDESC, fields...)
	return q
}
func (q *SelectQuery[F]) Limit(limit int) *SelectQuery[F] {
	q.limit = limit
	return q
}
func (q *SelectQuery[F]) Offset(offset int) *SelectQuery[F] {
	q.offset = offset
	return q
}
func (q *SelectQuery[F]) ForUpdate() *SelectQuery[F] {
	q.forUpdate = true
	return q
}
func (q *SelectQuery[F]) SetLimit(limit int) {
	q.limit = limit
}
func (q *SelectQuery[F]) SetOffset(offset int) {
	q.offset = offset
}
func (q *SelectQuery[F]) SetOrderBy(asc bool, fields ...F) {
	if asc {
		q.orderByASC = append(q.orderByASC, fields...)
	} else {
		q.orderByDESC = append(q.orderByDESC, fields...)
	}
}
