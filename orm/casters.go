package orm

import (
	"github.com/jackc/pgtype"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
	"time"
)

// ---------------------------------------------------------------------------------------------------------------------
// ---------------------------------------------------------------------------------------------------------------------
// ---------------------------------------------------------------------------------------------------------------------

func EnumToInt32[T protoreflect.Enum](v T) int32 {
	return int32(v.Number())
}
func EnumToSliceInt32[T protoreflect.Enum](v []T) []int32 {
	result := make([]int32, len(v))
	for i, el := range v {
		result[i] = EnumToInt32[T](el)
	}
	return result
}
func EnumFromInt32[T protoreflect.Enum](v int32) (ret T) {
	return ret.Type().New(protoreflect.EnumNumber(v)).(T)
}
func EnumFromSliceInt32[T protoreflect.Enum](v []int32) []T {
	result := make([]T, len(v))
	for i, el := range v {
		result[i] = EnumFromInt32[T](el)
	}
	return result
}

// ---------------------------------------------------------------------------------------------------------------------
// ---------------------------------------------------------------------------------------------------------------------
// ---------------------------------------------------------------------------------------------------------------------

func TimestampToTime(t *timestamppb.Timestamp) time.Time {
	return t.AsTime()
}
func TimestampToPtrTime(t *timestamppb.Timestamp) *time.Time {
	if t == nil {
		return nil
	}
	v := t.AsTime()
	return &v
}
func TimestampFromTime(t time.Time) *timestamppb.Timestamp {
	return timestamppb.New(t)
}
func TimestampFromPtrTime(t *time.Time) *timestamppb.Timestamp {
	if t == nil {
		return nil
	}
	return TimestampFromTime(*t)
}

func StringValueToString(v *wrapperspb.StringValue) string {
	if v == nil {
		return ""
	}
	return v.Value
}
func StringValueToPtrString(v *wrapperspb.StringValue) *string {
	if v == nil {
		return nil
	}
	return &v.Value
}
func StringValueFromString(v string) *wrapperspb.StringValue {
	return &wrapperspb.StringValue{Value: v}
}
func StringValueFromPtrString(v *string) *wrapperspb.StringValue {
	if v == nil {
		return nil
	}
	return StringValueFromString(*v)
}

func BoolValueToBool(v *wrapperspb.BoolValue) bool {
	if v == nil {
		return false
	}
	return v.Value
}
func BoolValueToPtrBool(v *wrapperspb.BoolValue) *bool {
	if v == nil {
		return nil
	}
	return &v.Value
}
func BoolValueFromBool(v bool) *wrapperspb.BoolValue {
	return &wrapperspb.BoolValue{Value: v}
}
func BoolValueFromPtrBool(v *bool) *wrapperspb.BoolValue {
	if v == nil {
		return nil
	}
	return BoolValueFromBool(*v)
}

func UInt32ValueToInt32(v *wrapperspb.UInt32Value) uint32 {
	if v == nil {
		return 0
	}
	return v.Value
}
func UInt32ValueToPtrInt32(v *wrapperspb.UInt32Value) *uint32 {
	if v == nil {
		return nil
	}
	return &v.Value
}
func UInt32ValueFromInt32(v uint32) *wrapperspb.UInt32Value {
	return &wrapperspb.UInt32Value{Value: v}
}
func UInt32ValueFromPtrInt32(v *uint32) *wrapperspb.UInt32Value {
	if v == nil {
		return nil
	}
	return UInt32ValueFromInt32(*v)
}

// ---------------------------------------------------------------------------------------------------------------------
// ---------------------------------------------------------------------------------------------------------------------
// ---------------------------------------------------------------------------------------------------------------------
func protoNew[T proto.Message]() (model T) {
	return model.ProtoReflect().Type().New().Interface().(T)
}
func MessageToJsonb[T proto.Message](v T) pgtype.JSONB {
	if !v.ProtoReflect().IsValid() {
		return pgtype.JSONB{Status: pgtype.Null}
	}
	return pgtype.JSONB{Bytes: MessageToSliceByte[T](v), Status: pgtype.Present}
}
func MessageToSliceByteSlice[T proto.Message](v []T) [][]byte {
	result := make([][]byte, len(v))
	for i, el := range v {
		result[i] = MessageToSliceByte[T](el)
	}
	return result
}
func MessageFromSliceByteSlice[T proto.Message](v [][]byte) []T {
	result := make([]T, len(v))
	for i, el := range v {
		result[i] = MessageFromSliceByte[T](el)
	}
	return result
}
func MessageToSliceByte[T proto.Message](v T) []byte {
	if !v.ProtoReflect().IsValid() {
		return nil
	}
	b, err := protojson.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}
func MessageFromSliceByte[T proto.Message](v []byte) T {
	ret := protoNew[T]()
	err := protojson.Unmarshal(v, ret)
	if err != nil {
		panic(err)
	}
	return ret
}
func MessageFromJsonb[T proto.Message](v pgtype.JSONB) T {
	ret := protoNew[T]()
	if v.Status != pgtype.Present {
		return ret
	}
	err := protojson.Unmarshal(v.Bytes, ret)
	if err != nil {
		panic(err)
	}
	return ret
}
