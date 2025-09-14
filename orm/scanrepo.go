package orm

import (
	"context"
	"errors"
)

type scannerCallOptions[F fieldAlias, S targeter[F]] struct {
	excludeFields  []F
	conflictFields []F
	copyFields     []F
	doNothing      bool
}

type ScannerCallOptions[F fieldAlias, S targeter[F]] interface {
	apply(*scannerCallOptions[F, S])
}

type ScannerCallOptionsFn[F fieldAlias, S targeter[F]] func(*scannerCallOptions[F, S])

func (f ScannerCallOptionsFn[F, S]) apply(opts *scannerCallOptions[F, S]) {
	f(opts)
}

func WithScannerExcludeFields[F fieldAlias, S targeter[F]](fields ...F) ScannerCallOptionsFn[F, S] {
	return func(opts *scannerCallOptions[F, S]) {
		opts.excludeFields = fields
	}
}

func WithScannerCopyFields[F fieldAlias, S targeter[F]](fields ...F) ScannerCallOptionsFn[F, S] {
	return func(opts *scannerCallOptions[F, S]) {
		opts.copyFields = fields
	}
}

func WithScannerConflictFields[F fieldAlias, S targeter[F]](fields ...F) ScannerCallOptionsFn[F, S] {
	return func(opts *scannerCallOptions[F, S]) {
		opts.conflictFields = fields
	}
}

type ScannerRepository[F fieldAlias, S targeter[F]] interface {
	Table() TableI[F, S]
	Insert(ctx context.Context, entity S, opts ...ScannerCallOptions[F, S]) error
	InsertRet(ctx context.Context, entity S, opts ...ScannerCallOptions[F, S]) (S, error)
	InsertMany(ctx context.Context, entities []S, opts ...ScannerCallOptions[F, S]) error
	Update(ctx context.Context, entity S, opts ...ScannerCallOptions[F, S]) error
	UpdateRet(ctx context.Context, entity S, opts ...ScannerCallOptions[F, S]) (S, error)
	Upsert(ctx context.Context, entity S, conflictFields []F, opts ...ScannerCallOptions[F, S]) error
	UpsertRet(ctx context.Context, entity S, conflictFields []F, opts ...ScannerCallOptions[F, S]) (S, error)
	UpsertIgnore(ctx context.Context, entity S, opts ...ScannerCallOptions[F, S]) error

	GetBy(ctx context.Context, query ormQuery, opts ...ScannerCallOptions[F, S]) (S, error)
	ListBy(ctx context.Context, query ormQuery, opts ...ScannerCallOptions[F, S]) ([]S, error)
	Exec(ctx context.Context, query ormQuery, opts ...ScannerCallOptions[F, S]) error
	ExecAffected(ctx context.Context, query ormQuery, opts ...ScannerCallOptions[F, S]) (int64, error)
}

type genericScannerRepository[F fieldAlias, S targeter[F]] struct {
	table    *table[F, S]
	dbGetter DbGetter
}

func newGenericScannerRepository[F fieldAlias, S targeter[F]](
	table *table[F, S],
	dbGetter DbGetter,
) *genericScannerRepository[F, S] {
	return &genericScannerRepository[F, S]{
		table:    table,
		dbGetter: dbGetter,
	}
}

func (g *genericScannerRepository[F, S]) Table() TableI[F, S] {
	return g.table
}

func (g *genericScannerRepository[F, S]) opts(opts ...ScannerCallOptions[F, S]) *scannerCallOptions[F, S] {
	op := &scannerCallOptions[F, S]{
		excludeFields: g.table.allFields,
		copyFields:    g.table.allFields,
	}
	for _, o := range opts {
		o.apply(op)
	}
	return op
}

func (g *genericScannerRepository[F, S]) Insert(
	ctx context.Context,
	entity S,
	_ ...ScannerCallOptions[F, S],
) error {
	_, err := g.table.Execute(
		ctx,
		g.dbGetter(ctx, SqlMutation),
		g.table.Insert().From(getFieldsSetters(entity, g.table.allFields...)...),
	)
	return err
}

func (g *genericScannerRepository[F, S]) InsertRet(
	ctx context.Context,
	entity S,
	_ ...ScannerCallOptions[F, S],
) (S, error) {
	return g.table.QueryRow(
		ctx,
		g.dbGetter(ctx, SqlMutation),
		g.table.Insert().From(getFieldsSetters(entity, g.table.allFields...)...).ReturningAll(),
	)
}

func (g *genericScannerRepository[F, S]) InsertMany(
	ctx context.Context,
	entities []S,
	opts ...ScannerCallOptions[F, S],
) error {
	opt := g.opts(opts...)
	_, err := g.table.CopyFrom(
		ctx,
		g.dbGetter(ctx, SqlMutation),
		entities,
		opt.copyFields...,
	)
	return err
}

func (g *genericScannerRepository[F, S]) Update(
	ctx context.Context,
	entity S,
	_ ...ScannerCallOptions[F, S],
) error {
	_, err := g.table.Execute(
		ctx,
		g.dbGetter(ctx, SqlMutation),
		g.table.Update().Set(getFieldsSetters(entity, g.table.allFields...)...),
	)
	return err
}

func (g *genericScannerRepository[F, S]) UpdateRet(
	ctx context.Context,
	entity S,
	_ ...ScannerCallOptions[F, S],
) (S, error) {
	return g.table.QueryRow(
		ctx,
		g.dbGetter(ctx, SqlMutation),
		g.table.Update().Set(getFieldsSetters(entity, g.table.allFields...)...).ReturningAll(),
	)
}

func (g *genericScannerRepository[F, S]) Upsert(
	ctx context.Context,
	entity S,
	conflictFields []F,
	opts ...ScannerCallOptions[F, S],
) error {
	opt := g.opts(opts...)
	if len(opt.excludeFields) == 0 {
		return errors.Join(ErrEmptyFields, errors.New("excluded fields are empty for upsert"))
	}
	if len(conflictFields) == 0 {
		return errors.Join(ErrEmptyFields, errors.New("conflict fields are empty for upsert"))
	}
	_, err := g.table.Execute(
		ctx,
		g.dbGetter(ctx, SqlMutation),
		g.table.Insert().
			From(getFieldsSetters(
				entity,
				g.table.allFields...,
			)...).
			OnConflict(conflictFields...).
			DoUpdate(opt.excludeFields...),
	)
	return err
}

func (g *genericScannerRepository[F, S]) UpsertRet(
	ctx context.Context,
	entity S,
	conflictFields []F,
	opts ...ScannerCallOptions[F, S],
) (ret S, err error) {
	opt := g.opts(opts...)
	if len(opt.excludeFields) == 0 {
		return ret, errors.Join(ErrEmptyFields, errors.New("excluded fields are empty for upsert"))
	}
	if len(conflictFields) == 0 {
		return ret, errors.Join(ErrEmptyFields, errors.New("conflict fields are empty for upsert"))
	}
	return g.table.QueryRow(
		ctx,
		g.dbGetter(ctx, SqlMutation),
		g.table.Insert().
			From(getFieldsSetters(
				entity,
				g.table.allFields...,
			)...).
			OnConflict(conflictFields...).
			DoUpdate(opt.excludeFields...).ReturningAll(),
	)
}

func (g *genericScannerRepository[F, S]) UpsertIgnore(
	ctx context.Context,
	entity S,
	opts ...ScannerCallOptions[F, S],
) error {
	opt := g.opts(opts...)
	if len(opt.conflictFields) == 0 {
		return errors.Join(ErrEmptyFields, errors.New("copy fields are empty for upsert ignore"))
	}
	_, err := g.table.Execute(
		ctx,
		g.dbGetter(ctx, SqlMutation),
		g.table.Insert().
			From(getFieldsSetters(
				entity,
				g.table.allFields...,
			)...).
			OnConflict(opt.conflictFields...).
			DoNothing(),
	)
	return err
}

func (g *genericScannerRepository[F, S]) GetBy(
	ctx context.Context,
	query ormQuery,
	_ ...ScannerCallOptions[F, S],
) (S, error) {
	return g.table.QueryRow(ctx, g.dbGetter(ctx, SqlQuery), query)
}

func (g *genericScannerRepository[F, S]) ListBy(
	ctx context.Context,
	query ormQuery,
	_ ...ScannerCallOptions[F, S],
) ([]S, error) {
	return g.table.Query(ctx, g.dbGetter(ctx, SqlQuery), query)
}

func (g *genericScannerRepository[F, S]) Exec(
	ctx context.Context,
	query ormQuery,
	_ ...ScannerCallOptions[F, S],
) error {
	_, err := g.table.Execute(ctx, g.dbGetter(ctx, SqlMutation), query)
	return err
}

func (g *genericScannerRepository[F, S]) ExecAffected(
	ctx context.Context,
	query ormQuery,
	_ ...ScannerCallOptions[F, S],
) (int64, error) {
	affected, err := g.table.Execute(ctx, g.dbGetter(ctx, SqlMutation), query)
	return affected, err
}
