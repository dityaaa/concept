// Copyright © 2022 Aditya Khoirul Anam <adit@ditya.dev>
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package cmd

import (
	"database/sql"
	"github.com/dityaaa/concept"
	"github.com/dityaaa/concept/database/mysql"
	"github.com/dityaaa/concept/source/file"
	mysql2 "github.com/go-sql-driver/mysql"
	"github.com/theckman/yacspin"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "concept",
	Short: "MySQL Database Migration by https://ditya.dev/",
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is .env)")
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.AddConfigPath("./")
		viper.SetConfigType("yaml")
		viper.SetConfigName("config")
	}

	viper.AutomaticEnv() // read in environment variables that match

	err := viper.ReadInConfig()
	cobra.CheckErr(err)

	viper.Set("_concept._config-initialized", true)
}

func newConcept(withDatabase bool, hooks *concept.Hooks) *concept.Concept {
	mysqlCfg := mysql2.NewConfig()
	mysqlCfg.Net = "tcp"
	mysqlCfg.Addr = viper.GetString("driver.mysql.host") + ":" + viper.GetString("driver.mysql.port")
	mysqlCfg.User = viper.GetString("driver.mysql.username")
	mysqlCfg.Passwd = viper.GetString("driver.mysql.password")
	mysqlCfg.DBName = viper.GetString("driver.mysql.database")
	mysqlCfg.Timeout = 10 * time.Second
	mysqlCfg.Params = map[string]string{
		"multiStatements": "true",
	}

	db, err := sql.Open("mysql", mysqlCfg.FormatDSN())
	cobra.CheckErr(err)

	db.SetConnMaxLifetime(time.Minute * 3)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)
	dbDrv, err := mysql.WithInstance(db, mysql.Config{
		HistoryTable: viper.GetString("history-table"),
		LockingTable: viper.GetString("locking-table"),
	})

	scDrv, err := file.Open("file://" + viper.GetString("migration-path"))

	c, err := concept.NewWithInstance(dbDrv, scDrv)
	cobra.CheckErr(err)

	c.SetHooks(hooks)

	return c
}

func newSpinner() *yacspin.Spinner {
	cfg := yacspin.Config{
		Frequency:         100 * time.Millisecond,
		CharSet:           yacspin.CharSets[14], //alt 14, 59
		Suffix:            " ",
		StopCharacter:     "✔",
		StopColors:        []string{"fgGreen"},
		StopFailCharacter: "✘",
		StopFailColors:    []string{"fgRed"},
	}

	spinner, err := yacspin.New(cfg)
	cobra.CheckErr(err)

	return spinner
}
