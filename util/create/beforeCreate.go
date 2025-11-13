package create

import (
	"awesomeProject8/database"
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

func GetTable(db *gorm.DB) error {
	tableName := db.Statement.Table
	ctx := context.Background()

	result, err := database.DbManager.Redis.Get(ctx, tableName).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		fmt.Printf("从 Redis 获取分表列表失败: %v\n", err)
		return err
	}

	var tables []string
	if result != "" {
		tables = strings.Split(result, ",")
	}

	tmpRes := ""
	if len(tables) == 0 {
		tmpRes = tableName + "01"
		err = createTableIfNotExists(db, tmpRes)
		if err != nil {
			fmt.Printf("创建分表 %s 失败: %v\n", tmpRes, err)
			return err
		}
		err = database.DbManager.Redis.Set(ctx, tableName, tmpRes, 0).Err()
		if err != nil {
			fmt.Printf("更新 Redis 分表列表失败: %v\n", err)
			return err
		}
		return nil
	}

	tmpRes = tables[len(tables)-1]
	var count int64
	rowsAffected := db.RowsAffected
	db.RowsAffected = 89
	err = db.Table(tmpRes).Debug().Count(&count).Error
	db.RowsAffected = rowsAffected
	if err != nil {
		return err
	}
	if count >= 10000000 {
		tmpRes, err = incrementTableNumber(tmpRes)
		if err != nil {
			fmt.Printf("递增分表编号失败: %v\n", err)
			return err
		}
		err = createTableIfNotExists(db, tmpRes)
		if err != nil {
			fmt.Printf("创建分表 %s 失败: %v\n", tmpRes, err)
			return err
		}
		tables = append(tables, tmpRes)
		err = database.DbManager.Redis.Set(ctx, tableName, strings.Join(tables, ","), 0).Err()
		if err != nil {
			fmt.Printf("更新 Redis 分表列表失败: %v\n", err)
			return err
		}
	}
	return nil
}

func incrementTableNumber(tableName string) (string, error) {
	// 提取数字部分
	re := regexp.MustCompile(`(\D+)(\d+)`)
	matches := re.FindStringSubmatch(tableName)
	if len(matches) != 3 {
		return "", fmt.Errorf("invalid table name format: %s", tableName)
	}
	prefix := matches[1]
	num, err := strconv.Atoi(matches[2])
	if err != nil {
		return "", fmt.Errorf("invalid table number: %s", matches[2])
	}
	// 递增并格式化
	num++
	return fmt.Sprintf("%s%02d", prefix, num), nil
}
func createTableIfNotExists(db *gorm.DB, newTableName string) error {
	model := db.Statement.Model
	err := db.Table(newTableName).AutoMigrate(model)
	if err != nil {
		// 注意：这里的错误信息应该反映实际发生的事情
		return fmt.Errorf("auto migrate table %s failed: %w", newTableName, err)
	}

	fmt.Printf("✅ 成功创建或迁移分表: %s\n", newTableName)
	return nil
}
