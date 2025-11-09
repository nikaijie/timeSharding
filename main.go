package main

import (
	"awesomeProject8/util"
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// 定义全局上下文和 Redis 客户端变量
var (
	ctx         = context.Background()
	redisClient *redis.Client
)

type Order struct {
	OrderID   uint64    `gorm:"primaryKey" timeSharding:"first:order_id"` // 订单ID，一级索引
	UserID    uint64    `sharding:"second:user_id"`                       // 用户ID，二级索引
	Amount    float64   // 订单金额
	CreatedAt time.Time `sharding:"first:created_at"` // 创建时间，一级索引
}

func main() {
	// --- 步骤 1: 初始化 Redis 客户端 ---
	// 这里我们先初始化 Redis，虽然还不用它，但这是标准流程
	redisClient = redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "",
		DB:       0,
	})
	if err := redisClient.Ping(ctx).Err(); err != nil {
		panic(fmt.Sprintf("Redis 连接失败: %v", err))
	}
	fmt.Println("Redis 连接成功")

	dsn := "root:123456789@tcp(127.0.0.1:3306)/order_db?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		panic(fmt.Sprintf("数据库连接失败: %v", err))
	}
	fmt.Println("数据库连接成功")
	db.Use(util.NewShardingPlugin())

	err = db.AutoMigrate(&Order{})
	if err != nil {
		panic(fmt.Sprintf("数据表迁移失败: %v", err))
	}
	fmt.Println("数据表迁移成功")

	newOrder := Order{
		UserID:    1001,
		Amount:    99.99,
		CreatedAt: time.Now(),
	}
	result := db.Create(&newOrder)
	if result.Error != nil {
		panic(fmt.Sprintf("创建订单失败: %v", result.Error))
	}
	fmt.Printf("创建订单成功，订单ID: %d\n", newOrder.OrderID)

	var queryOrder Order
	// 注意：我们的自定义中间件也将在这里介入，拦截这个 Find 请求
	result = db.Where("order_id = ?", newOrder.OrderID).
		Where("amount = ?", 90).
		First(&queryOrder)
	if result.Error != nil {
		panic(fmt.Sprintf("查询订单失败: %v", result.Error))
	}
	fmt.Printf("查询订单成功: %+v\n", queryOrder)
}
