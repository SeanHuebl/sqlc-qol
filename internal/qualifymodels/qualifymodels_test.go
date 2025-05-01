package qualifymodels

import (
	"go/format"
	"go/parser"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/seanhuebl/sqlc-qol/v2/internal/helpers"
	"github.com/stretchr/testify/require"
)

func TestRun(t *testing.T) {
	tests := []helpers.QualifymodelsTC{
		{
			BaseTestCase: helpers.BaseTestCase{
				Name: "Simple replacement",
				ExpectedContent: `package queries
import "internal/models"
func Foo() {
	var T models.Transaction
}
`,
			},

			ModelContent: `package models
type Transaction struct {}
`,
			QueryContent: `package queries
func Foo() {
	var T Transaction
}
`,
		},

		{
			BaseTestCase: helpers.BaseTestCase{
				Name: "Multiple replacement",
				ExpectedContent: `package queries
import "internal/models"
func Foo() {
	var T models.Transaction
}
func Bar() {
	var TR models.TransactionResponse
}
func FooBar() {
	var U models.User
}
`,
			},
			ModelContent: `package models
type Transaction struct {}
type TransactionResponse struct {}
type User struct {}
`,
			QueryContent: `package queries
func Foo() {
	var T Transaction
}
func Bar() {
	var TR TransactionResponse
}
func FooBar() {
	var U User
}
`,
		},
		{
			BaseTestCase: helpers.BaseTestCase{
				Name:              "simulate parse file error",
				ExpectedContent:   "",
				ParseErr:          true,
				ExpectedErrSubStr: "failed to parse",
			},
			ModelContent: `package models
type Transaction struct {}
`,
			QueryContent: `package queries
func Foo() {
	var T Transaction
}
`,
		},
		{
			BaseTestCase: helpers.BaseTestCase{
				Name:              "simulate walkDir error",
				ExpectedContent:   "",
				WalkErr:           true,
				ExpectedErrSubStr: "failed to walkDir",
			},
			ModelContent: `package models
type Transaction struct {}
`,
			QueryContent: `package queries
func Foo() {
	var T Transaction
}
`,
		},
		{
			BaseTestCase: helpers.BaseTestCase{
				Name:              "simulate create file error",
				ExpectedContent:   "",
				CreateErr:         true,
				ExpectedErrSubStr: "failed to open file",
			},
			ModelContent: `package models
type Transaction struct {}
`,
			QueryContent: `package queries
func Foo() {
	var T Transaction
}
`,
		},
		{
			BaseTestCase: helpers.BaseTestCase{
				Name:              "simulate format error",
				ExpectedContent:   "",
				FormatErr:         true,
				ExpectedErrSubStr: "failed to write updated file",
			},
			ModelContent: `package models
type Transaction struct {}
`,
			QueryContent: `package queries
func Foo() {
	var T Transaction
}
`,
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			// reset the testing state to avoid cases bleeding over
			parseFile = parser.ParseFile
			walkDir = filepath.WalkDir
			createFile = os.Create
			formatNode = format.Node

			tmpDir := t.TempDir()

			modelFile := filepath.Join(tmpDir, "models.go")
			queryFile := filepath.Join(tmpDir, "query.sql.go")

			if err := os.WriteFile(modelFile, []byte(tc.ModelContent), 0644); err != nil {
				t.Fatalf("failed to write model file: %v", err)
			}
			if err := os.WriteFile(queryFile, []byte(tc.QueryContent), 0644); err != nil {
				t.Fatalf("failed to write query file: %v", err)
			}
			parseFile, walkDir, createFile, formatNode = helpers.ExecuteBaseTCErrorsQM(tc.BaseTestCase, parseFile, walkDir, createFile, formatNode)

			err := Run(modelFile, queryFile, "internal/models")
			if tc.ExpectedErrSubStr != "" {
				require.Contains(t, err.Error(), tc.ExpectedErrSubStr)
				return
			} else if err != nil {
				t.Fatalf("run failed: %v", err)
			}

			got, err := os.ReadFile(queryFile)
			if err != nil {
				t.Fatalf("failed to read query file: %v", err)
			}
			formattedExpected, err := format.Source([]byte(tc.ExpectedContent))
			if err != nil {
				t.Fatalf("failed to format expected content with gofmt standards: %v", err)
			}
			if diff := cmp.Diff(string(formattedExpected), string(got)); diff != "" {
				t.Errorf("query file mismatch (-want +got)\n%s", diff)
			}
		})
	}
}
