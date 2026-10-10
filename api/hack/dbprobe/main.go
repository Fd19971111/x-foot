// Package main 临时数据库迁移程序（用完即删）
//
// 幂等地为 site_customer 补充「公司」与「身份证号」两列。
package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"
	"github.com/gogf/gf/v2/frame/g"
)

const sqlFile = "manifest/sql/customer_module.sql"

func main() {
	ctx := context.Background()
	db := g.DB()

	exists, err := columnExists(ctx, "site_customer", "company_name")
	if err != nil {
		fmt.Println("检查列失败:", err)
		os.Exit(1)
	}
	if exists {
		fmt.Println("company_name 已存在，跳过迁移")
		report(ctx)
		return
	}

	raw, err := os.ReadFile(sqlFile)
	if err != nil {
		fmt.Println("读取 SQL 文件失败:", err)
		os.Exit(1)
	}

	for _, stmt := range strings.Split(string(raw), ";") {
		lines := make([]string, 0, 8)
		for _, line := range strings.Split(stmt, "\n") {
			if strings.HasPrefix(strings.TrimSpace(line), "--") {
				continue
			}
			lines = append(lines, line)
		}

		sqlText := strings.TrimSpace(strings.Join(lines, "\n"))
		if sqlText == "" {
			continue
		}

		fmt.Println("执行:", strings.Join(strings.Fields(sqlText), " "))
		if _, err = db.Exec(ctx, sqlText); err != nil {
			fmt.Println("执行失败:", err)
			os.Exit(1)
		}
	}

	fmt.Println("迁移完成")
	report(ctx)
}

func report(ctx context.Context) {
	for _, col := range []string{"company_name", "id_card"} {
		ok, err := columnExists(ctx, "site_customer", col)
		fmt.Printf("  校验 site_customer.%s 存在: %v (err=%v)\n", col, ok, err)
	}
}

func columnExists(ctx context.Context, table, column string) (bool, error) {
	value, err := g.DB().GetValue(ctx,
		"SELECT COUNT(1) FROM information_schema.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ? AND COLUMN_NAME = ?",
		table, column)
	if err != nil {
		return false, err
	}

	return value.Int() > 0, nil
}
