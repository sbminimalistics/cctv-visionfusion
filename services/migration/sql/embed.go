package sql

import "embed"

//go:embed backend/*.sql
//go:embed dashboard/*.sql
var Files embed.FS
