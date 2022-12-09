// Copyright © 2022 Aditya Khoirul Anam <adit@ditya.dev>
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package cmd

import (
	"fmt"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var createWithReverseFile bool

var createCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new migration file",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		conceptCreate(args[0])
	},
}

func init() {
	rootCmd.AddCommand(createCmd)
	createCmd.Flags().BoolVar(&createWithReverseFile, "with-reverse", false, "create migration file with its reverse migration file")
}

func conceptCreate(name string) {
	con := newConcept(true, nil)

	files, err := con.Create(name, createWithReverseFile)
	cobra.CheckErr(err)

	fmt.Println("Migration files successfully created")
	for _, name := range files {
		fmt.Println(color.GreenString("✔"), name)
	}
}
