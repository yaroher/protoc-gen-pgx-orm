package orm

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"strconv"
	"strings"
	"sync"
)

// ---------------------------------------------------------------------------
//      ░█▀▀░█░█░█▀█░▀█▀░█▀█░█░░░█░█░█▀█░█▀▄
//      ░█░█░█░█░█▀▀░░█░░█▀█░█░░░█░█░█▀█░█▀▄
//      ░▀▀▀░▀▀▀░▀░░░░▀░░▀░▀░▀▀▀░▀▀▀░▀░▀░▀░▀

type DB interface {
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
	CopyFrom(context.Context, pgx.Identifier, []string, pgx.CopyFromSource) (int64, error)
}

// ---------------------------------------------------------------------------
// Pool of strings.table for reduce allocations-----------------------------
// ---------------------------------------------------------------------------
var sbPool = sync.Pool{
	New: func() any { return &strings.Builder{} },
}

type fieldAlias interface {
	fmt.Stringer
	IsCount() bool
}
type fieldAliasImpl string

func (f fieldAliasImpl) IsCount() bool  { return false }
func (f fieldAliasImpl) String() string { return string(f) }

type sqlBuilder interface {
	build(buf *strings.Builder, ta string, paramIndex *int, args *[]any)
}

type valuer interface {
	values() []any
}

type targeter[F fieldAlias] interface {
	valuer
	getTarget(string) func() any
	getSetter(F) func() ValueSetter[F]
	getValue(F) func() any
}

type ormQuery interface {
	sqlBuilder
	mustOrmQuery()
	tableAlias() string
	scanAbleFields() []string
	Build() (string, []any)
}

type baseQuery[F fieldAlias] struct {
	ta          string
	usingFields []F
	allFields   []F
}

func (q *baseQuery[F]) tableAlias() string { return q.ta }
func (q *baseQuery[F]) mustOrmQuery()      {}
func (q *baseQuery[F]) scanAbleFields() []string {
	mp := make([]string, len(q.usingFields))
	for i, f := range q.usingFields {
		mp[i] = f.String()
	}
	return mp
}
func (q *baseQuery[F]) buildReturning(sb *strings.Builder) {
	if len(q.usingFields) == len(q.allFields) {
		sb.WriteString(" RETURNING ")
		for i, field := range q.usingFields {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(field.String())
		}
		return
	}
	if len(q.usingFields) > 0 {
		sb.WriteString(" RETURNING ")
		for i, f := range q.usingFields {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(f.String())
		}
	}
}

type ValueSetter[F fieldAlias] interface {
	sqlBuilder
	Value() any
	Column() F
}
type valueSetterImpl[F fieldAlias] struct {
	field F
	value any
	expr  string
	raw   *RawExprClause[F]
}

func NewValueSetter[F fieldAlias](field F, value any) ValueSetter[F] {
	return &valueSetterImpl[F]{field: field, value: value}
}

func (s valueSetterImpl[F]) Column() F {
	return s.field
}

func (s valueSetterImpl[F]) Value() any {
	if s.expr != "" {
		return nil
	}
	return s.value
}

func (s valueSetterImpl[F]) build(buf *strings.Builder, ta string, paramIndex *int, args *[]any) {
	buf.WriteString(s.field.String())
	buf.WriteString(" = ")
	if s.expr != "" {
		buf.WriteString(s.expr)
		return
	}
	if s.raw != nil {
		s.raw.build(buf, ta, paramIndex, args)
	} else {
		buf.WriteByte('$')
		buf.WriteString(strconv.Itoa(*paramIndex))
		*paramIndex++
		*args = append(*args, s.value)
	}
}

var (
	ErrEmptyFields = errors.New("empty fields")
	ErrEmptyModel  = errors.New("empty model")
	ErrEmptyQuery  = errors.New("empty query")
)

type TypeCaster[A, B any] func(A) B
type SqlOpType string

const (
	SqlMutation SqlOpType = "mutation"
	SqlQuery    SqlOpType = "query"
)

type DbGetter func(ctx context.Context, operation SqlOpType) DB
