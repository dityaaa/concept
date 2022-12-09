// Copyright © 2022 Aditya Khoirul Anam <adit@ditya.dev>
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package cmd

import (
	_ "embed"
	"fmt"
	"github.com/dityaaa/concept"
	"github.com/spf13/cobra"
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

	con := newConcept(true, &concept.Hooks{
		PreMigrate: func(mg *concept.Migration) {
			nothingToMigrate = false
			spinner.Message(mg.AdvanceScript.Identifier)
			spinner.Start()
		},
		PostMigrate: func(mg *concept.Migration) {
			spinner.StopMessage(fmt.Sprintf("%s (%dms)", mg.AdvanceScript.Identifier, mg.ExecutionTime))
			spinner.Stop()
		},
		MigrateErr: func(mg *concept.Migration, err error) {
			spinner.StopFailMessage(fmt.Sprintf("%s (%dms)", mg.AdvanceScript.Identifier, mg.ExecutionTime))
			spinner.StopFail()
		},
	})

	err := con.Migrate(-1)
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
