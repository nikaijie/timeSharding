package database

import (
	"context"
	"fmt"

	"github.com/go-redis/redis/v8"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DbManager *DatabaseManager

type DatabaseManager struct {
	MySQL *gorm.DB
	Redis *redis.Client
	Ctx   context.Context
}

func NewDatabaseManager(mysqlDSN, redisAddr, redisPassword string, redisDB int) (*DatabaseManager, error) {
	ctx := context.Background()

	mysqlDB, err := gorm.Open(mysql.Open(mysqlDSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("MySQL 连接失败: %v", err)
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       redisDB,
	})

	if err := redisClient.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("Redis 连接失败: %v", err)
	}

	return &DatabaseManager{
		MySQL: mysqlDB,
		Redis: redisClient,
		Ctx:   ctx,
	}, nil
}
