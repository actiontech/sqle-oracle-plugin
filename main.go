package main

import (
	"context"
	"flag"
	"fmt"
	"strings"

	"github.com/actiontech/sqle/sqle/driver"
	adaptor "github.com/actiontech/sqle/sqle/pkg/driver"
)

var version string
var printVersion = flag.Bool("version", false, "Print version & exit")

type OracleDialector struct {}

func (d *OracleDialector) Dialect(dsn *driver.DSN) (string, string) {
	if dsn.DatabaseName == "" {
		dsn.DatabaseName = "xe"
	}
	return "oracle", fmt.Sprintf("oracle://%s:%s@%s:%s/%s",
		dsn.User, dsn.Password, dsn.Host, dsn.Port, dsn.DatabaseName)
}

func (d *OracleDialector) String() string {
	return "Oracle"
}

func (d *OracleDialector) ShowDatabaseSQL() string {
	return "select global_name from global_name"
}

func main() {
	flag.Parse()

	if *printVersion {
		fmt.Println(version)
		return
	}

	plugin := adaptor.NewAdaptor(&OracleDialector{})

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
	plugin.AddRule(aviodSelectAllColumn, aviodSelectAllColumnHandler)
	plugin.Serve()
}
