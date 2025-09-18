package tabletree

import (
	"fmt"
	"github.com/yaroher/protoc-gen-pgx-orm/protopgx"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
	"slices"
	"strings"
)

func NewTypeInfo(
	field *protogen.Field,
	opts *protopgx.SqlType,
	nullable bool,
	array bool,
) *protopgx.ParsedField_TypeInfo {
	pgxType := getPgxTypeInfo(opts.GetType(), nullable, array)
	if field == nil {
		return &protopgx.ParsedField_TypeInfo{
			SqlType:               opts,
			PgxType:               pgxType,
			Nullable:              nullable,
			IsArray:               array,
			ProtoKind:             protopgx.ParsedField_KIND_UNSPECIFIED,
			ForceUserDefineCaster: opts.UserCast,
			OverrideSqlName:       opts.Name,
		}
	}
	parsed := &protopgx.ParsedField_TypeInfo{
		SqlType:               opts,
		PgxType:               pgxType,
		Nullable:              nullable,
		IsArray:               array,
		ProtoKind:             protopgx.ParsedField_ProtoKind(field.Desc.Kind()),
		ForceUserDefineCaster: opts.UserCast,
		OverrideSqlName:       opts.Name,
	}
	parsed.DownCasterFn = getDowncast(field, parsed)
	parsed.UpCasterFn = getUpcast(field, parsed)
	return parsed
}

//goland:noinspection t
func getDowncast(field *protogen.Field, info *protopgx.ParsedField_TypeInfo) *protopgx.CasterFn {
	if strings.HasSuffix(field.Desc.Kind().String(), info.PgxType) && field.Desc.Kind().String()[0] == 'u' {
		return &protopgx.CasterFn{
			CallSignature: fmt.Sprintf("%s($var)", info.PgxType),
		}
	}
	if info.ForceUserDefineCaster {
		return &protopgx.CasterFn{
			Type:          userDefinedCastType(fieldGoType(field), info.PgxType),
			Name:          userDefinedCastName(downcastPrefix, field.GoName),
			CallSignature: plainSignature,
			UserDefined:   true,
		}
	}
	if isKnownType(field) {
		return knownTypeCaster(field, To, info)
	}
	if field.Desc.Kind() == protoreflect.MessageKind {
		return &protopgx.CasterFn{
			Name:          casterName("Message", To, info),
			CallSignature: genericSignature(fieldGoTypeClear(field), true),
		}
	}
	return &protopgx.CasterFn{
		CallSignature: noneSignature,
	}
}

//goland:noinspection t
func getUpcast(field *protogen.Field, info *protopgx.ParsedField_TypeInfo) *protopgx.CasterFn {
	if strings.HasSuffix(field.Desc.Kind().String(), info.PgxType) && field.Desc.Kind().String()[0] == 'u' {
		return &protopgx.CasterFn{
			CallSignature: fmt.Sprintf("u%s($var)", info.PgxType),
		}
	}
	if info.ForceUserDefineCaster {
		return &protopgx.CasterFn{
			Type:          userDefinedCastType(info.PgxType, fieldGoType(field)),
			Name:          userDefinedCastName(upcastPrefix, field.GoName),
			CallSignature: plainSignature,
			UserDefined:   true,
		}
	}
	if isKnownType(field) {
		return knownTypeCaster(field, From, info)
	}
	if field.Desc.Kind() == protoreflect.MessageKind {
		return &protopgx.CasterFn{
			Name:          casterName("Message", From, info),
			CallSignature: genericSignature(fieldGoTypeClear(field), true),
		}
	}
	return &protopgx.CasterFn{
		CallSignature: noneSignature,
	}
}

func knownTypeCaster(field *protogen.Field, dest castDest, info *protopgx.ParsedField_TypeInfo) *protopgx.CasterFn {
	if field.Desc.Kind() == protoreflect.MessageKind {
		switch field.Message.Desc.FullName() {
		case "google.protobuf.Timestamp":
			return &protopgx.CasterFn{
				Name:          casterName("Timestamp", dest, info),
				CallSignature: plainSignature,
			}
		case "google.protobuf.FloatValue":
			return &protopgx.CasterFn{
				Name:          casterName("FloatValue", dest, info),
				CallSignature: plainSignature,
			}
		case "google.protobuf.Int32Value":
			return &protopgx.CasterFn{
				Name:          casterName("Int32Value", dest, info),
				CallSignature: plainSignature,
			}
		case "google.protobuf.Int64Value":
			return &protopgx.CasterFn{
				Name:          casterName("Int64Value", dest, info),
				CallSignature: plainSignature,
			}
		case "google.protobuf.UInt32Value":
			return &protopgx.CasterFn{
				Name:          casterName("UInt32Value", dest, info),
				CallSignature: plainSignature,
			}
		case "google.protobuf.UInt64Value":
			return &protopgx.CasterFn{
				Name:          casterName("UInt64Value", dest, info),
				CallSignature: plainSignature,
			}
		case "google.protobuf.StringValue":
			return &protopgx.CasterFn{
				Name:          casterName("StringValue", dest, info),
				CallSignature: plainSignature,
			}
		case "google.protobuf.BytesValue":
			return &protopgx.CasterFn{
				Name:          casterName("BytesValue", dest, info),
				CallSignature: plainSignature,
			}
		case "google.protobuf.BoolValue":
			return &protopgx.CasterFn{
				Name:          casterName("BoolValue", dest, info),
				CallSignature: plainSignature,
			}
		case "google.protobuf.DoubleValue":
			return &protopgx.CasterFn{
				Name:          casterName("DoubleValue", dest, info),
				CallSignature: plainSignature,
			}
		}
	}
	if field.Desc.Kind() == protoreflect.EnumKind {
		if info.IsArray {
			return &protopgx.CasterFn{
				Name:          casterName("Enum", dest, info),
				CallSignature: genericSignature(fieldGoTypeClear(field), false),
			}
		}
		return &protopgx.CasterFn{
			Name:          casterName("Enum", dest, info),
			CallSignature: genericSignature(fieldGoTypeClear(field), false),
		}
	}
	panic("unknown type in knownTypeCaster")
}

func isKnownType(field *protogen.Field) bool {
	if field.Desc.Kind() == protoreflect.MessageKind {
		return slices.Contains([]protoreflect.FullName{
			"google.protobuf.Timestamp",
			"google.protobuf.FloatValue",
			"google.protobuf.Int32Value",
			"google.protobuf.Int64Value",
			"google.protobuf.UInt32Value",
			"google.protobuf.UInt64Value",
			"google.protobuf.StringValue",
			"google.protobuf.BytesValue",
			"google.protobuf.BoolValue",
			"google.protobuf.DoubleValue",
		}, field.Message.Desc.FullName())
	}
	if field.Desc.Kind() == protoreflect.EnumKind {
		return true
	}
	return false
}
