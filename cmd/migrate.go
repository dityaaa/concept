// Copyright © 2022 Aditya Khoirul Anam <adit@ditya.dev>
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package cmd

import (
	_ "embed"
	"fmt"
	"github.com/dityaaa/concept/migration"
	"github.com/spf13/cobra"
	"time"
)

var migrateFresh bool

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Run the database migrations",
	Run: func(cmd *cobra.Command, args []string) {
		conceptMigrate()
	},
}

func init() {
	rootCmd.AddCommand(migrateCmd)
	migrateCmd.LocalFlags().BoolVar(&migrateFresh, "fresh", false, "Drop all tables an re-run all migrations")
}

func conceptMigrate() {
	spinner := newSpinner()

	nothingToMigrate := true
	fmt.Println("Preparing...")

	mg := newMigration(true, &migration.Hooks{
		PreMigrate: func(dt migration.Data) {
			nothingToMigrate = false
			spinner.Message(dt.ScriptName)
			spinner.Start()
			time.Sleep(4 * time.Second)
		},
		PostMigrate: func(dt migration.Data) {
			spinner.StopMessage(fmt.Sprintf("%s (%dms)", dt.ScriptName, dt.ExecutionTime))
			spinner.Stop()
		},
		MigrateErr: func(dt migration.Data, err error) {
			spinner.StopFailMessage(fmt.Sprintf("%s (%dms)", dt.ScriptName, dt.ExecutionTime))
			spinner.StopFail()
		},
	})

	err := mg.Migrate()
	if err != nil {
		spinner.StopFail()
		cobra.CheckErr(err)
	}

	if nothingToMigrate {
		fmt.Println("Nothing to migrate")
		return
	}

	fmt.Println("Database migration completed")
}
