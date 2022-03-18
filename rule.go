package main

import (
	"context"

	"github.com/actiontech/sqle/sqle/driver"
)

type rawSQLRuleHandler func(ctx context.Context, rule *driver.Rule, rawSQL string) (string, error)

var rules            = []*driver.Rule{}
var ruleToRawHandler = map[string] /*rule name*/ rawSQLRuleHandler{}

func AddRule(r *driver.Rule, h rawSQLRuleHandler) {
	rules = append(rules, r)
	ruleToRawHandler[r.Name] = h
}