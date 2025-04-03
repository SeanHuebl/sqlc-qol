package qualifymodels

import (
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
)

func TestRun(t *testing.T) {
	tests := []struct {
		name              string
		modelContent      string
		queryContent      string
		expectedContent   string
		createErr         bool
		parseErr          bool
		globErr           bool
		formatErr         bool
		expectedErrSubStr string
	}{
		{
			name: "Simple replacement",
			modelContent: `package models
type Transaction struct {}
`,
			queryContent: `package queries
func Foo() {
	var T Transaction
}
`,
			expectedContent: `package queries
import "internal/models"
func Foo() {
	var T models.Transaction
}
`,
		},
		{
			name: "Multiple replacement",
			modelContent: `package models
type Transaction struct {}
type TransactionResponse struct {}
type User struct {}
`,
			queryContent: `package queries
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
			expectedContent: `package queries
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
		{
			name: "simulate create file error",
			modelContent: `package models
type Transaction struct {}
`,
			queryContent: `package queries
func Foo() {
	var T Transaction
}
`,
			expectedContent:   "",
			createErr:         true,
			expectedErrSubStr: "failed to open file",
		},	
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {

			tmpDir := t.TempDir()

			modelFile := filepath.Join(tmpDir, "models.go")
			queryFile := filepath.Join(tmpDir, "query.sql.go")

			if err := os.WriteFile(modelFile, []byte(tc.modelContent), 0644); err != nil {
				t.Fatalf("failed to write model file: %v", err)
			}
			if err := os.WriteFile(queryFile, []byte(tc.queryContent), 0644); err != nil {
				t.Fatalf("failed to write query file: %v", err)
			}
			parseFile, glob, createFile, formatNode = executeErrors(tc, parseFile, glob, createFile, formatNode)

			err := Run(modelFile, queryFile, "internal/models")
			if tc.expectedErrSubStr != "" {
				require.Contains(t, err.Error(), tc.expectedErrSubStr)
				return
			} else if err != nil {
				t.Fatalf("run failed: %v", err)
			}

			got, err := os.ReadFile(queryFile)
			if err != nil {
				t.Fatalf("failed to read query file: %v", err)
			}
			formattedExpected, err := format.Source([]byte(tc.expectedContent))
			if err != nil {
				t.Fatalf("failed to format expected content with gofmt standards: %v", err)
			}
			if diff := cmp.Diff(string(formattedExpected), string(got)); diff != "" {
				t.Errorf("query file mismatch (-want +got)\n%s", diff)
			}
		})
	}
}

// Helpers

func executeErrors(
	testCase struct {
		name              string
		modelContent      string
		queryContent      string
		expectedContent   string
		createErr         bool
		parseErr          bool
		globErr           bool
		formatErr         bool
		expectedErrSubStr string
	},
	parseFile func(fset *token.FileSet, filename string, src any, mode parser.Mode) (f *ast.File, err error),
	glob func(pattern string) (matches []string, err error),
	createFile func(name string) (*os.File, error),
	formatNode func(dst io.Writer, fset *token.FileSet, node any) error,
) (
	func(fset *token.FileSet, filename string, src any, mode parser.Mode) (f *ast.File, err error),
	func(pattern string) (matches []string, err error),
	func(name string) (*os.File, error),
	func(dst io.Writer, fset *token.FileSet, node any) error,
) {

	pf := parseFile
	gf := glob
	cf := createFile
	fn := formatNode

	if testCase.parseErr {
		pf = func(fset *token.FileSet, filename string, src any, mode parser.Mode) (f *ast.File, err error) {
			return nil, fmt.Errorf("simulated parsing error")
		}
		return pf, gf, cf, fn

	}

	if testCase.globErr {
		gf = func(pattern string) (matches []string, err error) {
			return nil, fmt.Errorf("simluated glob error")
		}
		return pf, gf, cf, fn

	}

	if testCase.createErr {
		cf = func(name string) (*os.File, error) {
			return nil, fmt.Errorf("simulated create error")
		}
		return pf, gf, cf, fn

	}

	if testCase.formatErr {
		fn = func(dst io.Writer, fset *token.FileSet, node any) error {
			return fmt.Errorf("simulated format error")
		}
		return pf, gf, cf, fn
	}

	return pf, gf, cf, fn
}
