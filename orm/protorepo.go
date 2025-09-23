package orm

import (
	"context"
	"google.golang.org/protobuf/proto"
)

type protoCallOptions[F fieldAlias, S targeter[F], T proto.Message] struct {
	excludeFields  []F
	conflictFields []F
	copyFields     []F
}

func (o *protoCallOptions[F, S, T]) toScannerCallOptions() []ScannerCallOptions[F, S] {
	ret := make([]ScannerCallOptions[F, S], 0)
	if len(o.excludeFields) > 0 {
		ret = append(ret, WithScannerExcludeFields[F, S](o.excludeFields...))
	}
	if len(o.conflictFields) > 0 {
		ret = append(ret, WithScannerConflictFields[F, S](o.conflictFields...))
	}
	if len(o.copyFields) > 0 {
		ret = append(ret, WithScannerCopyFields[F, S](o.copyFields...))
	}
	return ret
}

type ProtoCallOption[F fieldAlias, S targeter[F], T proto.Message] interface {
	apply(*protoCallOptions[F, S, T])
}

type callOptionsFn[F fieldAlias, S targeter[F], T proto.Message] func(*protoCallOptions[F, S, T])

func (f callOptionsFn[F, S, T]) apply(opts *protoCallOptions[F, S, T]) {
	f(opts)
}

func WithExcludeFields[F fieldAlias, S targeter[F], T proto.Message](fields ...F) ProtoCallOption[F, S, T] {
	return callOptionsFn[F, S, T](func(opts *protoCallOptions[F, S, T]) {
		opts.excludeFields = fields
	})
}

func WithConflictFields[F fieldAlias, S targeter[F], T proto.Message](fields ...F) ProtoCallOption[F, S, T] {
	return callOptionsFn[F, S, T](func(opts *protoCallOptions[F, S, T]) {
		opts.conflictFields = fields
	})
}

func WithCopyFields[F fieldAlias, S targeter[F], T proto.Message](fields ...F) ProtoCallOption[F, S, T] {
	return callOptionsFn[F, S, T](func(opts *protoCallOptions[F, S, T]) {
		opts.copyFields = fields
	})
}

type ProtoRepository[F fieldAlias, S targeter[F], T proto.Message] interface {
	Table() TableI[F, S]
	ScannerRepository() ScannerRepository[F, S]
	Insert(ctx context.Context, entity T, opts ...ProtoCallOption[F, S, T]) error
	InsertRet(ctx context.Context, entity T, opts ...ProtoCallOption[F, S, T]) (T, error)
	InsertMany(ctx context.Context, entities []T, opts ...ProtoCallOption[F, S, T]) error
	Update(ctx context.Context, entity T, clause Clause[F], opts ...ProtoCallOption[F, S, T]) error
	UpdateRet(ctx context.Context, entity T, clause Clause[F], opts ...ProtoCallOption[F, S, T]) (T, error)
	//Upsert(ctx context.Context, entity T, conflictFields []F, opts ...ProtoCallOption[F, S, T]) error
	//UpsertRet(ctx context.Context, entity T, conflictFields []F, opts ...ProtoCallOption[F, S, T]) (T, error)
	//UpsertIgnore(ctx context.Context, entity T, opts ...ProtoCallOption[F, S, T]) error

	GetBy(ctx context.Context, query ormQuery, opts ...ProtoCallOption[F, S, T]) (T, error)
	ListBy(ctx context.Context, query ormQuery, opts ...ProtoCallOption[F, S, T]) ([]T, error)
	Exec(ctx context.Context, query ormQuery, opts ...ProtoCallOption[F, S, T]) error
	ExecAffected(ctx context.Context, query ormQuery, opts ...ProtoCallOption[F, S, T]) (int64, error)
}

type genericRepository[F fieldAlias, S targeter[F], T proto.Message] struct {
	scannerRepo *genericScannerRepository[F, S]
	downcast    func(T) S
	upcast      func(S) T
	defaultOpts []ProtoCallOption[F, S, T]
}

func newGenericRepository[F fieldAlias, S targeter[F], T proto.Message](
	genericScannerRepo *genericScannerRepository[F, S],
	downcast func(T) S,
	upcast func(S) T,
	defaultOpts ...ProtoCallOption[F, S, T],
) *genericRepository[F, S, T] {
	return &genericRepository[F, S, T]{
		scannerRepo: genericScannerRepo,
		downcast:    downcast,
		upcast:      upcast,
		defaultOpts: defaultOpts,
	}
}

func (g *genericRepository[F, S, T]) Table() TableI[F, S] {
	return g.scannerRepo.table
}
func (g *genericRepository[F, S, T]) ScannerRepository() ScannerRepository[F, S] {
	return g.scannerRepo
}

func (g *genericRepository[F, S, T]) opts(opts []ProtoCallOption[F, S, T]) *protoCallOptions[F, S, T] {
	op := &protoCallOptions[F, S, T]{
		excludeFields: g.scannerRepo.table.allFields,
		copyFields:    g.scannerRepo.table.allFields,
	}
	for _, o := range g.defaultOpts {
		o.apply(op)
	}
	for _, o := range opts {
		o.apply(op)
	}
	return op
}
func (g *genericRepository[F, S, T]) Insert(
	ctx context.Context,
	entity T,
	opts ...ProtoCallOption[F, S, T],
) error {
	return g.scannerRepo.Insert(ctx, g.downcast(entity), g.opts(opts).toScannerCallOptions()...)
}
func (g *genericRepository[F, S, T]) InsertRet(
	ctx context.Context,
	entity T,
	opts ...ProtoCallOption[F, S, T],
) (T, error) {
	model, err := g.scannerRepo.InsertRet(ctx, g.downcast(entity), g.opts(opts).toScannerCallOptions()...)
	return g.upcast(model), err
}
func (g *genericRepository[F, S, T]) InsertMany(
	ctx context.Context,
	entities []T,
	opts ...ProtoCallOption[F, S, T],
) error {
	models := make([]S, 0)
	for _, e := range entities {
		models = append(models, g.downcast(e))
	}
	return g.scannerRepo.InsertMany(ctx, models, g.opts(opts).toScannerCallOptions()...)
}

func (g *genericRepository[F, S, T]) Update(
	ctx context.Context,
	entity T,
	clause Clause[F],
	opts ...ProtoCallOption[F, S, T],
) error {
	return g.scannerRepo.Update(ctx, g.downcast(entity), clause, g.opts(opts).toScannerCallOptions()...)
}

func (g *genericRepository[F, S, T]) UpdateRet(
	ctx context.Context,
	entity T,
	clause Clause[F],
	opts ...ProtoCallOption[F, S, T],
) (T, error) {
	model, err := g.scannerRepo.UpdateRet(ctx, g.downcast(entity), clause, g.opts(opts).toScannerCallOptions()...)
	return g.upcast(model), err
}

// func (g *genericRepository[F, S, T]) Upsert(
//
//	ctx context.Context,
//	entity T,
//	conflictFields []F,
//	opts ...ProtoCallOption[F, S, T],
//
//	) error {
//		return g.scannerRepo.Upsert(ctx, g.downcast(entity), conflictFields, g.opts(opts).toScannerCallOptions()...)
//	}
//
// func (g *genericRepository[F, S, T]) UpsertRet(
//
//	ctx context.Context,
//	entity T,
//	conflictFields []F,
//	opts ...ProtoCallOption[F, S, T],
//
//	) (T, error) {
//		model, err := g.scannerRepo.UpsertRet(ctx, g.downcast(entity), conflictFields, g.opts(opts).toScannerCallOptions()...)
//		return g.upcast(model), err
//	}
//
// func (g *genericRepository[F, S, T]) UpsertIgnore(
//
//	ctx context.Context,
//	entity T,
//	opts ...ProtoCallOption[F, S, T],
//
//	) error {
//		return g.scannerRepo.UpsertIgnore(ctx, g.downcast(entity), g.opts(opts).toScannerCallOptions()...)
//	}
func (g *genericRepository[F, S, T]) GetBy(
	ctx context.Context,
	query ormQuery,
	opts ...ProtoCallOption[F, S, T],
) (T, error) {
	opt := g.opts(opts)
	model, err := g.scannerRepo.GetBy(ctx, query, opt.toScannerCallOptions()...)
	entity := g.upcast(model)
	if err != nil {
		return entity, err
	}
	return entity, nil
}
func (g *genericRepository[F, S, T]) ListBy(
	ctx context.Context,
	query ormQuery,
	opts ...ProtoCallOption[F, S, T],
) ([]T, error) {
	opt := g.opts(opts)
	models, err := g.scannerRepo.ListBy(ctx, query, opt.toScannerCallOptions()...)
	if err != nil {
		return nil, err
	}
	rets := make([]T, 0, len(models))
	for _, model := range models {
		entity := g.upcast(model)
		rets = append(rets, entity)
	}
	return rets, nil
}
func (g *genericRepository[F, S, T]) Exec(
	ctx context.Context,
	query ormQuery,
	opts ...ProtoCallOption[F, S, T],
) error {
	return g.scannerRepo.Exec(ctx, query, g.opts(opts).toScannerCallOptions()...)
}

func (g *genericRepository[F, S, T]) ExecAffected(
	ctx context.Context,
	query ormQuery,
	opts ...ProtoCallOption[F, S, T],
) (int64, error) {
	return g.scannerRepo.ExecAffected(ctx, query, g.opts(opts).toScannerCallOptions()...)
}
