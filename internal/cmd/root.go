// Copyright © 2022 Aditya Khoirul Anam <adit@ditya.dev>
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package cmd

import (
	"github.com/dityaaa/concept/database"
	"github.com/dityaaa/concept/database/mysql"
	"github.com/dityaaa/concept/migration"
	"github.com/theckman/yacspin"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "concept",
	Short: "MySQL Database Migration",
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
		viper.SetConfigType("dotenv")
		viper.SetConfigName(".env")
	}

	viper.AutomaticEnv() // read in environment variables that match

	err := viper.ReadInConfig()
	cobra.CheckErr(err)
}

func newMigration(withDatabase bool, hooks *migration.Hooks) *migration.Properties {
	var db database.Database

	if withDatabase {
		dbCfg, err := database.New(&database.Config{
			Host:     viper.GetString("host"),
			Port:     viper.GetUint32("port"),
			Username: viper.GetString("username"),
			Password: viper.GetString("password"),
			DbName:   viper.GetString("database"),
		})
		cobra.CheckErr(err)

		db, err = mysql.New(dbCfg)
		cobra.CheckErr(err)
	}

	mg, err := migration.New(&migration.Config{
		Path:     viper.GetString("migration_path"),
		Database: db,
		Hooks:    hooks,
	})
	cobra.CheckErr(err)

	return mg
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
