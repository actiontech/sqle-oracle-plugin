package main

import (
	"context"
	"flag"
	"fmt"
	"strings"

	"github.com/actiontech/sqle/sqle/driver"
	_ "github.com/sijms/go-ora/v2"
)

var version string
var printVersion = flag.Bool("version", false, "Print version & exit")

func main() {
	flag.Parse()

	if *printVersion {
		fmt.Println(version)
		return
	}

	aviodSelectAllColumn := &driver.Rule{
		Name:     "aviod_select_all_column",
		Desc:     "避免查询所有的列",
		Category: "DQL规范",
		Level:    driver.RuleLevelError,
	}
	aviodSelectAllColumnHandler := func(ctx context.Context, rule *driver.Rule, sql string) (string, error) {
		if strings.Contains(sql, "select *") {
			return rule.Desc, nil
		}
		return "", nil
	}
	AddRule(aviodSelectAllColumn, aviodSelectAllColumnHandler)

	driver.ServePlugin(&OracleDriverImpl{}, NewDriver)
}
