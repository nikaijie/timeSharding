package util

import (
	"fmt"
	"reflect"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func getVars(db *gorm.DB) map[string]interface{} {
	stmt := db.Statement
	vars := make(map[string]interface{})
	if whereClause, ok := stmt.Clauses["WHERE"]; ok {
		expression := whereClause.Expression
		where := expression.(clause.Where)
		for _, expr := range where.Exprs {
			exprImpl, ok := expr.(clause.Expr)
			if !ok {
				continue
			}
			vars[exprImpl.SQL] = exprImpl.Vars
		}
	}
	return vars
}

func getShardingKey(db *gorm.DB) ([]string, []string) {
	model := db.Statement.Model
	modelType := reflect.TypeOf(model).Elem()
	fmt.Printf("查询的模型是: %T\n", modelType)
	var first, second []string
	for i := 0; i < modelType.NumField(); i++ {
		field := modelType.Field(i)
		timeShardingTag := field.Tag.Get("timeSharding")
		if timeShardingTag == "" {
			continue
		}
		parts := strings.Split(timeShardingTag, ":")
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		switch key {
		case "first":
			first = append(first, value)
		case "second":
			second = append(second, value)
		}
	}
	return first, second
}

func getTableName(first []string, second []string, vars map[string]interface{}, tableName string) string {
	return ""
}
