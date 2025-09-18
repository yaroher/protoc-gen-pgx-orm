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

type Field struct {
	*protopgx.ParsedField
}

func NewFromProtoField(field *protogen.Field) *Field {
	nullable := isFieldNullable(field)
	array := isFieldArray(field)

	opts := field.Desc.Options().(*descriptorpb.FieldOptions)
	sqlField, _ := proto.GetExtension(opts, protopgx.E_SqlField).(*protopgx.SqlField)
	if sqlField == nil {
		sqlField = &protopgx.SqlField{
			SqlType: &protopgx.SqlType{Type: getSqlTypeFromProtoType(field)},
			Constraints: &protopgx.SqlConstraint{
				DefaultValue: getProtoFieldSQLDefaultValue(field, nullable, array),
			},
		}
	}
	if sqlField.GetSkip() {
		return nil
	}
	if field.Desc.Kind() == protoreflect.MessageKind &&
		!isKnownType(field) &&
		!isSerializedMessage(field) &&
		!isUserDefineCast(field) {
		help.Logger.Warn(
			"skip message field cause not serialized mark",
			zap.String("name", string(field.Desc.FullName())),
		)
		return nil
	}
	if sqlField.GetSqlType().GetType() == protopgx.SqlFiledType_UNSPECIFIED {
		sqlField.SqlType = &protopgx.SqlType{
			Type: getSqlTypeFromProtoType(field),
		}
	}
	return &Field{
		ParsedField: &protopgx.ParsedField{
			ProtoName: string(field.Desc.FullName()),
			Virtual:   false,
			TypeInfo: NewTypeInfo(
				field,
				sqlField.GetSqlType(),
				nullable,
				array,
			),
			Constraint: sqlField.GetConstraints(),
		},
	}
}

func NewFromVirtualField(message *protogen.Message, field *protopgx.SqlVirtualField) *Field {
	if field.GetSqlType() == nil {
		panic("virtual field must have sql type: " + string(message.Desc.FullName()) + "." + field.GetSqlName())
	}
	return &Field{
		ParsedField: &protopgx.ParsedField{
			Virtual: true,
			ProtoName: string(message.Desc.FullName().Parent() +
				"." +
				protoreflect.FullName(strcase.ToCamel(field.GetSqlName()))),
			TypeInfo: NewTypeInfo(
				nil,
				field.GetSqlType(),
				field.GetIsNullable(),
				field.GetIsArray(),
			),
			Constraint: field.GetConstraints(),
		},
	}
}

func (t *Field) SqlFieldName() string {
	return strcase.ToSnake(string(protoreflect.FullName(t.ProtoName).Name()))
}

func (t *Field) PgxType() string {
	return t.TypeInfo.PgxType
}

func (t *Field) IsUserUpCasterNeeded() bool {
	return t.GetTypeInfo().GetUpCasterFn().GetUserDefined()
}
func (t *Field) IsUserDownCasterNeeded() bool {
	return t.GetTypeInfo().GetDownCasterFn().GetUserDefined()
}

func (t *Field) ApplyAbleToCaster() bool {
	if t.Virtual {
		return false
	}
	if t.GetFromOneOfField() != "" {
		return false
	}
	return true
}

func (t *Field) ToDownCaster() string {
	if t.GetFromOneOfField() != "" {
		return strings.ReplaceAll(
			strings.ReplaceAll(
				t.TypeInfo.DownCasterFn.CallSignature,
				"$var",
				"entity.Get"+t.GoName()+"()",
			), "$name", t.TypeInfo.DownCasterFn.Name)
	}
	return strings.ReplaceAll(
		strings.ReplaceAll(
			t.TypeInfo.DownCasterFn.CallSignature,
			"$var",
			"entity."+t.GoName(),
		), "$name", t.TypeInfo.DownCasterFn.Name)
}

func (t *Field) ToUpCaster() string {
	return strings.ReplaceAll(
		strings.ReplaceAll(
			t.TypeInfo.UpCasterFn.CallSignature,
			"$var",
			"model."+t.GoName(),
		), "$name", t.TypeInfo.UpCasterFn.Name)
}

func (t *Field) AvailableOperands() []string {
	ret := []string{
		"CommonOperator",
	}
	if isScalarType(t.TypeInfo.SqlType.Type) {
		ret = append(ret, "ScalarOperator")
	}
	if isStringLikeType(t.TypeInfo.SqlType.Type) {
		ret = append(ret, "LikeOperator")
	}
	if t.TypeInfo.Nullable {
		ret = append(ret, "IsNullOperator")
	}
	return ret
}

func (t *Field) ToSql() string {
	typed := t.TypeInfo.SqlType.GetType().String() // + t.TypeInfo.SqlType.GetAdd()
	if t.TypeInfo.IsArray {
		typed = typed + "[]"
	}
	return fmt.Sprintf(
		"\t%s %s %s",
		strcase.ToSnake(string(protoreflect.FullName(t.ProtoName).Name())),
		strings.ToUpper(typed),
		getFieldConstraint(t.GetConstraint(), t.GetTypeInfo().GetNullable()),
	)
}

func (t *Field) GoName() string {
	return strcase.ToCamel(string(protoreflect.FullName(t.ProtoName).Name()))
}

func CollectFieldsFromMessage(message *protogen.Message) []*Field {
	messageOpts := message.Desc.Options().(*descriptorpb.MessageOptions)
	msgOpts, _ := proto.GetExtension(messageOpts, protopgx.E_SqlTable).(*protopgx.SqlTable)
	if msgOpts == nil {
		return nil
	}
	var fields []*Field
	for _, field := range message.Fields {
		opts := field.Desc.Options().(*descriptorpb.FieldOptions)
		f, _ := proto.GetExtension(opts, protopgx.E_SqlField).(*protopgx.SqlField)
		if f == nil {
			f = &protopgx.SqlField{}
		}
		if f.GetSkip() {
			continue
		}
		generated := NewFromProtoField(field)
		if generated == nil {
			continue
		}
		fields = append(fields, generated)
	}
	for _, virtual := range msgOpts.GetVirtualFields() {
		fields = append(fields, NewFromVirtualField(message, virtual))
	}
	for _, oneof := range message.Oneofs {
		if strings.Contains(oneof.GoName, "X") {
			continue
		}
		for _, oneofField := range oneof.Fields {
			for _, field := range fields {
				if field.GetProtoName() == string(oneofField.Desc.FullName()) {
					help.Logger.Warn(
						"found oneof field",
						zap.String("name", string(oneofField.Desc.FullName())),
						zap.String("ofName", string(oneof.Desc.FullName())),
						zap.String("fName", string(oneofField.Desc.FullName())),
					)
					field.FromOneOfField = string(oneof.Desc.FullName())
					field.FromOneOfFieldType = qualifiedGoIdent(oneofField.GoIdent)
				}
			}
		}
	}
	return fields
}
