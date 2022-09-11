// Copyright © 2022 Aditya Khoirul Anam <adit@ditya.dev>
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package cmd

import (
	"fmt"
	"github.com/dityaaa/concept/migration"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show the status of each migration",
	Run: func(cmd *cobra.Command, args []string) {
		conceptStatus()
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func conceptStatus() {
	fmt.Println("Preparing...")
	mg := newMigration(true, nil)

	res, err := mg.Get()
	cobra.CheckErr(err)

	for _, dt := range res {
		fmt.Println(dt.ScriptName, " | ", migration.TranslateState(dt.Status))
	}
}
