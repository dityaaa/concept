// Copyright © 2022 Aditya Khoirul Anam <adit@ditya.dev>
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package cmd

import (
	"github.com/spf13/cobra"
)

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Show the status of each migration",
	Run: func(cmd *cobra.Command, args []string) {
		conceptClean()
	},
}

func init() {
	rootCmd.AddCommand(cleanCmd)
}

func conceptClean() {
	panic("implement me")
}
