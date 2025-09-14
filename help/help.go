package help

import (
	"os"
	"strings"

	"github.com/iancoleman/strcase"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type customLogger struct {
}

func (c customLogger) Write(p []byte) (n int, err error) {
	return os.Stderr.WriteString(string(p))
}

func (c customLogger) Sync() error {
	return nil
}

var Logger = zap.New(zapcore.NewCore(zapcore.NewConsoleEncoder(
	zapcore.EncoderConfig{
		MessageKey:     "msg",
		LevelKey:       "level",
		NameKey:        "logger",
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
	}), &customLogger{}, zapcore.InfoLevel)).Named("protoc-gen-pgx-orm")

func getMessageByFullName(p *protogen.Plugin, fullName string) *protogen.Message {
	for _, file := range p.Files {
		for _, message := range file.Messages {
			if string(message.Desc.FullName()) == fullName {
				return message
			}
		}
	}
	return nil
}

func lowerSnake(s protoreflect.Name) string {
	return strcase.ToSnake(strings.ToLower(string(s)))
}

func StringOrDefault(s string, d string) string {
	if s != "" {
		return s
	}
	return d
}

func ListStringOrDefault(s []string, d []string) []string {
	if len(s) > 0 {
		return s
	}
	return d
}
