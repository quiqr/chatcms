package db

import (
	"embed"
	"fmt"
	"io/fs"
)

//go:embed all:migrations
var migrationsRoot embed.FS

func Migrations() fs.FS {
	sub, err := fs.Sub(migrationsRoot, "migrations")
	if err != nil {
		panic(fmt.Sprintf("embedded migrations are unreadable: %v", err))
	}
	return sub
}
