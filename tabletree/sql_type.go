package tabletree

import (
	"fmt"
	"github.com/yaroher/protoc-gen-pgx-orm/help"
	"github.com/yaroher/protoc-gen-pgx-orm/protopgx"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
	"slices"
	"strings"
)

//goland:noinspection t
func getPgxTypeInfo(sqlType protopgx.SqlFiledType, nullable bool, array bool) string {
	retType := ""
	switch sqlType {
	case protopgx.SqlFiledType_TEXT:
		retType = "string"
	case protopgx.SqlFiledType_INTEGER:
		retType = "int32"
	case protopgx.SqlFiledType_BIGINT:
		retType = "int64"
	case protopgx.SqlFiledType_SMALLINT:
		retType = "int16"
	case protopgx.SqlFiledType_DOUBLE_PRECISION:
		retType = "float64"
	case protopgx.SqlFiledType_REAL:
		retType = "float32"
	case protopgx.SqlFiledType_BOOLEAN:
		retType = "bool"
	case protopgx.SqlFiledType_TIMESTAMPTZ:
		retType = "time.Time"
	case protopgx.SqlFiledType_HSTORE:
		retType = "pgtype.Hstore"
	case protopgx.SqlFiledType_CHAR:
		retType = "string"
	case protopgx.SqlFiledType_JSONB:
		if !nullable {
			if array {
				return "[][]byte"
			}
			return "[]byte"
		} else {
			if array {
				return "pgtype.JSONBArray"
			}
			return "pgtype.JSONB"
		}
	}
	if array {
		return "[]" + retType
	}
	if nullable {
		return "*" + retType
	}
	return retType
}

func isScalarType(sqlType protopgx.SqlFiledType) bool {
	if slices.Contains([]protopgx.SqlFiledType{
		protopgx.SqlFiledType_TEXT,
		protopgx.SqlFiledType_INTEGER,
		protopgx.SqlFiledType_BIGINT,
		protopgx.SqlFiledType_SMALLINT,
		protopgx.SqlFiledType_DOUBLE_PRECISION,
		protopgx.SqlFiledType_REAL,
		protopgx.SqlFiledType_BOOLEAN,
		protopgx.SqlFiledType_TIMESTAMPTZ,
		protopgx.SqlFiledType_CHAR,
	}, sqlType) {
		return true
	}
	return false
}

func isStringLikeType(sqlType protopgx.SqlFiledType) bool {
	if slices.Contains([]protopgx.SqlFiledType{
		protopgx.SqlFiledType_TEXT,
		protopgx.SqlFiledType_CHAR,
	}, sqlType) {
		return true
	}
	return false
}

func getSqlTypeFromProtoType(protoField *protogen.Field) protopgx.SqlFiledType {
	if protoField.Desc.IsMap() {
		return protopgx.SqlFiledType_HSTORE
	}
	switch protoField.Desc.Kind() {
	case protoreflect.BoolKind:
		return protopgx.SqlFiledType_BOOLEAN
	case protoreflect.EnumKind:
		return protopgx.SqlFiledType_INTEGER
	case protoreflect.Int32Kind:
		return protopgx.SqlFiledType_INTEGER
	case protoreflect.Sint32Kind:
		return protopgx.SqlFiledType_INTEGER
	case protoreflect.Uint32Kind:
		return protopgx.SqlFiledType_INTEGER
	case protoreflect.Int64Kind:
		return protopgx.SqlFiledType_BIGINT
	case protoreflect.Sint64Kind:
		return protopgx.SqlFiledType_BIGINT
	case protoreflect.Uint64Kind:
		return protopgx.SqlFiledType_BIGINT
	case protoreflect.Sfixed32Kind:
		return protopgx.SqlFiledType_INTEGER
	case protoreflect.Fixed32Kind:
		return protopgx.SqlFiledType_INTEGER
	case protoreflect.FloatKind:
		return protopgx.SqlFiledType_DOUBLE_PRECISION
	case protoreflect.Sfixed64Kind:
		return protopgx.SqlFiledType_DOUBLE_PRECISION
	case protoreflect.Fixed64Kind:
		return protopgx.SqlFiledType_DOUBLE_PRECISION
	case protoreflect.DoubleKind:
		return protopgx.SqlFiledType_DOUBLE_PRECISION
	case protoreflect.StringKind:
		return protopgx.SqlFiledType_TEXT
	case protoreflect.BytesKind:
		return protopgx.SqlFiledType_JSONB
	case protoreflect.MessageKind:
		if protoField.Message.Desc.FullName() == "google.protobuf.Timestamp" {
			return protopgx.SqlFiledType_TIMESTAMPTZ
		}
		if protoField.Message.Desc.FullName() == "google.protobuf.Duration" {
			panic(fmt.Sprintf("Unsupported type: %s", protoField.Message.Desc.FullName()))
		}
		if protoField.Message.Desc.FullName() == "google.protobuf.FloatValue" {
			return protopgx.SqlFiledType_DOUBLE_PRECISION
		}
		if protoField.Message.Desc.FullName() == "google.protobuf.Int32Value" {
			return protopgx.SqlFiledType_INTEGER
		}
		if protoField.Message.Desc.FullName() == "google.protobuf.Int64Value" {
			return protopgx.SqlFiledType_BIGINT
		}
		if protoField.Message.Desc.FullName() == "google.protobuf.UInt32Value" {
			return protopgx.SqlFiledType_INTEGER
		}
		if protoField.Message.Desc.FullName() == "google.protobuf.UInt64Value" {
			return protopgx.SqlFiledType_BIGINT
		}
		if protoField.Message.Desc.FullName() == "google.protobuf.StringValue" {
			return protopgx.SqlFiledType_TEXT
		}
		if protoField.Message.Desc.FullName() == "google.protobuf.BytesValue" {
			return protopgx.SqlFiledType_JSONB
		}
		if protoField.Message.Desc.FullName() == "google.protobuf.BoolValue" {
			return protopgx.SqlFiledType_BOOLEAN
		}
		if protoField.Message.Desc.FullName() == "google.protobuf.DoubleValue" {
			return protopgx.SqlFiledType_DOUBLE_PRECISION
		}
		return protopgx.SqlFiledType_JSONB
	default:
		panic(fmt.Sprintf("can't userDefinedCastType proto type %s", protoField.Desc.Kind()))
	}
}

func getProtoFieldSQLDefaultValue(field *protogen.Field, nullable bool, array bool) string {
	if array {
		if nullable {
			return ""
		}
		return "'{}'"
	}
	if nullable {
		return "null"
	}
	switch field.Desc.Kind() {
	case protoreflect.EnumKind:
		return "0"
	case protoreflect.Int32Kind:
		return "0"
	case protoreflect.Sint32Kind:
		return "0"
	case protoreflect.Uint32Kind:
		return "0"
	case protoreflect.Int64Kind:
		return "0"
	case protoreflect.Sint64Kind:
		return "0"
	case protoreflect.Uint64Kind:
		return "0"
	case protoreflect.Sfixed32Kind:
		return "0"
	case protoreflect.Fixed32Kind:
		return "0"
	case protoreflect.FloatKind:
		return "0"
	case protoreflect.Sfixed64Kind:
		return "0"
	case protoreflect.Fixed64Kind:
		return "0"
	case protoreflect.DoubleKind:
		return "0"
	case protoreflect.StringKind:
		return ""
	case protoreflect.BytesKind:
		return ""
	case protoreflect.BoolKind:
		return "FALSE"
	case protoreflect.MessageKind:
		if field.Message.Desc.FullName() == "google.protobuf.Timestamp" {
			return ""
		}
		if field.Message.Desc.FullName() == "google.protobuf.Duration" {
			return "0"
		}
		if field.Message.Desc.FullName() == "google.protobuf.FloatValue" {
			return "0"
		}
		if field.Message.Desc.FullName() == "google.protobuf.Int32Value" {
			return "0"
		}
		if field.Message.Desc.FullName() == "google.protobuf.Int64Value" {
			return "0"
		}
		if field.Message.Desc.FullName() == "google.protobuf.UInt32Value" {
			return "0"
		}
		if field.Message.Desc.FullName() == "google.protobuf.UInt64Value" {
			return "0"
		}
		if field.Message.Desc.FullName() == "google.protobuf.StringValue" {
			return ""
		}
		if field.Message.Desc.FullName() == "google.protobuf.BytesValue" {
			return ""
		}
		if field.Message.Desc.FullName() == "google.protobuf.BoolValue" {
			return "FALSE"
		}
		if field.Message.Desc.FullName() == "google.protobuf.DoubleValue" {
			return "0"
		}
		return ""
	default:
		panic(fmt.Sprintf("can't get default value for proto type %s", field.Desc.Kind()))
	}
}

func getFieldConstraint(info *protopgx.SqlConstraint, nullable bool) string {
	stringBuffer := &strings.Builder{}
	if info.GetPrimaryKey() {
		stringBuffer.WriteString(" PRIMARY KEY")
	}
	if info.GetUnique() {
		stringBuffer.WriteString(" UNIQUE")
	}
	if nullable {
		stringBuffer.WriteString(" NULL")
	} else {
		stringBuffer.WriteString(" NOT NULL")
	}
	if info.GetDefaultValue() != "" {
		stringBuffer.WriteString(" DEFAULT " + info.GetDefaultValue())
	}
	return help.StringOrDefault(info.GetConstraint(), stringBuffer.String())
}
