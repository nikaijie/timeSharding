package create

import (
	"awesomeProject8/database"
	"awesomeProject8/util/query"
	"context"
	"fmt"
	"reflect"

	"gorm.io/gorm"
)

func CreateCache(db *gorm.DB) error {
	first, second := query.GetShardingKey(db)
	fmt.Println(first, second)
	if len(first) != 0 {
		err := createFirst(db, first)
		if err != nil {
			return err
		}
	}
	if len(second) != 0 {
		err := createSecond(db, second)
		if err != nil {
			return err
		}
	}
	return nil
}

func createFirst(db *gorm.DB, arr []string) error {
	stmt := db.Statement
	ctx := context.Background()
	nameMap := make(map[string]bool)
	for _, item := range arr {
		nameMap[item] = true
	}
	for _, field := range stmt.Schema.Fields {
		value, _ := field.ValueOf(stmt.Context, reflect.ValueOf(stmt.Dest))
		valueStr := fmt.Sprintf("%v", value)
		if nameMap[field.DBName] == false {
			continue
		}
		tableName := "first" + stmt.Table + field.DBName

		// 尝试从Redis获取map
		result, err := database.DbManager.Redis.HGetAll(ctx, tableName).Result()
		if err != nil {
			fmt.Printf("从 Redis 获取map失败: %v\n", err)
			return err
		}

		// 如果map为空，则创建新的map
		if len(result) == 0 {
			fmt.Printf("Redis中不存在key %s，创建新的map\n", tableName)
			newMap := make(map[string]string)
			newMap["start"] = valueStr
			newMap["end"] = valueStr
			err := database.DbManager.Redis.HMSet(ctx, tableName, newMap).Err()
			if err != nil {
				fmt.Printf("创建新map到Redis失败: %v\n", err)
				return err
			}
			fmt.Printf("成功创建新map到Redis，key: %s\n", tableName)
		} else {
			result["end"] = valueStr
			err := database.DbManager.Redis.HMSet(ctx, tableName, result).Err()
			if err != nil {
				fmt.Printf("创建新map到Redis失败: %v\n", err)
				return err
			}
		}

		fmt.Printf("字段名: %s, 数据库列名: %s, 值: %s\n", field.Name, field.DBName, valueStr)
	}
	return nil
}

func createSecond(db *gorm.DB, arr []string) error {
	stmt := db.Statement
	ctx := context.Background()
	nameMap := make(map[string]bool)
	for _, item := range arr {
		nameMap[item] = true
	}
	for _, field := range stmt.Schema.Fields {
		if nameMap[field.DBName] == false {
			continue
		}
		model := fmt.Sprintf("%v", stmt.Model)
		key := "second" + model + field.DBName

		// 获取Redis中的Set数据
		result, err := database.DbManager.Redis.SMembers(ctx, key).Result()
		if err != nil {
			fmt.Printf("从 Redis 获取Set失败: %v\n", err)
			return err
		}
		tableName := stmt.Table

		// 检查result中是否包含tableName
		found := false
		for _, item := range result {
			if item == tableName {
				found = true
				break
			}
		}

		// 如果没有包含tableName，则添加进去并保存到Redis
		if !found {
			err := database.DbManager.Redis.SAdd(ctx, key, tableName).Err()
			if err != nil {
				fmt.Printf("添加tableName到Redis Set失败: %v\n", err)
				return err
			}
			fmt.Printf("成功将 %s 添加到Redis Set，key: %s\n", tableName, key)
		} else {
			fmt.Printf("tableName %s 已存在于Redis Set中\n", tableName)
		}

		fmt.Printf("获取到Redis Set数据，key: %s, 内容: %v\n", key, result)
	}
	return nil
}
