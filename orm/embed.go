package orm

import (
	"embed"
)

//go:embed *.go
//go:embed *.tmpl
var Contet embed.FS
