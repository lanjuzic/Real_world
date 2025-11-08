package data

import (
	"context"
	//"database/sql"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"kratos-realworld/internal/conf"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(NewData, NewRealWorldRepo)

// Data .
type Data struct {
	DB  *gorm.DB
	RDB *redis.Client
	// TODO wrapped database client
}

// NewData .
func NewData(c *conf.Data, logger log.Logger) (*Data, func(), error) {
	logHelper := log.NewHelper(logger)

	db, dbCleanup, err := NewDB(c, logger)
	if err != nil {
		return nil, nil, err
	}
	rdb, rdbCleanup, err := NewRedis(c, logger)
	if err != nil {
		dbCleanup()
		return nil, nil, err
	}

	d := &Data{
		DB:  db,
		RDB: rdb,
	}

	cleanup := func() {
		logHelper.Info("closing the data resources")
		rdbCleanup()
		dbCleanup()
	}
	return d, cleanup, nil
}

// 初始化数据库
func NewDB(c *conf.Data, logger log.Logger) (*gorm.DB, func(), error) {
	logHelper := log.NewHelper(logger)

	var (
		db  *gorm.DB
		err error
	)

	switch c.Database.Driver {
	case "postgres":
		db, err = gorm.Open(postgres.Open(c.Database.Source), &gorm.Config{
			Logger: gormlogger.Default.LogMode(gormlogger.Warn),
		})
	// 如果要兼容 mysql，放开注释
	// case "mysql":
	// 	db, err = gorm.Open(mysql.Open(c.Database.Source), &gorm.Config{
	// 		Logger: gormlogger.Default.LogMode(gormlogger.Warn),
	// 	})
	default:
		return nil, nil, fmt.Errorf("unsupported db driver: %s", c.Database.Driver)

	}
	if err != nil {
		return nil, nil, err
	}
	//连接池配置
	sqlDB, err := db.DB()
	if err != nil {
		return nil, nil, err
	}
	sqlDB.SetMaxOpenConns(30)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxLifetime(60 * time.Minute)
	sqlDB.SetConnMaxIdleTime(10 * time.Minute)

	//启动时做一次快速探活
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if pingErr := sqlDB.PingContext(ctx); pingErr != nil {
		return nil, nil, pingErr
	}

	cleanup := func() {
		_ = sqlDB.Close()
	}

	logHelper.Info("database connected: driver= ", c.Database.Driver)

	return db, cleanup, nil

}

// 初始化redis
func NewRedis(c *conf.Data, logger log.Logger) (*redis.Client, func(), error) {
	logHelper := log.NewHelper(logger)

	opt := &redis.Options{
		Addr:         c.Redis.Addr,
		Password:     c.Redis.Password,
		DB:           0,
		ReadTimeout:  c.Redis.ReadTimeout.AsDuration(),
		WriteTimeout: c.Redis.WriteTimeout.AsDuration(),
	}
	rdb := redis.NewClient(opt)

	//探活
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, nil, err
	}

	cleanup := func() {
		_ = rdb.Close()
	}

	logHelper.Info("redis connected")

	return rdb, cleanup, nil
}
