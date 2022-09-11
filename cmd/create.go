// Copyright © 2022 Aditya Khoirul Anam <adit@ditya.dev>
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package cmd

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new migration file",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		conceptCreate(args[0])
	},
}

func init() {
	rootCmd.AddCommand(createCmd)
}

func conceptCreate(name string) {
	mg := newMigration(false, nil)

	files, err := mg.Create(name)
	cobra.CheckErr(err)

	fmt.Println("Migration files successfully created")
	fmt.Println(color.GreenString("✔"), files[0])
	fmt.Println(color.GreenString("✔"), files[1])
}
