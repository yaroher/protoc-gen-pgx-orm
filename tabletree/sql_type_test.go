package tabletree

import (
	"github.com/yaroher/protoc-gen-pgx-orm/protopgx"
	"testing"
)

func TestGetPgxTypeInfo(t *testing.T) {
	tests := []struct {
		name     string
		sqlType  protopgx.SqlFiledType
		nullable bool
		array    bool
		want     string
		wantErr  bool
	}{
		// Basic types
		{"TEXT", protopgx.SqlFiledType_TEXT, false, false, "string", false},
		{"INTEGER", protopgx.SqlFiledType_INTEGER, false, false, "int32", false},
		{"BIGINT", protopgx.SqlFiledType_BIGINT, false, false, "int64", false},
		{"SMALLINT", protopgx.SqlFiledType_SMALLINT, false, false, "int16", true}, // SMALLINT not implemented
		{"DOUBLE", protopgx.SqlFiledType_DOUBLE_PRECISION, false, false, "float64", false},
		{"REAL", protopgx.SqlFiledType_REAL, false, false, "float32", false},
		{"BOOLEAN", protopgx.SqlFiledType_BOOLEAN, false, false, "bool", false},
		{"TIMESTAMPTZ", protopgx.SqlFiledType_TIMESTAMPTZ, false, false, "time.Time", false},
		{"HSTORE", protopgx.SqlFiledType_HSTORE, false, false, "pgtype.Hstore", false},
		{"CHAR", protopgx.SqlFiledType_CHAR, false, false, "string", false},
		{"JSONB", protopgx.SqlFiledType_JSONB, false, false, "[]byte", false},

		// Nullable types
		{"nullable TEXT", protopgx.SqlFiledType_TEXT, true, false, "*string", false},
		{"nullable INTEGER", protopgx.SqlFiledType_INTEGER, true, false, "*int32", false},
		{"nullable TIMESTAMPTZ", protopgx.SqlFiledType_TIMESTAMPTZ, true, false, "*time.Time", false},

		// Array types
		{"array TEXT", protopgx.SqlFiledType_TEXT, false, true, "[]string", false},
		{"array INTEGER", protopgx.SqlFiledType_INTEGER, false, true, "[]int32", false},
		{"array TIMESTAMPTZ", protopgx.SqlFiledType_TIMESTAMPTZ, false, true, "[]time.Time", false},

		// Nullable array types
		{"nullable array TEXT", protopgx.SqlFiledType_TEXT, true, true, "[]string", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getPgxTypeInfo(tt.sqlType, tt.nullable, tt.array)
			if got != tt.want {
				t.Errorf("getPgxTypeInfo() = %v, want %v", got, tt.want)
			}
		})
	}
}
