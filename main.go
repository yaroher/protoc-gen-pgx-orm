package main

import (
	"bytes"
	"os"
	"path"
	"strings"

	"go.uber.org/zap"
	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/pluginpb"

	"github.com/yaroher/protoc-gen-pgx-orm/help"
	"github.com/yaroher/protoc-gen-pgx-orm/orm"
	"github.com/yaroher/protoc-gen-pgx-orm/tabletree"
)

const (
	SqlOutFolderParamName = "sql_file"
	OrmOutFolderParamName = "orm_folder"
	DefaultSqlFolder      = "./sql/models.sql"
	DefaultOrmFolder      = "./sql/orm"
)

var (
	SqlOutFile   = DefaultSqlFolder
	OrmOutFolder = DefaultOrmFolder
)

func initParams(p *protogen.Plugin) error {
	paramsMap := make(map[string]string)
	help.Logger.Warn(p.Request.GetParameter())
	params := strings.Split(p.Request.GetParameter(), ",")
	for _, param := range params {
		paramSplit := strings.Split(param, "=")
		paramsMap[paramSplit[0]] = paramSplit[1]
	}

	folder, ok := paramsMap[SqlOutFolderParamName]
	if ok {
		SqlOutFile = folder
	}
	ormFolder, ok := paramsMap[OrmOutFolderParamName]
	if ok {
		OrmOutFolder = ormFolder
	}
	return nil
}

func Generate(p *protogen.Plugin) error {
	e := initParams(p)
	if e != nil {
		return e
	}
	tables := tabletree.CollectTablesFromProto(p.Files)
	help.Logger.Info("nodes len", zap.Int("len", len(tables)))
	createSqls := make([]string, 0)
	for _, table := range tables {
		help.Logger.Info(string(table.Name), zap.String("sql_alias", table.SqlTableName()))
		help.Logger.Info("***")
		for _, field := range table.Fields {
			help.Logger.Info(field.ToSql(), zap.String(
				"sql_type",
				field.TypeInfo.SqlType.String(),
			), zap.String(
				"pgx_type",
				field.TypeInfo.PgxType,
			), zap.Bool(
				"nullable",
				field.TypeInfo.Nullable,
			), zap.Bool(
				"array",
				field.TypeInfo.IsArray,
			))
		}
		for _, cnst := range table.Constraints {
			help.Logger.Info("***")
			help.Logger.Info(
				"CONSTRAINT",
				zap.String("sql", cnst),
			)
		}
		for _, rel := range table.Relations {
			help.Logger.Info("***")
			help.Logger.Info(
				"RELATION",
				zap.String("from", table.SqlTableName()),
				zap.String("to", rel.To.SqlTableName()),
				zap.String("from_field", rel.ToField.SqlFieldName()),
				zap.String("to_field", rel.FromField.SqlFieldName()),
			)
		}
		for _, rel := range table.Backwards {
			help.Logger.Info("***")
			help.Logger.Info(
				"BACKWARD RELATION",
				zap.String("from", rel.From.SqlTableName()),
				zap.String("to", table.SqlTableName()),
				zap.String("from_field", rel.ToField.SqlFieldName()),
				zap.String("to_field", rel.FromField.SqlFieldName()),
			)
		}
		sql := table.ToSql()
		createSqls = append(createSqls, sql)
		help.Logger.Info("---------------------------------------------------------------------------------")
	}
	strBuff := bytes.NewBuffer(make([]byte, 0))
	for _, strs := range createSqls {
		strBuff.WriteString(strs)
	}
	err := os.MkdirAll(path.Dir(SqlOutFile), 0755)
	if err != nil {
		return err
	}
	err = os.WriteFile(
		SqlOutFile,
		strBuff.Bytes(),
		0644,
	)
	if err != nil {
		return err
	}
	err = os.MkdirAll(path.Dir(OrmOutFolder), 0755)
	if err != nil {
		return err
	}
	orm.GenerateOrm(p, tables, OrmOutFolder)

	return nil
}

func main() {
	protogen.Options{}.Run(func(plugin *protogen.Plugin) error {
		plugin.SupportedFeatures = uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL)
		return Generate(plugin)
	})
}
