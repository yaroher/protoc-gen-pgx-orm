package tabletree

import (
	"fmt"
	"github.com/iancoleman/strcase"
	"github.com/yaroher/protoc-gen-pgx-orm/help"
	"github.com/yaroher/protoc-gen-pgx-orm/protopgx"
	"go.uber.org/zap"
	"go/token"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
	"path"
	"strings"
	"unicode"
	"unicode/utf8"
)

const plainSignature = "$name($var)"
const noneSignature = "$var"

func cleanTypeName(name string) string {
	ret := name
	hasPtr := strings.HasPrefix(ret, "*")
	hasSlice := strings.HasPrefix(ret, "[]")
	cleanType := strings.Replace(strings.ReplaceAll(name, "*", ""), "[]", "", 1)
	if strings.Contains(cleanType, ".") {
		ret = cleanType[strings.LastIndex(cleanType, ".")+1:]
	}
	if cleanType == "[]byte" {
		cleanType = "ByteSlice"
	} else {
		cleanType = strcase.ToCamel(ret)
	}
	if hasPtr {
		ret = fmt.Sprintf("Ptr_%s", cleanType)
	}
	if hasSlice {
		ret = fmt.Sprintf("Slice_%s", cleanType)
	}
	help.Logger.Debug(
		"cleanTypeName",
		zap.String("name", name),
		zap.String("clean", cleanType),
		zap.String("ret", strcase.ToCamel(ret)),
	)
	return strcase.ToCamel(ret)
}

func genericSignature(genericType string, withPtr bool) string {
	if withPtr {
		return fmt.Sprintf("$name[*%s]($var)", genericType)
	}
	return fmt.Sprintf("$name[%s]($var)", genericType)
}

type castDest string

const (
	From castDest = "From"
	To   castDest = "To"
)

func casterName(prefix string, dest castDest, info *protopgx.ParsedField_TypeInfo) string {
	return strcase.ToCamel(fmt.Sprintf(
		"%s%s%s",
		prefix,
		dest,
		cleanTypeName(info.PgxType),
	))
}

func isFieldNullable(field *protogen.Field) bool {
	oneOf := isFieldOneOf(field)
	help.Logger.Debug(
		"isFieldNullable",
		zap.String("name", string(field.Desc.FullName())),
		zap.Bool("oneOf", oneOf),
	)
	return field.Desc.HasOptionalKeyword() || oneOf
}

func isFieldArray(field *protogen.Field) bool {
	array := field.Desc.IsList()
	opts := field.Desc.Options().(*descriptorpb.FieldOptions)
	sqlField, _ := proto.GetExtension(opts, protopgx.E_SqlField).(*protopgx.SqlField)
	if sqlField.GetSqlType().GetForceNotArray() {
		if array {
			help.Logger.Warn(
				"use not array type for field cause force not array",
				zap.String("name", string(field.Desc.FullName())),
			)
		}
		array = false
	}
	return array
}
func isSerializedMessage(field *protogen.Field) bool {
	opts := field.Desc.Options().(*descriptorpb.FieldOptions)
	sqlField, _ := proto.GetExtension(opts, protopgx.E_SqlField).(*protopgx.SqlField)
	return sqlField.GetSerializedMessage()
}

func isUserDefineCast(field *protogen.Field) bool {
	opts := field.Desc.Options().(*descriptorpb.FieldOptions)
	sqlField, _ := proto.GetExtension(opts, protopgx.E_SqlField).(*protopgx.SqlField)
	return sqlField.GetSqlType().GetUserCast()
}

func isEmbeddedMessage(field *protogen.Field) bool {
	opts := field.Desc.Options().(*descriptorpb.FieldOptions)
	sqlField, _ := proto.GetExtension(opts, protopgx.E_SqlField).(*protopgx.SqlField)
	return sqlField.GetEmbeddedMessage()
}

func userDefinedCastType(left, right string) string {
	return fmt.Sprintf("TypeCaster[%s, %s]", left, right)
}

const downcastPrefix = "downcast"
const upcastPrefix = "upcast"

func userDefinedCastName(prefix string, name string) string {
	return strcase.ToLowerCamel(fmt.Sprintf("%s%s", prefix, name))
}

// GoSanitized converts a string to a valid Go identifier.
func goSanitized(s string) string {
	// Sanitize the input to the set of valid characters,
	// which must be '_' or be in the Unicode L or N categories.
	s = strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return r
		}
		return '_'
	}, s)

	// Prepend '_' in the event of a Go keyword conflict or if
	// the identifier is invalid (does not start in the Unicode L category).
	r, _ := utf8.DecodeRuneInString(s)
	if token.Lookup(s).IsKeyword() || !unicode.IsLetter(r) {
		return "_" + s
	}
	return s
}

func cleanPackageName(name string) protogen.GoPackageName {
	return protogen.GoPackageName(goSanitized(name))
}

func qualifiedGoIdent(ident protogen.GoIdent) string {
	//if ident.GoImportPath == g.goImportPath {
	//	return ident.GoName
	//}
	//if packageName, ok := g.packageNames[ident.GoImportPath]; ok {
	//	return string(packageName) + "." + ident.GoName
	//}
	packageName := cleanPackageName(path.Base(string(ident.GoImportPath)))
	return string(packageName) + "." + ident.GoName
}

// fieldGoType returns the Go type used for a field.
//
// If it returns pointer=true, the struct field is a pointer to the type.
func fieldGoType(field *protogen.Field) string {
	goType := ""
	switch field.Desc.Kind() {
	case protoreflect.BoolKind:
		goType = "bool"
	case protoreflect.EnumKind:
		goType = qualifiedGoIdent(field.Enum.GoIdent)
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
		goType = "int32"
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind:
		goType = "uint32"
	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		goType = "int64"
	case protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		goType = "uint64"
	case protoreflect.FloatKind:
		goType = "float32"
	case protoreflect.DoubleKind:
		goType = "float64"
	case protoreflect.StringKind:
		goType = "string"
	case protoreflect.BytesKind:
		goType = "[]byte"
	case protoreflect.MessageKind, protoreflect.GroupKind:
		goType = "*" + qualifiedGoIdent(field.Message.GoIdent)
	}
	switch {
	case field.Desc.IsList():
		return "[]" + goType
	case field.Desc.IsMap():
		keyType := fieldGoType(field.Message.Fields[0])
		valType := fieldGoType(field.Message.Fields[1])
		return fmt.Sprintf("map[%v]%v", keyType, valType)
	}
	return goType
}

func fieldGoTypeClear(field *protogen.Field) string {
	goType := fieldGoType(field)

	return strings.TrimPrefix(strings.TrimPrefix(goType, "[]"), "*")
}

func isFieldOneOf(field *protogen.Field) bool {
	return field.Oneof != nil && strings.Index(field.GoName, "_") != 0
}
