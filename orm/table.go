package orm

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
)

type TableI[F fieldAlias, T targeter[F]] interface {
	AllFields() []F
	AllFieldsExcept(field ...F) []F
	Name() string
	NewScanner() T
	Select(field ...F) *SelectQuery[F]
	Select1() *SelectQuery[F]
	SelectAll() *SelectQuery[F]
	Insert() *InsertQuery[F]
	Update() *UpdateQuery[F]
	Delete() *DeleteQuery[F]
	Query(ctx context.Context, db DB, query ormQuery) ([]T, error)
	QueryRow(ctx context.Context, db DB, query ormQuery) (T, error)
	Execute(ctx context.Context, db DB, query ormQuery) (int64, error)
}
type table[F fieldAlias, T targeter[F]] struct {
	alias       string
	allFields   []F // hack to set outside
	scanFactory func() T
}

func newTable[F fieldAlias, T targeter[F]](
	alias string,
	scanFactory func() T,
	fields ...F,
) *table[F, T] {
	return &table[F, T]{
		alias:       alias,
		allFields:   fields,
		scanFactory: scanFactory,
	}
}

func (t *table[F, T]) baseQuery(ta string, field ...F) baseQuery[F] {
	return baseQuery[F]{
		ta:          ta,
		usingFields: field,
		allFields:   t.allFields,
	}
}
func (t *table[F, T]) Name() string {
	return t.alias
}
func (t *table[F, T]) AllFields() []F {
	return t.allFields
}
func (t *table[F, T]) AllFieldsExcept(field ...F) []F {
	ret := make([]F, 0)
	for _, f := range t.allFields {
		needed := true
		for _, skip := range field {
			if skip.String() == f.String() {
				needed = false
			}
		}
		if needed {
			continue
		}
		ret = append(ret, f)
	}
	return ret
}
func (t *table[F, T]) NewScanner() T {
	return t.scanFactory()
}
func (t *table[F, T]) Select(field ...F) *SelectQuery[F] {
	return &SelectQuery[F]{
		baseQuery: t.baseQuery(t.alias, field...),
	}
}
func (t *table[F, T]) Select1() *SelectQuery[F] {
	return t.Select()
}
func (t *table[F, T]) SelectAll() *SelectQuery[F] {
	return t.Select(t.allFields...)
}
func (t *table[F, T]) Update() *UpdateQuery[F] {
	return &UpdateQuery[F]{
		baseQuery: t.baseQuery(t.alias),
	}
}
func (t *table[F, T]) Delete() *DeleteQuery[F] {
	return &DeleteQuery[F]{
		baseQuery: t.baseQuery(t.alias),
	}
}
func (t *table[F, T]) Insert() *InsertQuery[F] {
	return &InsertQuery[F]{
		baseQuery: t.baseQuery(t.alias),
	}
}

func (t *table[F, T]) QueryRow(ctx context.Context, db DB, query ormQuery) (trg T, err error) {
	trg = t.scanFactory()
	scanAbleFields := query.scanAbleFields()
	sql, args := query.Build()
	targets := make([]any, 0, len(scanAbleFields))
	for _, f := range scanAbleFields {
		targets = append(targets, trg.getTarget(f)())
	}
	err = db.QueryRow(ctx, sql, args...).Scan(targets...)
	if err != nil {
		return trg, err
	}
	return trg, nil
}

func (t *table[F, T]) Query(ctx context.Context, db DB, query ormQuery) (trgs []T, err error) {
	scanAbleFields := query.scanAbleFields()
	sql, args := query.Build()
	rows, err := db.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		trg := t.scanFactory()
		targets := make([]any, len(scanAbleFields))
		for i, f := range scanAbleFields {
			targets[i] = trg.getTarget(f)()
		}
		err = rows.Scan(targets...)
		if err != nil {
			return nil, err
		}
		trgs = append(trgs, trg)
	}
	return trgs, nil
}

func (t *table[F, T]) Execute(ctx context.Context, db DB, query ormQuery) (int64, error) {
	sql, args := query.Build()
	tag, err := db.Exec(ctx, sql, args...)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}
func (t *table[F, T]) Raw(sql string, args ...any) Clause[F] {
	return &RawExprClause[F]{sql, args}
}
func (t *table[F, T]) ExistsRaw(sql string, args ...any) Clause[F] {
	return &ExistsClause[F]{SubQuery: &RawExprClause[F]{sql, args}, Negate: false}
}
func (t *table[F, T]) NotExistsRaw(sql string, args ...any) Clause[F] {
	return &ExistsClause[F]{SubQuery: &RawExprClause[F]{sql, args}, Negate: true}
}
func (t *table[F, T]) ExistsOf(query ormQuery) Clause[F] {
	return &ExistsClause[F]{SubQuery: &SubQueryExprClause[F]{query}, Negate: false}
}
func (t *table[F, T]) NotExistsOf(query ormQuery) Clause[F] {
	return &ExistsClause[F]{SubQuery: &SubQueryExprClause[F]{query}, Negate: true}
}
func (t *table[F, T]) And(clauses ...Clause[F]) Clause[F] {
	return &AndClause[F]{clauses}
}
func (t *table[F, T]) Or(clauses ...Clause[F]) Clause[F] {
	return &OrClause[F]{clauses}
}

type copyIterator[T valuer] struct {
	rows                 []T
	skippedFirstNextCall bool
}

func newCopyIterator[T valuer](rows []T) *copyIterator[T] {
	return &copyIterator[T]{
		rows: rows,
	}
}
func (r *copyIterator[T]) Err() error { return nil }
func (r *copyIterator[T]) Next() bool {
	if len(r.rows) == 0 {
		return false
	}
	if !r.skippedFirstNextCall {
		r.skippedFirstNextCall = true
		return true
	}
	r.rows = r.rows[1:]
	return len(r.rows) > 0
}
func (r *copyIterator[F]) Values() ([]interface{}, error) {
	return r.rows[0].values(), nil
}

func (t *table[F, T]) CopyFrom(ctx context.Context, db DB, values []T, fields ...F) (int64, error) {
	if len(fields) == 0 {
		return 0, errors.New("pgx-orm: fields is empty")
	}
	if len(fields) > 0 {
		fields = t.allFields
	}
	fieldsStrings := make([]string, len(fields))
	for i, f := range fields {
		fieldsStrings[i] = f.String()
	}
	return db.CopyFrom(ctx, pgx.Identifier{t.alias}, fieldsStrings, newCopyIterator[T](values))
}
