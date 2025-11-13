package util

import (
	"fmt"

	"gorm.io/gorm"
)

type ShardingPlugin struct {
}

func NewShardingPlugin() *ShardingPlugin {
	return &ShardingPlugin{}
}

func (p *ShardingPlugin) Name() string {
	return "sharding_plugin"
}

func (p *ShardingPlugin) Initialize(db *gorm.DB) error {

	db.Callback().Create().Before("gorm:create").Register("sharding:before_create", p.BeforeCreate)
	db.Callback().Create().After("gorm:create").Register("sharding:after_create", p.AfterCreate)

	db.Callback().Update().Before("gorm:update").Register("sharding:before_update", p.BeforeUpdate)
	db.Callback().Update().After("gorm:update").Register("sharding:after_update", p.AfterUpdate)

	db.Callback().Delete().Before("gorm:delete").Register("sharding:before_delete", p.BeforeDelete)
	db.Callback().Delete().After("gorm:delete").Register("sharding:after_delete", p.AfterDelete)

	db.Callback().Query().Before("gorm:query").Register("sharding:before_query", p.BeforeQuery)

	fmt.Println("分库分表中间件初始化完成。")
	return nil
}

// BeforeQuery 在查询记录之前执行
func (p *ShardingPlugin) BeforeQuery(db *gorm.DB) {
	fmt.Println("before query")
	vars := getVars(db)
	first, second := getShardingKey(db)
	tableName := db.Statement.Table
	tableName = getTableName(first, second, vars, tableName)

}

func (p *ShardingPlugin) BeforeCreate(db *gorm.DB) {
	getTables(db)
}

func (p *ShardingPlugin) BeforeUpdate(db *gorm.DB) {

}

func (p *ShardingPlugin) BeforeDelete(db *gorm.DB) {

}

func (p *ShardingPlugin) AfterCreate(db *gorm.DB) {

}

func (p *ShardingPlugin) AfterUpdate(db *gorm.DB) {

}

func (p *ShardingPlugin) AfterDelete(db *gorm.DB) {

}
