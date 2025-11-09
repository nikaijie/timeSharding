package util

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func getVars(db *gorm.DB) []string {
	stmt := db.Statement
	var vars []string
	if whereClause, ok := stmt.Clauses["WHERE"]; ok {
		expression := whereClause.Expression
		where := expression.(clause.Where)
		for _, expr := range where.Exprs {
			exprImpl, ok := expr.(clause.Expr)
			if !ok {
				continue
			}
			parts := strings.SplitN(exprImpl.SQL, " = ?", 2)
			vars = append(vars, parts[0])
		}

	}
	fmt.Println(vars)
	return vars
}
