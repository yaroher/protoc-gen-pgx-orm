package orm

//
//type StField interface {
//	fieldAlias
//	mustSomeTableColumn()
//}
//
//func (f fieldAliasImpl) mustSomeTableColumn() {}
//func (c *countImpl[F]) mustSomeTableColumn()  {}
//
//type stIdFieldImpl struct {
//	*column[int, StField]
//}
//
//func (f stIdFieldImpl) mustSomeTableColumn() {}
//
//type StScanAble struct {
//	Id int
//}
//
//func (s *StScanAble) values() []any {
//	return []any{&s.Id}
//}
//func (s *StScanAble) getTarget(field string) func() any {
//	switch field {
//	case "id":
//		return func() any { return &s.Id }
//	}
//}
//func (s *StScanAble) getSetter(field StField) func() ValueSetter[StField] {
//	switch field.String() {
//	case "id":
//		return func() ValueSetter[StField] { return NewValueSetter[StField](field, s.Id) }
//
//	}
//}
//func (s *StScanAble) getValue(field StField) func() any {
//	return s.valueMap[field]
//}
//
//type StImpl struct {
//	*table[StField, *StScanAble]
//	Id interface {
//		StField
//		setterOperator[int, StField]
//		ScalarOperator[int, StField]
//		LikeOperator[int, StField]
//		CommonOperator[int, StField]
//	}
//}
//
//func newSomeTable() *StImpl {
//	idColumn := &stIdFieldImpl{column: newColumn[int, StField](fieldAliasImpl("id"))}
//	return &StImpl{
//		table: newTable[StField, *StScanAble](
//			"table_name",
//			func() *StScanAble {
//				value := &StScanAble{}
//				value.targetsMap = map[StField]func() any{
//					idColumn: func() any { return &value.Id },
//				}
//				value.settersMap = map[StField]func() ValueSetter[StField]{
//					idColumn: func() ValueSetter[StField] { return NewValueSetter[StField](idColumn, value.Id) },
//				}
//				value.valueMap = map[StField]func() any{
//					idColumn: func() any { return value.Id },
//				}
//				return value
//			},
//			idColumn,
//		),
//		Id: idColumn,
//	}
//}
//
//var SomeTable = newSomeTable()
//
//type St2Field interface {
//	fieldAlias
//	must2SomeTableColumn()
//}
//
//func (f fieldAliasImpl) must2SomeTableColumn() {}
//
//type st2IdFieldImpl struct {
//	*column[int, St2Field]
//}
//
//func (f st2IdFieldImpl) must2SomeTableColumn() {}
//
//type St2ScanAble struct {
//	Id         int
//	targetsMap map[St2Field]func() any
//	settersMap map[St2Field]func() ValueSetter[St2Field]
//	valueMap   map[St2Field]func() any
//}
//
//func (s *St2ScanAble) values() []any {
//	return []any{&s.Id}
//}
//func (s *St2ScanAble) getTarget(field string) func() any {
//	return s.targetsMap[fieldAliasImpl(field)]
//}
//func (s *St2ScanAble) getSetter(field St2Field) func() ValueSetter[St2Field] {
//	return s.settersMap[field]
//}
//func (s *St2ScanAble) getValue(field St2Field) func() any {
//	return s.valueMap[field]
//}
//
//type St2Impl struct {
//	*table[St2Field, *St2ScanAble]
//	Id interface {
//		St2Field
//		setterOperator[int, St2Field]
//		ScalarOperator[int, St2Field]
//		LikeOperator[int, St2Field]
//		CommonOperator[int, St2Field]
//	}
//}
//
//func new2SomeTable() *St2Impl {
//	idColumn := &st2IdFieldImpl{column: newColumn[int, St2Field](fieldAliasImpl("id_2"))}
//	return &St2Impl{
//		table: newTable[St2Field, *St2ScanAble](
//			"table_name_2",
//			func() *St2ScanAble {
//				value := &St2ScanAble{}
//				value.targetsMap = map[St2Field]func() any{
//					idColumn: func() any { return &value.Id },
//				}
//				value.settersMap = map[St2Field]func() ValueSetter[St2Field]{
//					idColumn: func() ValueSetter[St2Field] { return NewValueSetter[St2Field](idColumn, value.Id) },
//				}
//				value.valueMap = map[St2Field]func() any{
//					idColumn: func() any { return value.Id },
//				}
//				return value
//			},
//			idColumn,
//		),
//		Id: idColumn,
//	}
//}
//
//var SomeTable2 = new2SomeTable()
//
//func TestTyping(t *testing.T) {
//	sql, args := SomeTable.Select(SomeTable.Id.Count()).Where(SomeTable.Id.Eq(123)).Build()
//	t.Log(sql)
//	t.Log(args)
//}
//
////type ProtoMessage struct {
////	proto.Message
////	Id int
////}
////
////func downcastProtoMessageToStScanAble() TypeCaster[*ProtoMessage, *StScanAble] {
////	return func(entity *ProtoMessage) *StScanAble { return &StScanAble{Id: entity.Id} }
////}
////
////func upcastStScanAbleToProtoMessage() TypeCaster[*StScanAble, *ProtoMessage] {
////	return func(model *StScanAble) *ProtoMessage { return &ProtoMessage{Id: model.Id} }
////}
////
////type StRepository = genericRepository[StField, *StScanAble, *ProtoMessage]
////type StRepositoryOption = ProtoCallOption[StField, *StScanAble, *ProtoMessage]
////
////
////type SomeTableModel struct {
////	Id int
////}
////
////func TestA(t *testing.T) {
////	ctx := context.Background()
////	pool, _ := pgxpool.Connect(ctx, "postgres://komeet:developer@localhost:5432/postgres?sslmode=disable")
////	sql, args := SomeTable.DeleteDraftById().Build()
////	t.Logf("SQL: %s", sql)
////	t.Logf("ARGS: %v", args)
////	_, err := SomeTable.Execute(
////		ctx,
////		pool,
////		SomeTable.DeleteDraftById(),
////	)
////	if err != nil {
////		t.Fatal(err)
////	}
////	sql, args = SomeTable.Insert().From(SomeTable.Id.Set(123)).Build()
////	t.Logf("SQL: %s", sql)
////	t.Logf("ARGS: %v", args)
////	_, err = SomeTable.Execute(
////		ctx,
////		pool,
////		SomeTable.Insert().From(SomeTable.Id.Set(123)),
////	)
////	if err != nil {
////		t.Fatal(err)
////	}
////	sql, args = SomeTable.Update().Set(SomeTable.Id.Set(123)).Build()
////	t.Logf("SQL: %s", sql)
////	t.Logf("ARGS: %v", args)
////	_, err = SomeTable.Execute(
////		ctx,
////		pool,
////		SomeTable.Update().Set(SomeTable.Id.Set(123)),
////	)
////	if err != nil {
////		t.Fatal(err)
////	}
////	sql, args = SomeTable.Select(SomeTable.Id).Where(SomeTable.Id.Eq(1234)).Build()
////	t.Logf("SQL: %s", sql)
////	t.Logf("ARGS: %v", args)
////	someData, err := SomeTable.QueryRow(
////		ctx,
////		pool,
////		SomeTable.Select(SomeTable.Id).Where(SomeTable.Id.Eq(1234)),
////	)
////	if err != nil {
////		t.Fatal(err)
////	}
////	t.Logf("DATA: %v", someData)
////}
