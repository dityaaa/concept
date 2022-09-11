// Copyright © 2022 Aditya Khoirul Anam <adit@ditya.dev>
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package database

import (
	"database/sql"
	_ "embed"
)

type Driver rune

const (
	MySQLDriver Driver = iota
)

type Row struct {
	Sequence      uint64
	Category      string
	Version       string
	ScriptName    string
	Description   string
	Checksum      string
	AppliedBy     string
	AppliedAt     uint64
	ExecutionTime uint32
	Success       bool
}

type Config struct {
	Driver   Driver
	Host     string
	Port     uint32
	Username string
	Password string
	DbName   string

	HistoryTable string

	Db *sql.DB
}

type Database interface {
	Read(cat string) ([]*Row, error)
	Insert(cat, ver string, name string, desc string, checksum string) (*Row, error)
	Update(seq uint64, execTime uint32, success bool) error
	Exec(script string) error
}

func New(config *Config) (*Config, error) {
	if config.Host == "" {
		config.Host = "localhost"
	}

	if config.HistoryTable == "" {
		config.HistoryTable = "schema_history"
	}

	return config, nil
}
