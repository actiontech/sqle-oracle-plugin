package main

import (
	"context"
	"database/sql"
	_driver "database/sql/driver"
	"fmt"
	"os"
	"strings"

	"github.com/actiontech/sqle/sqle/driver"
	"github.com/actiontech/sqle/sqle/pkg/params"
	"github.com/hashicorp/go-hclog"
	"github.com/percona/go-mysql/query"
	"github.com/pkg/errors"
	"vitess.io/vitess/go/vt/sqlparser"
)

const serverNameKey = "service_name"

func NewDriver(cfg *driver.Config) driver.Driver {
	di := &OracleDriverImpl{
		cfg: cfg,
		l: hclog.New(&hclog.LoggerOptions{
			Level:      hclog.Trace,
			Output:     os.Stderr,
			JSONFormat: true,
		}),
	}
	return di
}

type OracleDriverImpl struct {
	l   hclog.Logger
	cfg *driver.Config

	isConnected bool
	db          *sql.DB
	conn        *sql.Conn
}

func (o *OracleDriverImpl) Name() string {
	return "Oracle"
}

func (o *OracleDriverImpl) Rules() []*driver.Rule {
	return rules
}

func (o *OracleDriverImpl) AdditionalParams() params.Params {
	return params.Params{
		&params.Param{
			Key:   serverNameKey,
			Value: "XE",
			Desc:  "service name",
			Type:  params.ParamTypeString,
		},
	}
}

func (o *OracleDriverImpl) getConn() (*sql.Conn, error) {
	if o.isConnected {
		return o.conn, nil
	}
	var (
		db   *sql.DB
		conn *sql.Conn
		err  error
	)
	dsn := o.cfg.DSN
	if dsn == nil {
		return nil, fmt.Errorf("open database failed, the rerquest dsn is invalid")
	}
	serviceName := dsn.AdditionalParams.GetParam(serverNameKey).String()
	if serviceName == "" {
		serviceName = "XE"
	}
	db, err = sql.Open("oracle", fmt.Sprintf("oracle://%s:%s@%s:%s/%s",
		dsn.User, dsn.Password, dsn.Host, dsn.Port, serviceName))
	if err != nil {
		return nil, errors.Wrap(err, "open database failed")
	}
	defer func() {
		if err != nil {
			conn.Close()
			db.Close()
		}
	}()
	conn, err = db.Conn(context.TODO())
	if err != nil {
		return nil, errors.Wrap(err, "get database connection failed")
	}
	err = conn.PingContext(context.TODO())
	if err != nil {
		return nil, errors.Wrap(err, "ping database connection failed")
	}
	if dsn.DatabaseName != "" {
		_, err = conn.ExecContext(context.TODO(), fmt.Sprintf("ALTER SESSION SET CURRENT_SCHEMA = %s", dsn.DatabaseName))
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("switch to schema %s failed", dsn.DatabaseName))
		}
	}
	o.db = db
	o.conn = conn
	o.isConnected = true
	return o.conn, nil
}

func (o *OracleDriverImpl) Close(ctx context.Context) {
	if !o.isConnected {
		return
	}
	if err := o.conn.Close(); err != nil {
		o.l.Error("failed to close connection in driver adaptor", "err", err)
	}
	if err := o.db.Close(); err != nil {
		o.l.Error("failed to close database in driver adaptor", "err", err)
	}
}

func (o *OracleDriverImpl) Ping(ctx context.Context) error {
	conn, err := o.getConn()
	if err != nil {
		return err
	}
	if err := conn.PingContext(ctx); err != nil {
		return errors.Wrap(err, "ping in driver adaptor")
	}
	return nil
}

func (o *OracleDriverImpl) Exec(ctx context.Context, sql string) (_driver.Result, error) {
	conn, err := o.getConn()
	if err != nil {
		return nil, err
	}
	res, err := conn.ExecContext(ctx, sql)
	if err != nil {
		return nil, errors.Wrap(err, "exec sql in driver adaptor")
	}
	return res, nil
}

func (o *OracleDriverImpl) Tx(ctx context.Context, sqls ...string) ([]_driver.Result, error) {
	var (
		err error
		tx  *sql.Tx
	)
	conn, err := o.getConn()
	if err != nil {
		return nil, err
	}
	tx, err = conn.BeginTx(ctx, nil)
	if err != nil {
		return nil, errors.Wrap(err, "begin tx in driver adaptor")
	}

	defer func() {
		if err != nil {
			if err = tx.Rollback(); err != nil {
				err = errors.Wrap(err, "rollback tx in driver adaptor")
				return
			}
		} else {
			if err = tx.Commit(); err != nil {
				err = errors.Wrap(err, "commit tx in driver adaptor")
				return
			}
		}
	}()

	results := make([]_driver.Result, 0, len(sqls))
	for _, sql := range sqls {
		result, e := tx.ExecContext(ctx, sql)
		if e != nil {
			err = errors.Wrap(e, "exec sql in driver adaptor")
			return nil, err
		}
		results = append(results, result)
	}

	return results, nil
}

func (o *OracleDriverImpl) Schemas(ctx context.Context) ([]string, error) {
	conn, err := o.getConn()
	if err != nil {
		return nil, err
	}
	query := `
SELECT username FROM all_users WHERE username NOT IN(
'AUD_SYS','ANONYMOUS','APEX_030200','APEX_040000', 'APEX_PUBLIC_USER', 
'APPQOSSYS', 'BI USERS', 'CTXSYS', 'DBSNMP','DIP USERS', 'EXFSYS', 
'FLOWS_FILES', 'HR USERS', 'IX USERS', 'MDDATA', 'MDSYS', 'MGMT_VIEW', 
'OE USERS','OLAPSYS', 'ORACLE_OCM', 'ORDDATA', 'ORDPLUGINS', 'ORDSYS', 
'OUTLN', 'OWBSYS', 'OWBSYS_AUDIT', 'PM USERS','SCOTT', 'SH USERS', 
'SI_INFORMTN_SCHEMA', 'SPATIAL_CSW_ADMIN_USR', 'SPATIAL_WFS_ADMIN_USR', 'SYS',
'SYSMAN', 'SYSTEM', 'WMSYS', 'XDB', 'XS$NULL', 'DIP', 'OJVMSYS', 'LBACSYS') 
ORDER BY username`
	rows, err := conn.QueryContext(ctx, query)
	if err != nil {
		return nil, errors.Wrap(err, "query schema in driver adaptor")
	}
	defer rows.Close()

	var schemas []string
	for rows.Next() {
		var schema string
		if err := rows.Scan(&schema); err != nil {
			return nil, errors.Wrap(err, "scan schema in driver adaptor")
		}
		schemas = append(schemas, schema)
	}

	if rows.Err() != nil {
		return nil, errors.Wrap(rows.Err(), "scan schema in driver adaptor")
	}

	return schemas, nil
}

func (o *OracleDriverImpl) Parse(ctx context.Context, sql string) ([]driver.Node, error) {
	sqls, err := sqlparser.SplitStatementToPieces(sql)
	if err != nil {
		return nil, errors.Wrap(err, "split sql")
	}
	if err != nil {
		return nil, errors.Wrapf(err, "split sql %s error", sql)
	}

	nodes := make([]driver.Node, 0, len(sqls))
	for _, sql := range sqls {
		n := driver.Node{
			Text:        sql,
			Type:        classifySQL(sql),
			Fingerprint: query.Fingerprint(sql),
		}
		nodes = append(nodes, n)
	}
	return nodes, nil
}

func classifySQL(sql string) (sqlType string) {
	lowerSQL := strings.TrimSpace(strings.ToLower(sql))
	if strings.HasPrefix(lowerSQL, "update") ||
		strings.HasPrefix(lowerSQL, "insert") ||
		strings.HasPrefix(lowerSQL, "delete") {
		return driver.SQLTypeDML
	}
	return driver.SQLTypeDDL
}

func (o *OracleDriverImpl) Audit(ctx context.Context, sql string) (*driver.AuditResult, error) {
	result := driver.NewInspectResults()
	for _, rule := range o.cfg.Rules {
		handler, ok := ruleToRawHandler[rule.Name]
		if ok {
			msg, err := handler(ctx, rule, sql)
			if err != nil {
				return nil, errors.Wrapf(err, "audit SQL %s in driver adaptor", sql)
			}
			result.Add(rule.Level, msg)
		}
	}
	return result, nil
}

func (o *OracleDriverImpl) GenRollbackSQL(ctx context.Context, sql string) (string, string, error) {
	return "", "", nil
}
