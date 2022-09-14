// Copyright © 2022 Aditya Khoirul Anam <adit@ditya.dev>
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package mysql

import (
	"database/sql"
	_ "embed"
	"fmt"
	"github.com/dityaaa/concept/database"
	"github.com/go-sql-driver/mysql"
	"time"
)

//go:embed schema.sql
var script string

type Driver struct {
	database string
	table    string
	username string
	db       *sql.DB
}

func New(config *database.Config) (database.Database, error) {
	mysqlConfig := mysql.NewConfig()
	mysqlConfig.Net = "tcp"
	mysqlConfig.Addr = fmt.Sprintf("%s:%d", config.Host, config.Port)
	mysqlConfig.User = config.Username
	mysqlConfig.Passwd = config.Password
	mysqlConfig.DBName = config.DbName
	mysqlConfig.Params = map[string]string{
		"multiStatements": "true",
	}

	dsn := mysqlConfig.FormatDSN()
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}

	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	return Driver{
		database: config.DbName,
		table:    config.HistoryTable,
		username: config.Username,
		db:       db,
	}, nil
}

func (d Driver) Read(cat string) ([]*database.Row, error) {
	if _, err := d.tableExists(true); err != nil {
		return nil, err
	}

	if cat == "*" {
		cat = "%"
	}

	query := fmt.Sprintf("SELECT * FROM `%s` AS `h1` WHERE `h1`.`category` LIKE ? AND `h1`.`sequence` = (SELECT MAX(`h2`.`sequence`) FROM `schema_history` AS `h2` WHERE `h2`.`version` = `h1`.`version` AND `h2`.`category` = `h1`.`category`)", d.table)
	rows, err := d.db.Query(query, cat)
	if err != nil {
		return nil, err
	}

	res := make([]*database.Row, 0)

	for rows.Next() {
		var row database.Row

		err := rows.Scan(
			&row.Sequence,
			&row.Category,
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

func (d Driver) Insert(cat, ver string, name string, desc string, checksum string) (*database.Row, error) {
	query := fmt.Sprintf("INSERT INTO `%s` VALUES (NULL, ?, ?, ?, ?, ?, ?, ?, 0, 0)", d.table)

	currentMillis := time.Now().UnixMilli()
	res, err := d.db.Exec(query, cat, ver, name, desc, checksum, d.username, currentMillis)
	if err != nil {
		return nil, err
	}

	seq, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	return &database.Row{
		Sequence:    uint64(seq),
		Category:    cat,
		Version:     ver,
		ScriptName:  name,
		Description: desc,
		Checksum:    checksum,
		AppliedBy:   d.username,
		AppliedAt:   uint64(currentMillis),
	}, nil
}

func (d Driver) Update(seq uint64, execTime uint32, success bool) error {
	query := fmt.Sprintf("UPDATE `%s` SET `execution_time` = ?, `success` = ? WHERE `sequence` = ?", d.table)
	_, err := d.db.Exec(query, execTime, success, seq)
	return err
}

func (d Driver) Exec(script string) error {
	_, err := d.db.Exec(script)
	return err
}

func (d Driver) tableExists(create bool) (bool, error) {
	exists := false
	query := "SELECT true FROM information_schema.TABLES WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?"
	if err := d.db.QueryRow(query, d.database, d.table).Scan(&exists); err != nil {
		if err != sql.ErrNoRows {
			return false, err
		}

		if !create {
			return false, nil
		}
	}

	if !exists {
		if _, err := d.db.Exec(script); err != nil {
			return false, err
		}
	}

	return true, nil
}
