package cmd

import (
	"github.com/seanhuebl/sqlc-qol/internal/addnosec"
	"github.com/spf13/cobra"
)

var (
	addTargets string
	addCSV     string
)

func init() {

	cmd := &cobra.Command{
		Use:   "add-nosec",
		Short: "Add gosec // #nosec comments to SQLC generated code for targeted consts",
		Long: `Scans Go source files matching a glob pattern for targeted consts that are flagged by gosec as hardcoded credentials.
It adds a // #nosec comment to the const declaration to ignore the gosec warning.`,
		Args: cobra.ExactArgs(1), // Expecting a single argument: the glob pattern
		RunE: func(cmd *cobra.Command, args []string) error {
			globPattern := args[0]
			return addnosec.Run(globPattern, addTargets, addCSV, cfg)
		},
	}

	cmd.Flags().
		StringVarP(&addTargets,
			"targets", "t",
			"",
			"Comma-separated list of target consts to add gosec ignore comments for")

	cmd.Flags().
		StringVarP(&addCSV,
			"csv",
			"c",
			"",
			"path to CSV file containing target consts (no headers)")

	cmd.MarkFlagsMutuallyExclusive("targets", "csv")
	_ = cmd.MarkFlagFilename("csv", "csv")

	rootCmd.AddCommand(cmd)
}
