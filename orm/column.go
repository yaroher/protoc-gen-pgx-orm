package orm

type LikeOperator[V any, F fieldAlias] interface {
	Like(string) Clause[F]
	NotLike(string) Clause[F]
	ILike(string) Clause[F]
	NotILike(string) Clause[F]
}
type IsNullOperator[V any, F fieldAlias] interface {
	IsNull() Clause[F]
	IsNotNull() Clause[F]
}

type binaryOperator[V any, F fieldAlias] interface {
	Gt(V) Clause[F]
	Gte(V) Clause[F]
	Lt(V) Clause[F]
	Lte(V) Clause[F]
}
type betweenOperator[V any, F fieldAlias] interface {
	Between(V, V) Clause[F]
	NotBetween(V, V) Clause[F]
}
type inOperator[V any, F fieldAlias] interface {
	In(...V) Clause[F]
	NotIn(...V) Clause[F]
	InOf(query ormQuery) Clause[F]
	InRaw(string, ...any) Clause[F]
}
type anyOperator[V any, F fieldAlias] interface {
	Any(...V) Clause[F]
	NotAny(...V) Clause[F]
	AnyOf(query ormQuery) Clause[F]
	AnyRaw(string, ...any) Clause[F]
}
type ScalarOperator[V any, F fieldAlias] interface {
	binaryOperator[V, F]
	anyOperator[V, F]
	inOperator[V, F]
	betweenOperator[V, F]
}

type setterOperator[V any, F fieldAlias] interface {
	SetExpr(string) *valueSetterImpl[F]
	Set(V) *valueSetterImpl[F]
	SetRaw(sql string, value V) *valueSetterImpl[F]
}
type eqOperator[V any, F fieldAlias] interface {
	Eq(V) Clause[F]
	Neq(V) Clause[F]
	EqOf(query ormQuery) Clause[F]
	EqRaw(string, ...any) Clause[F]
}
type logicalOperator[V any, F fieldAlias] interface {
	Or(clause ...Clause[F]) Clause[F]
	And(clause ...Clause[F]) Clause[F]
	Not(clause Clause[F]) Clause[F]
}
type anyQOperator[V any, F fieldAlias] interface {
	Of(query ormQuery) Clause[F]
	NotOf(query ormQuery) Clause[F]
	Raw(string, string, ...any) Clause[F]
	NotRaw(string, string, ...any) Clause[F]
	ExistsOf(query ormQuery) Clause[F]
	ExistsRaw(string, ...any) Clause[F]
}
type CommonOperator[V any, F fieldAlias] interface {
	Count() *countImpl[F]
	setterOperator[V, F]
	eqOperator[V, F]
	logicalOperator[V, F]
	anyQOperator[V, F]
}

type countImpl[F fieldAlias] struct {
	field F
}

func (c *countImpl[F]) IsCount() bool { return true }
func (c *countImpl[F]) String() string {
	return c.field.String()
}

type column[V any, F fieldAlias] struct {
	fieldAlias  F
	constructor func() *V
}

func (f *column[V, F]) IsCount() bool { return false }
func newColumn[V any, F fieldAlias](fa F) *column[V, F] {
	return &column[V, F]{
		fieldAlias: fa,
	}
}

func (f *column[V, F]) Count() *countImpl[F] {
	return &countImpl[F]{
		field: f.fieldAlias,
	}
}

func (f *column[V, F]) Set(val V) *valueSetterImpl[F] {
	return &valueSetterImpl[F]{
		field: f.fieldAlias,
		value: val,
	}
}
func (f *column[V, F]) SetExpr(expr string) *valueSetterImpl[F] {
	return &valueSetterImpl[F]{
		field: f.fieldAlias,
		expr:  expr,
	}
}
func (f *column[V, F]) SetRaw(sql string, value V) *valueSetterImpl[F] {
	return &valueSetterImpl[F]{
		field: f.fieldAlias,
		raw:   &RawExprClause[F]{sql, []any{value}},
	}
}
func (f *column[V, F]) String() string {
	return f.fieldAlias.String()
}

func (f *column[V, F]) Eq(val V) Clause[F] {
	return &FieldClause[F]{Field: f.fieldAlias, Operator: "=", Right: &ParamExprClause[F]{Value: val}}
}
func (f *column[V, F]) Neq(val V) Clause[F] {
	return &FieldClause[F]{Field: f.fieldAlias, Operator: "!=", Right: &ParamExprClause[F]{Value: val}}
}
func (f *column[V, F]) EqOf(query ormQuery) Clause[F] {
	return &FieldClause[F]{Field: f.fieldAlias, Operator: "=", Right: &SubQueryExprClause[F]{query}}
}
func (f *column[V, F]) EqRaw(sql string, args ...any) Clause[F] {
	return &FieldClause[F]{Field: f.fieldAlias, Operator: "=", Right: &RawExprClause[F]{sql, args}}
}

func (f *column[V, F]) Gt(val V) Clause[F] {
	return &FieldClause[F]{Field: f.fieldAlias, Operator: ">", Right: &ParamExprClause[F]{Value: val}}
}
func (f *column[V, F]) Gte(val V) Clause[F] {
	return &FieldClause[F]{Field: f.fieldAlias, Operator: ">=", Right: &ParamExprClause[F]{Value: val}}
}
func (f *column[V, F]) Lt(val V) Clause[F] {
	return &FieldClause[F]{Field: f.fieldAlias, Operator: "<", Right: &ParamExprClause[F]{Value: val}}
}
func (f *column[V, F]) Lte(val V) Clause[F] {
	return &FieldClause[F]{Field: f.fieldAlias, Operator: "<=", Right: &ParamExprClause[F]{Value: val}}
}

func (f *column[V, F]) In(vals ...V) Clause[F] {
	return &FieldClause[F]{Field: f.fieldAlias, Operator: "IN", Right: &SliceExprClause[F]{vals}}
}
func (f *column[V, F]) NotIn(vals ...V) Clause[F] {
	return &FieldClause[F]{Field: f.fieldAlias, Operator: "IN", Right: &SliceExprClause[F]{vals}, Negate: true}
}
func (f *column[V, F]) InOf(query ormQuery) Clause[F] {
	return &FieldClause[F]{Field: f.fieldAlias, Operator: "IN", Right: &SubQueryExprClause[F]{query}}
}
func (f *column[V, F]) InRaw(sql string, args ...any) Clause[F] {
	return &FieldClause[F]{Field: f.fieldAlias, Operator: "IN", Right: &RawExprClause[F]{sql, args}}
}

func (f *column[V, F]) Any(vals ...V) Clause[F] {
	return &FieldClause[F]{Field: f.fieldAlias, Operator: "= ANY", Right: &SliceExprClause[F]{vals}}
}
func (f *column[V, F]) NotAny(vals ...V) Clause[F] {
	return &FieldClause[F]{Field: f.fieldAlias, Operator: "!= ALL", Right: &SliceExprClause[F]{vals}}
}
func (f *column[V, F]) AnyOf(query ormQuery) Clause[F] {
	return &FieldClause[F]{Field: f.fieldAlias, Operator: "= ANY", Right: &SubQueryExprClause[F]{query}}
}
func (f *column[V, F]) AnyRaw(sql string, args ...any) Clause[F] {
	return &FieldClause[F]{Field: f.fieldAlias, Operator: "= ANY", Right: &RawExprClause[F]{sql, args}}
}

func (f *column[V, F]) Between(lower, upper V) Clause[F] {
	return &AndClause[F]{Clauses: []Clause[F]{
		&FieldClause[F]{Field: f.fieldAlias, Operator: ">=", Right: &ParamExprClause[F]{Value: lower}},
		&FieldClause[F]{Field: f.fieldAlias, Operator: "<=", Right: &ParamExprClause[F]{Value: upper}},
	}}
}
func (f *column[V, F]) NotBetween(lower, upper V) Clause[F] {
	return &NotClause[F]{Inner: f.Between(lower, upper)}
}

func (f *column[V, F]) Like(pattern string) Clause[F] {
	return &FieldClause[F]{Field: f.fieldAlias, Operator: "LIKE", Right: &ParamExprClause[F]{Value: pattern, LikeWrapp: true}}
}
func (f *column[V, F]) NotLike(pattern string) Clause[F] {
	return &FieldClause[F]{Field: f.fieldAlias, Operator: "LIKE", Right: &ParamExprClause[F]{Value: pattern, LikeWrapp: true}, Negate: true}
}
func (f *column[V, F]) ILike(pattern string) Clause[F] {
	return &FieldClause[F]{Field: f.fieldAlias, Operator: "ILIKE", Right: &ParamExprClause[F]{Value: pattern, LikeWrapp: true}}
}
func (f *column[V, F]) NotILike(pattern string) Clause[F] {
	return &FieldClause[F]{Field: f.fieldAlias, Operator: "ILIKE", Right: &ParamExprClause[F]{Value: pattern, LikeWrapp: true}, Negate: true}
}

func (f *column[V, F]) IsNull() Clause[F] {
	return &FieldClause[F]{Field: f.fieldAlias, Operator: "IS NULL", Right: &RawExprClause[F]{SQL: ""}}
}
func (f *column[V, F]) IsNotNull() Clause[F] {
	return &FieldClause[F]{Field: f.fieldAlias, Operator: "IS NOT NULL", Right: &RawExprClause[F]{SQL: ""}}
}

func (f *column[V, F]) Or(clauses ...Clause[F]) Clause[F] {
	return &OrClause[F]{Clauses: clauses}
}
func (f *column[V, F]) And(clauses ...Clause[F]) Clause[F] {
	return &AndClause[F]{Clauses: clauses}
}
func (f *column[V, F]) Not(clause Clause[F]) Clause[F] {
	return &NotClause[F]{Inner: clause}
}

func (f *column[V, F]) Of(query ormQuery) Clause[F] {
	return &FieldClause[F]{Field: f.fieldAlias, Operator: "=", Right: &SubQueryExprClause[F]{query}}
}
func (f *column[V, F]) NotOf(query ormQuery) Clause[F] {
	return &FieldClause[F]{Field: f.fieldAlias, Operator: "!=", Right: &SubQueryExprClause[F]{query}}
}
func (f *column[V, F]) Raw(operator string, sql string, args ...any) Clause[F] {
	return &FieldClause[F]{Field: f.fieldAlias, Operator: operator, Right: &RawExprClause[F]{sql, args}}
}
func (f *column[V, F]) NotRaw(operator string, sql string, args ...any) Clause[F] {
	return &NotClause[F]{Inner: f.Raw(operator, sql, args...)}
}
func (f *column[V, F]) ExistsOf(query ormQuery) Clause[F] {
	return &ExistsClause[F]{SubQuery: &SubQueryExprClause[F]{query}, Negate: false}
}
func (f *column[V, F]) ExistsRaw(sql string, args ...any) Clause[F] {
	return &ExistsClause[F]{SubQuery: &RawExprClause[F]{sql, args}, Negate: false}
}
