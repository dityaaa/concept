// Copyright © 2022 Aditya Khoirul Anam <adit@ditya.dev>
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package cmd

import (
	"fmt"
	"github.com/dityaaa/concept"
	"github.com/spf13/cobra"
)

var rollbackSteps int

var rollbackCmd = &cobra.Command{
	Use:   "rollback",
	Short: "Rollback the last database migration",
	Run: func(cmd *cobra.Command, args []string) {
		conceptRollback()
	},
}

func init() {
	rootCmd.AddCommand(rollbackCmd)
	rollbackCmd.Flags().IntVar(&rollbackSteps, "steps", 1, "The number of migrations to be reverted")
}

func conceptRollback() {
	spinner := newSpinner()

	nothingToRollback := true
	fmt.Println("Preparing...")

	con := newConcept(true, &concept.Hooks{
		PreMigrate: func(mg *concept.Migration) {
			nothingToRollback = false
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

	err := con.Rollback(rollbackSteps)
	if err != nil {
		spinner.StopFail()
		cobra.CheckErr(err)
	}

	if nothingToRollback {
		fmt.Println("Rollback is not available")
		return
	}

	fmt.Println("Migration successfully reverted")
}
