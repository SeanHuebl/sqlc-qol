package cmd

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "sql-qol",
	Short: "CLI tool to enhance SQLC generated code",
	Long: `sqlc-qol is a CLI tool that provides post-processing commands for SQLC generated code.
It includes features such as:
Adding gosec ignore comments when a query gets incorrectly flagged as hardcoded credentials.
Qualifying model references for when the models get moved to an external models directory,
It will go through the structs and replace all references in the SQLC content with 'models.'.
Use one of the subcommands for the desired operation.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("sqlc-col: use -h to see available subcommands.")
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
