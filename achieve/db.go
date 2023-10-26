package achieve

import (
	"errors"

	_ "github.com/go-sql-driver/mysql"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

// DriverName 数据库驱动名称
type DriverName string

const (
	Postgresql DriverName = "postgresql"
	MySQL                 = "mysql"
	SQLite                = "sqlite"
	SQLServer             = "sqlserver"
	TiDB       DriverName = "tidb"
)

// DBConnect init database connect
func DBConnect(driver DriverName, dsn string, option options) (*gorm.DB, error) {
	var (
		gormConfig = &gorm.Config{}
		db         *gorm.DB
		err        error
	)

	if option.logger != nil {
		gormConfig.Logger = option.logger
	}
	if option.tablePrefix != "" {
		gormConfig.NamingStrategy = schema.NamingStrategy{TablePrefix: option.tablePrefix}
	}
	//
	switch driver {
	case Postgresql:
		db, err = gorm.Open(postgres.Open(dsn), gormConfig)
		if err != nil {
			return nil, err
		}
	case MySQL:
		db, err = gorm.Open(mysql.New(mysql.Config{
			DriverName: "go-sql-driver",
			DSN:        dsn,
		}), gormConfig)
		if err != nil {
			return nil, err
		}
	case SQLite:
		db, err = gorm.Open(sqlite.Open(dsn), gormConfig)
		if err != nil {
			return nil, err
		}
	}

	if db == nil {
		err = errors.New("database connect failed")
		return nil, err
	}
	// true 自动迁移
	if option.autoMigrate && len(option.autoMigrateDst) > 0 {
		if err = db.AutoMigrate(option.autoMigrateDst...); err != nil {
			return nil, err
		}
	}
	// 使用链路追踪
	if option.opentracingPlugin != nil {
		if err = db.Use(option.opentracingPlugin); err != nil {
			return nil, err
		}
	}
	return db, nil
}
