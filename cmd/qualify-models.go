package cmd

import (
	"github.com/seanhuebl/sqlc-qol/internal/qualifymodels"
	"github.com/spf13/cobra"
)

var (
	modelFilePath string
	globPattern   string
	importPath    string
)

func init() {
	cmd := &cobra.Command{
		Use:   "qualify-models",
		Short: "Qualify bare model types in SQLC-generated code",
		Long: `Parses your SQLC models file to discover the struct names, then
re-writes the SQLC-generated .go files to qualify those types
(e.g. Transaction -> models.Transaction)
this is to be used in tandem with a script that moves
the SQLC models into an external global models package`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return qualifymodels.Run(modelFilePath, globPattern, importPath)
		},
	}

	cmd.Flags().
		StringVarP(&modelFilePath,
			"models",
			"m",
			"",
			"path to the Go source file defining your models (e.g. internal/models/models.go)")
	_ = cmd.MarkFlagRequired("models")

	cmd.Flags().
		StringVarP(&globPattern,
			"queries",
			"q",
			"",
			"glob matching SQLC-generated query files (e.g. internal/database/queries/*.sql.go)")
	_ = cmd.MarkFlagRequired("queries")

	cmd.Flags().
		StringVarP(&importPath,
			"import",
			"i",
			"",
			"import path for your models package (e.g. internal/models)")
	_ = cmd.MarkFlagRequired("import")

	rootCmd.AddCommand(cmd)
}
