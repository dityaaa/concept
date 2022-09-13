// Copyright © 2022 Aditya Khoirul Anam <adit@ditya.dev>
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package cmd

import (
	"fmt"
	"github.com/dityaaa/concept/migration"
	"github.com/spf13/cobra"
	"time"
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

	mg := newMigration(true, &migration.Hooks{
		PreRollback: func(dt migration.Data) {
			nothingToRollback = false
			spinner.Message(dt.ScriptName)
			spinner.Start()
			time.Sleep(4 * time.Second)
		},
		PostRollback: func(dt migration.Data) {
			spinner.StopMessage(fmt.Sprintf("%s (%dms)", dt.ScriptName, dt.ExecutionTime))
			spinner.Stop()
		},
		RollbackErr: func(dt migration.Data, err error) {
			spinner.StopFailMessage(fmt.Sprintf("%s (%dms)", dt.ScriptName, dt.ExecutionTime))
			spinner.StopFail()
		},
	})

	err := mg.Rollback(rollbackSteps)
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
