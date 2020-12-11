package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "goduyaoss",
	Short: "A simple Golang tool that crawls data from https://www.duyaoss.com",
	Long: `Extract information from the listed images on the website, and
			save the (updated) data to a local text file or database.
			`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Placeholder")
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
