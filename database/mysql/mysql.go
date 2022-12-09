package mysql

import (
	"database/sql"
	_ "embed"
	"fmt"
	"github.com/dityaaa/concept/database"
	"github.com/go-sql-driver/mysql"
	"io"
	nurl "net/url"
	"strings"
	"time"
)

var _ database.Driver = (*MySQL)(nil)

//go:embed shistory.sql
var sHistoryScript string

//go:embed slocking.sql
var sLockingScript string

type Config struct {
	HistoryTable string
	LockingTable string
}

type MySQL struct {
	db *sql.DB

	historyTable string
	lockingTable string

	booted    bool
	tUsername string
	rows      *sql.Rows
}

func Open(url string) (database.Driver, error) {
	purl, err := nurl.Parse(url)
	if err != nil {
		return nil, err
	}

	if purl.Hostname() == "" {
		purl.Host = "127.0.0.1" + purl.Host
	}

	if purl.Port() == "" {
		purl.Host = purl.Host + ":3306"
	}

	mysqlCfg := mysql.NewConfig()
	mysqlCfg.Net = "tcp"
	mysqlCfg.Addr = purl.Host
	mysqlCfg.User = purl.User.Username()
	mysqlCfg.Passwd, _ = purl.User.Password()
	mysqlCfg.DBName = strings.TrimPrefix(purl.Path, "/")
	mysqlCfg.Timeout = 10 * time.Second
	mysqlCfg.Params = map[string]string{
		"multiStatements": "true",
	}

	db, err := sql.Open("mysql", mysqlCfg.FormatDSN())
	if err != nil {
		return nil, err
	}

	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	return WithInstance(db, Config{
		HistoryTable: purl.Query().Get("x-history-table"),
		LockingTable: purl.Query().Get("x-locking-table"),
	})
}

func WithInstance(inst *sql.DB, cfg Config) (database.Driver, error) {
	if cfg.HistoryTable == "" {
		cfg.HistoryTable = "migration_history"
	}

	return &MySQL{
		db:           inst,
		historyTable: cfg.HistoryTable,
		lockingTable: cfg.LockingTable,
		booted:       true,
	}, nil
}

func (i *MySQL) Name() string {
	return "mysql"
}

func (i *MySQL) Close() error {
	return i.db.Close()
}

func (i *MySQL) Read() ([]*database.History, error) {
	if _, err := i.historyTableExists(true); err != nil {
		return nil, err
	}

	query := fmt.Sprintf(
		"SELECT * FROM `%s` AS `h1` WHERE `h1`.`rank` = (SELECT MAX(`h2`.`rank`) FROM `%s` AS `h2` WHERE `h2`.`version` = `h1`.`version` AND `h2`.`mode` = `h1`.`mode`)",
		i.historyTable,
		i.historyTable,
	)
	rows, err := i.db.Query(query)
	if err != nil {
		return nil, err
	}

	res := make([]*database.History, 0)

	for rows.Next() {
		var row database.History

		err := rows.Scan(
			&row.Rank,
			&row.Mode,
			&row.Version,
			&row.ScriptName,
			&row.Description,
			&row.Checksum,
			&row.AppliedBy,
			&row.AppliedAt,
			&row.ExecutionTime,
			&row.Success,
		)
		if err != nil {
			return nil, err
		}

		res = append(res, &row)
	}

	return res, nil
}

func (i *MySQL) Write(history *database.History) error {
	if history.AppliedBy == "" {
		history.AppliedBy = i.username()
	}

	var res sql.Result
	var insertedRank any = nil
	query := fmt.Sprintf("INSERT INTO `%s` VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", i.historyTable)
	if history.Rank > 0 {
		query = fmt.Sprintf("REPLACE INTO `%s` VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", i.historyTable)
		insertedRank = int64(history.Rank)
	}

	res, err := i.db.Exec(
		query,
		insertedRank,
		history.Mode,
		history.Version,
		history.ScriptName,
		history.Description,
		history.Checksum,
		history.AppliedBy,
		history.AppliedAt,
		history.ExecutionTime,
		history.Success,
	)
	if err != nil {
		return err
	}

	if insertedRank == nil {
		insertedRank, err = res.LastInsertId()
		if err != nil {
			return err
		}
	}

	castRank, ok := insertedRank.(int64)
	if !ok {
		panic("failed to assert insertedRank type")
	}
	history.Rank = uint64(castRank)
	return nil
}

func (i *MySQL) Run(migration io.Reader) error {
	mg, err := io.ReadAll(migration)
	if err != nil {
		return err
	}

	query := string(mg)

	_, err = i.db.Exec(query)
	return err
}

func (i *MySQL) Purge() []error {
	errorItems := make([]error, 0)

	if err := i.purgeTables(); err != nil {
		errorItems = append(errorItems, err)
		fmt.Println(err)
	}

	if err := i.purgeStoredProcedures(); err != nil {
		errorItems = append(errorItems, err)
		fmt.Println(err)
	}

	return errorItems
}

func (i *MySQL) purgeTables() error {
	query := "SELECT TABLE_NAME, TABLE_TYPE FROM information_schema.TABLES WHERE TABLE_SCHEMA = DATABASE()"
	rows, err := i.db.Query(query)
	if err != nil {
		return err
	}

	quoteName := func(name string) string {
		return "`" + name + "`"
	}

	tables := make([]string, 0)
	views := make([]string, 0)

	for rows.Next() {
		var tableName string
		var tableType string

		if err := rows.Scan(&tableName, &tableType); err != nil {
			return err
		}

		if tableType == "BASE TABLE" {
			tables = append(tables, quoteName(tableName))
		}

		if tableType == "VIEW" {
			views = append(views, quoteName(tableName))
		}
	}

	if _, err = i.db.Exec("SET foreign_key_checks = 0"); err != nil {
		return err
	}

	if len(tables) > 0 {
		query = "DROP TABLE IF EXISTS " + strings.Join(tables, ", ")
		if _, err = i.db.Exec(query); err != nil {
			return err
		}
	}

	if len(tables) > 0 {
		query = "DROP VIEW IF EXISTS " + strings.Join(views, ", ")
		if _, err = i.db.Exec(query); err != nil {
			return err
		}
	}

	if _, err = i.db.Exec("SET foreign_key_checks = 1"); err != nil {
		return err
	}

	return nil
}

func (i *MySQL) purgeStoredProcedures() error {
	query := "SELECT ROUTINE_NAME FROM information_schema.ROUTINES WHERE ROUTINE_SCHEMA = DATABASE() AND ROUTINE_TYPE = 'PROCEDURE'"
	rows, err := i.db.Query(query)
	if err != nil {
		return err
	}

	quoteName := func(name string) string {
		return "`" + name + "`"
	}

	items := make([]string, 0)

	for rows.Next() {
		var name string

		if err := rows.Scan(&name); err != nil {
			return err
		}

		items = append(items, quoteName(name))
	}

	for _, procedure := range items {
		query = "DROP PROCEDURE IF EXISTS " + quoteName(procedure)
		if _, err = i.db.Exec(query); err != nil {
			return err
		}
	}

	return nil
}

func (i *MySQL) historyTableExists(create bool) (bool, error) {
	if create {
		return i.tableExists(i.historyTable, fmt.Sprintf(sHistoryScript, i.historyTable))
	}

	return i.tableExists(i.historyTable, "")
}

func (i *MySQL) lockingTableExists(create bool) (bool, error) {
	if create && i.lockingTable != "" {
		return i.tableExists(i.lockingTable, fmt.Sprintf(sLockingScript, i.lockingTable))
	}

	return i.tableExists(i.lockingTable, "")
}

func (i *MySQL) tableExists(table string, script string) (bool, error) {
	exists := false
	query := "SELECT true FROM information_schema.TABLES WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = ?"
	if err := i.db.QueryRow(query, table).Scan(&exists); err != nil {
		if err != sql.ErrNoRows {
			return false, err
		}

		if script == "" {
			return false, nil
		}
	}

	if !exists {
		if _, err := i.db.Exec(script); err != nil {
			return false, err
		}
	}

	return true, nil
}

func (i *MySQL) username() string {
	if i.tUsername != "" {
		return i.tUsername
	}

	query := "SELECT CURRENT_USER()"
	if err := i.db.QueryRow(query).Scan(i.tUsername); err != nil {
		i.tUsername = "-"
	}

	return i.tUsername
}
