package create

import (
	"awesomeProject8/database"
	"awesomeProject8/util/query"
	"context"
	"encoding/json"
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
		tableName := "first_" + stmt.Table + "_" + field.DBName

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
		value, _ := field.ValueOf(stmt.Context, reflect.ValueOf(stmt.Dest))
		valueStr := fmt.Sprintf("%v", value)
		modelType := reflect.TypeOf(stmt.Model)
		if modelType.Kind() == reflect.Ptr {
			modelType = modelType.Elem()
		}
		modelName := modelType.Name()
		hashKey := "second_" + modelName + "_" + field.DBName
		tableName := stmt.Table

		// 获取Redis中的Hash数据
		result, err := database.DbManager.Redis.HGetAll(ctx, hashKey).Result()
		if err != nil {
			fmt.Printf("从 Redis 获取Hash失败: %v\n", err)
			return err
		}

		// 如果Hash不存在（result为空），创建新的Hash
		if len(result) == 0 {
			fmt.Printf("Redis中不存在Hash key %s，创建新的Hash\n", hashKey)
			result = make(map[string]string)
		}

		// 检查是否存在valueStr这个field
		if setJSON, exists := result[valueStr]; exists {
			// 解析JSON字符串为slice
			var tableSet []string
			if err := json.Unmarshal([]byte(setJSON), &tableSet); err != nil {
				fmt.Printf("解析JSON失败: %v\n", err)
				return err
			}

			// 检查tableName是否已在Set中
			found := false
			for _, table := range tableSet {
				if table == tableName {
					found = true
					break
				}
			}

			if !found {
				// 添加tableName到Set中
				tableSet = append(tableSet, tableName)
				newJSON, err := json.Marshal(tableSet)
				if err != nil {
					fmt.Printf("序列化JSON失败: %v\n", err)
					return err
				}

				// 更新Hash中的field
				err = database.DbManager.Redis.HSet(ctx, hashKey, valueStr, string(newJSON)).Err()
				if err != nil {
					fmt.Printf("更新Hash失败: %v\n", err)
					return err
				}
				fmt.Printf("成功将 %s 添加到valueStr %s 的Set中\n", tableName, valueStr)
			} else {
				fmt.Printf("tableName %s 已存在于valueStr %s 的Set中\n", tableName, valueStr)
			}
		} else {
			// 如果不存在valueStr这个field，创建新的Set
			tableSet := []string{tableName}
			setJSON, err := json.Marshal(tableSet)
			if err != nil {
				fmt.Printf("序列化JSON失败: %v\n", err)
				return err
			}

			// 设置Hash中的field
			err = database.DbManager.Redis.HSet(ctx, hashKey, valueStr, string(setJSON)).Err()
			if err != nil {
				fmt.Printf("设置Hash失败: %v\n", err)
				return err
			}
			fmt.Printf("创建新Set，valueStr: %s, 包含: %s\n", valueStr, tableName)
		}

		fmt.Printf("字段名: %s, 数据库列名: %s, 值: %s\n", field.Name, field.DBName, valueStr)
	}
	return nil
}
