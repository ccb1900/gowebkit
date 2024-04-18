package db

import (
	"embed"

	"github.com/ccb1900/gocommon/logger"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"

	// _ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

// 根据配置适配不同的数据库类型

func Migrate(name string, path embed.FS) {
	mydb := Default()

	dd, _ := mydb.DB()

	driver, err := sqlite.WithInstance(dd, &sqlite.Config{})
	if err != nil {
		panic(err)
	}
	d, _ := iofs.New(path, "migrations")
	m, err := migrate.NewWithInstance(
		"iofs",
		d,
		// "file://./"+path,
		name,
		driver)
	if err != nil {
		logger.Default().Error("migrate error", "err", err)
		panic(err)
	}
	if err := m.Up(); err != nil {
		if err.Error() == "no change" {
			return
		}
		panic(err)
	}
}

func Rollback(name string, path embed.FS) {
	mydb := Default()

	dd, _ := mydb.DB()
	driver, err := sqlite.WithInstance(dd, &sqlite.Config{})
	if err != nil {
		panic(err)
	}
	d, _ := iofs.New(path, "migrations")
	m, err := migrate.NewWithInstance(
		"iofs",
		d,
		name, driver)
	if err != nil {
		logger.Default().Error("migrate error", "err", err)
		panic(err)
	}
	if err := m.Down(); err != nil {
		if err.Error() == "no change" {
			return
		}
		panic(err)
	}
}
