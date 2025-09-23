package tabletree

import (
	"fmt"
	"github.com/iancoleman/strcase"
	"github.com/yaroher/protoc-gen-pgx-orm/help"
	"github.com/yaroher/protoc-gen-pgx-orm/protopgx"
	"go.uber.org/zap"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
	"strings"
)

// debugLog prints debug information if debug mode is enabled
func debugLog(format string, args ...interface{}) {
	help.Logger.Sugar().Infof(format, args...)
}

type BackwardRelation struct {
	From      *TableNode
	FromField *Field
	ToField   *Field
}
type Relation struct {
	To        *TableNode
	ToField   *Field
	FromField *Field
}

type Encapsulation struct {
	TargetProtoName protoreflect.FullName
	Fields          []*Field
}

type TableNode struct {
	// Basic table information
	GoIdent         protogen.GoIdent
	Name            protoreflect.FullName
	OverrideSqlName *string
	Fields          []*Field
	Constraints     []string
	Virtual         bool

	OneOfs    map[protoreflect.FullName]*Encapsulation
	Embeds    map[protoreflect.FullName]*Encapsulation
	Relations []*Relation
	Backwards []*BackwardRelation
}

func (t *TableNode) ProtoName() string {
	return string(t.Name.Name())
}

func (t *TableNode) ToSql() string {
	fields := make([]string, 0)
	for _, field := range t.Fields {
		fields = append(fields, field.ToSql())
	}
	for _, constraint := range t.Constraints {
		fields = append(fields, fmt.Sprintf("\t%s", constraint))
	}

	sql := fmt.Sprintf(
		"CREATE TABLE IF NOT EXISTS \"%s\" (\n%s\n);\n",
		t.SqlTableName(),
		strings.Join(fields, ",\n"),
	)
	help.Logger.Info(
		"SQL CODE",
		zap.String("name", t.SqlTableName()),
		zap.String("sql", sql),
	)
	return sql
}

func (t *TableNode) ToOneOffFieldsCast() string {
	buff := strings.Builder{}
	for _, off := range t.OneOfs {
		for _, field := range off.Fields {
			targetName := strcase.ToCamel(string(protoreflect.FullName(field.GetFromOneOfField()).Name()))
			buff.WriteString(fmt.Sprintf("if model.%s.Status != pgtype.Null {\n", field.GoName()))
			buff.WriteString(fmt.Sprintf(
				"entity.%s = &%s{%s: %s}",
				targetName,
				field.GetFromOneOfFieldType(),
				field.GoName(),
				field.ToUpCaster(),
			))
			buff.WriteString("\n")
			buff.WriteString("}\n")
		}
	}
	return buff.String()
}

func (t *TableNode) ToEmbeddedFieldsCast() string {
	buff := strings.Builder{}
	for _, off := range t.Embeds {
		if len(off.Fields) == 0 {
			continue
		}
		targetName := strcase.ToCamel(string(protoreflect.FullName(off.Fields[0].GetFromEmbeddedMessageField()).Name()))
		buff.WriteString(fmt.Sprintf(
			"entity.%s = &%s{\n",
			targetName,
			off.Fields[0].GetFromEmbeddedMessageType(),
		))
		for _, field := range off.Fields {
			buff.WriteString(fmt.Sprintf(
				"%s: %s,",
				field.GoName(),
				field.ToUpCaster(),
			))
			buff.WriteString("\n")
		}
		buff.WriteString("}\n")
	}
	return buff.String()
}

func (t *TableNode) HasVirtualFields() bool {
	for _, field := range t.Fields {
		if field.Virtual {
			return true
		}
	}
	return false
}

func (t *TableNode) GetVirtualFields() []*Field {
	ret := make([]*Field, 0)
	for _, field := range t.Fields {
		if field.Virtual {
			ret = append(ret, field)
		}
	}
	return ret
}

func (t *TableNode) AllUserCasters() []*protopgx.CasterFn {
	ret := make([]*protopgx.CasterFn, 0)
	for _, field := range t.Fields {
		if field.GetTypeInfo().GetUpCasterFn().GetUserDefined() {
			ret = append(ret, field.GetTypeInfo().GetUpCasterFn())
		}
		if field.GetTypeInfo().GetDownCasterFn().GetUserDefined() {
			ret = append(ret, field.GetTypeInfo().GetDownCasterFn())
		}
	}
	return ret
}

func (t *TableNode) FindField(name string) (*Field, bool) {
	for _, field := range t.Fields {
		if strcase.ToSnake(string(protoreflect.FullName(field.ProtoName).Name())) == name {
			return field, true
		}
	}
	return nil, false
}

func (t *TableNode) SqlTableName() string {
	if t.OverrideSqlName != nil {
		return *t.OverrideSqlName
	}
	return strcase.ToSnake(string(t.Name.Name()))
}

func (t *TableNode) GoName() string {
	return strcase.ToCamel(string(t.Name.Name()))
}

func CollectTablesFromProto(files []*protogen.File) []*TableNode {
	tables := make([]*TableNode, 0)
	for _, file := range files {
		if len(file.Messages) == 0 {
			continue
		}
		for _, message := range file.Messages {
			opts := message.Desc.Options().(*descriptorpb.MessageOptions)
			sqlTable, ok := proto.GetExtension(opts, protopgx.E_SqlTable).(*protopgx.SqlTable)
			if !ok || sqlTable == nil || sqlTable.GetGenerate() == false {
				continue
			}
			t := &TableNode{
				GoIdent:         message.GoIdent,
				Name:            message.Desc.FullName(),
				OverrideSqlName: sqlTable.TableName,
				Fields:          CollectFieldsFromMessage(message),
				Constraints:     sqlTable.GetConstraints(),
				OneOfs:          make(map[protoreflect.FullName]*Encapsulation),
				Embeds:          make(map[protoreflect.FullName]*Encapsulation),
			}
			for _, field := range t.Fields {
				if field.GetFromOneOfField() != "" {
					if off, ok := t.OneOfs[protoreflect.FullName(field.GetFromOneOfField())]; ok {
						off.Fields = append(off.Fields, field)
					} else {
						t.OneOfs[protoreflect.FullName(field.GetFromOneOfField())] = &Encapsulation{
							TargetProtoName: protoreflect.FullName(field.GetFromOneOfField()),
							Fields:          []*Field{field},
						}
					}
				}
			}
			for _, field := range t.Fields {
				if field.Embedded {
					if off, ok := t.Embeds[protoreflect.FullName(field.GetFromEmbeddedMessageField())]; ok {
						off.Fields = append(off.Fields, field)
					} else {
						t.Embeds[protoreflect.FullName(field.GetFromEmbeddedMessageField())] = &Encapsulation{
							TargetProtoName: protoreflect.FullName(field.GetFromEmbeddedMessageField()),
							Fields:          []*Field{field},
						}
					}
				}
			}
			tables = append(tables, t)
		}
	}
	return collectRelations(files, tables)
}
func findTable(tables []*TableNode, name protoreflect.FullName) (*TableNode, bool) {
	for _, t := range tables {
		if t.Name == name || t.GoIdent.GoName == string(name) {
			return t, true
		}
	}
	return nil, false
}

//goland:noinspection t
func collectRelations(files []*protogen.File, tables []*TableNode) []*TableNode {
	for _, file := range files {
		for _, message := range file.Messages {
			for _, field := range message.Fields {
				opts := field.Desc.Options().(*descriptorpb.FieldOptions)
				relation, _ := proto.GetExtension(opts, protopgx.E_SqlRelation).(*protopgx.SqlRelation)
				if relation == nil {
					continue
				}
				if field.Desc.Kind() != protoreflect.MessageKind || !field.Desc.IsList() {
					panic("relations support only repeated Message fields")
				}
				switch relation.GetRelation().(type) {
				case *protopgx.SqlRelation_OneToMany_:
					targetTable, ok := findTable(tables, field.Message.Desc.FullName())
					if !ok {
						panic(fmt.Sprintf("target table %s for relatrion %s not found", field.Message.Desc.FullName(), field.Desc.Name()))
					}
					sourceTable, ok := findTable(tables, message.Desc.FullName())
					if !ok {
						panic(fmt.Sprintf("source table %s for relatrion %s not found", message.Desc.FullName(), field.Desc.Name()))
					}
					sourceField, ok := sourceTable.FindField("id")
					if !ok {
						panic(fmt.Sprintf("source 'id' field from %s for relatrion %s not found in message", field.Desc.Name(), field.Desc.Name()))
					}
					var targetField *Field
					if relation.GetOneToMany().GetExistedField() {
						if relation.GetOneToMany().GetRefName() == "" {
							panic(fmt.Sprintf("target field %s not found cause ref_name is empty", field.Desc.Name()))
						}
						targetField, ok = targetTable.FindField(relation.GetOneToMany().GetRefName())
						if !ok {
							panic(fmt.Sprintf("target field %s for relatrion %s not found", relation.GetOneToMany().GetRefName(), field.Desc.Name()))
						}
					} else {
						onDelete := ""
						if relation.GetOneToMany().GetOnDeleteCascade() {
							onDelete = " ON DELETE CASCADE"
						}
						targetField = &Field{
							ParsedField: &protopgx.ParsedField{
								Virtual: true,
								ProtoName: help.StringOrDefault(
									relation.GetOneToMany().GetRefName(),
									fmt.Sprintf("%s_id", strings.ToLower(string(message.Desc.Name()))),
								),
								TypeInfo: sourceField.GetTypeInfo(),
								Constraint: &protopgx.SqlConstraint{
									Constraint: help.StringOrDefault(
										relation.GetOneToMany().GetConstraint(),
										fmt.Sprintf("REFERENCES %s (id)%s", sourceTable.SqlTableName(), onDelete),
									),
								},
							},
						}
						targetTable.Fields = append(targetTable.Fields, targetField)
					}
					targetTable.Relations = append(targetTable.Relations, &Relation{
						To:        sourceTable,
						ToField:   targetField,
						FromField: sourceField,
					})
					sourceTable.Backwards = append(sourceTable.Backwards, &BackwardRelation{
						From:      targetTable,
						FromField: targetField,
						ToField:   sourceField,
					})

				case *protopgx.SqlRelation_ManyToMany_:
					targetTable, ok := findTable(tables, field.Message.Desc.FullName())
					if !ok {
						panic(fmt.Sprintf("target table %s for relatrion %s not found", field.Message.Desc.FullName(), field.Desc.Name()))
					}
					sourceTable, ok := findTable(tables, message.Desc.FullName())
					if !ok {
						panic(fmt.Sprintf("source table %s for relatrion %s not found", message.Desc.FullName(), field.Desc.Name()))
					}
					sourceField, ok := sourceTable.FindField("id")
					if !ok {
						panic(fmt.Sprintf("source 'id' field from %s for relatrion %s not found", field.Desc.Name(), field.Desc.Name()))
					}
					targetField, ok := targetTable.FindField("id")
					if !ok {
						panic(fmt.Sprintf("target 'id' field from %s for relatrion %s not found", field.Desc.Name(), field.Desc.Name()))
					}
					fwdOnDelete := ""
					if relation.GetManyToMany().GetRefOnDeleteCascade() {
						fwdOnDelete = " ON DELETE CASCADE"
					}
					newForwardRelField := &Field{
						ParsedField: &protopgx.ParsedField{
							ProtoName: fmt.Sprintf("%s_id", targetTable.SqlTableName()),
							Virtual:   true,
							TypeInfo:  targetField.GetTypeInfo(),
							Constraint: &protopgx.SqlConstraint{
								Constraint: help.StringOrDefault(
									relation.GetManyToMany().GetRefConstraint(),
									fmt.Sprintf(
										"NOT NULL REFERENCES %s (id)%s",
										targetTable.SqlTableName(), fwdOnDelete,
									),
								),
							},
						},
					}
					bwdOnDelete := ""
					if relation.GetManyToMany().GetBackRefOnDeleteCascade() {
						bwdOnDelete = " ON DELETE CASCADE"
					}
					newBackwardRelField := &Field{
						ParsedField: &protopgx.ParsedField{
							ProtoName: fmt.Sprintf("%s_id", sourceTable.SqlTableName()),
							Virtual:   true,
							TypeInfo:  targetField.GetTypeInfo(),
							Constraint: &protopgx.SqlConstraint{
								Constraint: help.StringOrDefault(
									relation.GetManyToMany().GetBackRefConstraint(),
									fmt.Sprintf(
										"NOT NULL REFERENCES %s (id)%s",
										sourceTable.SqlTableName(), bwdOnDelete,
									),
								),
							},
						},
					}
					virtualTable := &TableNode{
						GoIdent: targetTable.GoIdent,
						Name: protoreflect.FullName(help.StringOrDefault(
							relation.GetManyToMany().GetTable().GetTableName(),
							fmt.Sprintf(
								"%s.%s%s",
								message.Desc.FullName().Parent(),
								targetTable.Name.Name(),
								sourceTable.Name.Name(),
							),
						)),
						Fields: []*Field{
							newForwardRelField,
							newBackwardRelField,
						},
						Constraints: help.ListStringOrDefault(relation.GetManyToMany().GetTable().GetConstraints(), []string{
							fmt.Sprintf(
								"UNIQUE (%s, %s)",
								protoreflect.FullName(sourceField.ProtoName).Name(),
								protoreflect.FullName(targetField.ProtoName).Name(),
							),
						}),
						Virtual: true,
					}
					for _, f := range relation.GetManyToMany().GetTable().GetVirtualFields() {
						virtualTable.Fields = append(virtualTable.Fields, NewFromVirtualField(message, f))
					}
					virtualTable.Relations = append(virtualTable.Relations, &Relation{
						To:        sourceTable,
						ToField:   newBackwardRelField,
						FromField: sourceField,
					})
					virtualTable.Relations = append(virtualTable.Relations, &Relation{
						To:        targetTable,
						ToField:   newForwardRelField,
						FromField: targetField,
					})
					tables = append(tables, virtualTable)

					sourceTable.Backwards = append(sourceTable.Backwards, &BackwardRelation{
						From:      virtualTable,
						FromField: newBackwardRelField,
						ToField:   sourceField,
					})
					targetTable.Backwards = append(targetTable.Backwards, &BackwardRelation{
						From:      virtualTable,
						FromField: newForwardRelField,
						ToField:   targetField,
					})
				}
			}
		}
	}
	return tables
}
