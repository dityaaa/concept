// Copyright © 2022 Aditya Khoirul Anam <adit@ditya.dev>
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package cmd

import (
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show the status of each migration",
	Run: func(cmd *cobra.Command, args []string) {
		conceptVersion()
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

func conceptVersion() {
	panic("implement me")
}
