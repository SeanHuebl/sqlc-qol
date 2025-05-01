package addnosec

import (
	"go/format"
	"go/parser"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/seanhuebl/sqlc-qol/internal/config"
	"github.com/seanhuebl/sqlc-qol/internal/helpers"
	"github.com/stretchr/testify/require"
)

func TestRun(t *testing.T) {
	tests := []helpers.AddnosecTC{
		{
			BaseTestCase: helpers.BaseTestCase{
				Name: "single target, no csv success",
				ExpectedContent: `package foo

const bar = "false flagged hardcoded credentials" // #nosec
`,
			},
			InitContent: `package foo

const bar = "false flagged hardcoded credentials"
`,
			Targets: "bar",
		},
		{
			BaseTestCase: helpers.BaseTestCase{
				Name: "multiple targets, no csv success",
				ExpectedContent: `package foo

const bar = "false flagged hardcoded credentials" // #nosec
const foobar = "false flagged hardcoded credentials" // #nosec
const c = "false flagged hardcoded credentials" // #nosec
`,
			},
			InitContent: `package foo

const bar = "false flagged hardcoded credentials"
const foobar = "false flagged hardcoded credentials"
const c = "false flagged hardcoded credentials"
`,
			Targets: "bar,foobar,c",
		},
		{
			BaseTestCase: helpers.BaseTestCase{
				Name: "single target, csv success",
				ExpectedContent: `package foo

const bar = "false flagged hardcoded credentials" // #nosec
`,
			},
			InitContent: `package foo

const bar = "false flagged hardcoded credentials"
`,
			HasCsv:     true,
			CsvTargets: "bar",
		},
		{
			BaseTestCase: helpers.BaseTestCase{
				Name: "multiple targets, csv success",
				ExpectedContent: `package foo

const bar = "false flagged hardcoded credentials" // #nosec
const foobar = "false flagged hardcoded credentials" // #nosec
const c = "false flagged hardcoded credentials" // #nosec
`,
			},
			InitContent: `package foo

const bar = "false flagged hardcoded credentials"
const foobar = "false flagged hardcoded credentials"
const c = "false flagged hardcoded credentials"
`,
			HasCsv:     true,
			CsvTargets: "bar,foobar,c",
		},
		{
			BaseTestCase: helpers.BaseTestCase{
				Name:              "simulate parse file error",
				ExpectedContent:   "",
				ParseErr:          true,
				ExpectedErrSubStr: "failed to parse",
			},
			InitContent: `package foo
const bar = "false flagged hardcoded credentials"
`,
			Targets: "bar",
		},
		{
			BaseTestCase: helpers.BaseTestCase{
				Name:              "simulate glob error",
				ExpectedContent:   "",
				GlobErr:           true,
				ExpectedErrSubStr: "failed to glob files with pattern",
			},
			InitContent: `package foo
const bar = "false flagged hardcoded credentials"
`,
			Targets: "bar",
		},
		{
			BaseTestCase: helpers.BaseTestCase{
				Name:              "simulate create file error",
				ExpectedContent:   "",
				CreateErr:         true,
				ExpectedErrSubStr: "failed to open file",
			},
			InitContent: `package foo
const bar = "false flagged hardcoded credentials"
`,
			Targets: "bar",
		},
		{
			BaseTestCase: helpers.BaseTestCase{
				Name:              "simulate format error",
				ExpectedContent:   "",
				FormatErr:         true,
				ExpectedErrSubStr: "failed to write formatted file",
			},
			InitContent: `package foo
const bar = "false flagged hardcoded credentials"
`,
			Targets: "bar",
		},
		{
			BaseTestCase: helpers.BaseTestCase{
				Name:              "simulate openFile error",
				ExpectedContent:   "",
				ExpectedErrSubStr: "failed to open CSV file",
			},
			InitContent: `package foo
const bar = "false flagged hardcoded credentials"
`,
			HasCsv:     true,
			OpenErr:    true,
			CsvTargets: "bar",
		},
		{
			BaseTestCase: helpers.BaseTestCase{
				Name:              "simulate pathAbs error",
				ExpectedContent:   "",
				ExpectedErrSubStr: "failed to get absolute path",
			},
			InitContent: `package foo
const bar = "false flagged hardcoded credentials"
`,
			HasCsv:     true,
			PathErr:    true,
			CsvTargets: "bar",
		},
		{
			BaseTestCase: helpers.BaseTestCase{
				Name:              "simulate baseAbs error",
				ExpectedContent:   "",
				ExpectedErrSubStr: "failed to get absolute base directory",
			},
			InitContent: `package foo
const bar = "false flagged hardcoded credentials"
`,
			HasCsv:     true,
			BaseDirErr: true,
			CsvTargets: "bar",
		},
		{
			BaseTestCase: helpers.BaseTestCase{
				Name:              "simulate hasPrefix error",
				ExpectedContent:   "",
				ExpectedErrSubStr: "is not within the allowed directory",
			},
			InitContent: `package foo
const bar = "false flagged hardcoded credentials"
`,
			HasCsv:     true,
			PrefixErr:  true,
			CsvTargets: "bar",
		},
		{
			BaseTestCase: helpers.BaseTestCase{
				Name:              "both csv and targets filled",
				ExpectedContent:   "",
				ExpectedErrSubStr: "cannot specify both targets and csvPath",
			},
			InitContent: `package foo
const bar = "false flagged hardcoded credentials"
`,
			HasCsv:     true,
			CsvTargets: "bar",
			Targets:    "bar",
		},
		{
			BaseTestCase: helpers.BaseTestCase{
				Name:              "both csv and targets empty",
				ExpectedContent:   "",
				ExpectedErrSubStr: "must specify either targets or csvPath",
			},
			InitContent: `package foo
const bar = "false flagged hardcoded credentials"
`,
			HasCsv:  false,
			Targets: "",
		},
	}
	for _, tc := range tests {
		t.Run(tc.Name, func(t *testing.T) {
			tc := tc

			parseFile = parser.ParseFile
			glob = filepath.Glob
			createFile = os.Create
			formatNode = format.Node

			tmpDir := t.TempDir()

			contentFile := filepath.Join(tmpDir, "content.sql.go")
			if err := os.WriteFile(contentFile, []byte(tc.InitContent), 0644); err != nil {
				t.Fatalf("failed to write content file: %v", err)
			}
			parseFile, glob, createFile, formatNode = helpers.ExecuteBaseTCErrorsANS(tc.BaseTestCase, parseFile, glob, createFile, formatNode)

			var err error
			var tempCSV *os.File
			if tc.HasCsv {
				openFile = os.Open
				pathAbs = filepath.Abs
				baseAbs = filepath.Abs
				hasPrefix = strings.HasPrefix

				openFile, pathAbs, baseAbs, hasPrefix = helpers.ExecuteAddnosecErrors(tc, openFile, pathAbs, baseAbs, hasPrefix)

				tmpDataDir := filepath.Join(tmpDir, "data")
				os.Mkdir(tmpDataDir, 0755)

				tempCSV, err = os.CreateTemp(tmpDataDir, "*.csv")

				if err != nil {
					t.Fatalf("failed to create temp csv: %v", err)
				}
				defer tempCSV.Close()
				if _, err := tempCSV.Write([]byte(tc.CsvTargets)); err != nil {
					t.Fatalf("failed to write to temp csv: %v", err)
				}
				err = Run(contentFile, tc.Targets, tempCSV.Name(), config.Config{AllowedBaseDir: tmpDataDir})

			} else {
				err = Run(contentFile, tc.Targets, "", config.Config{})
			}

			if tc.ExpectedErrSubStr != "" {
				require.Contains(t, err.Error(), tc.ExpectedErrSubStr)
				return
			} else if err != nil {
				t.Fatalf("run failed: %v", err)
			}

			got, err := os.ReadFile(contentFile)
			if err != nil {
				t.Fatalf("failed to read content file: %v", err)
			}
			formattedExpected, err := format.Source([]byte(tc.ExpectedContent))
			if err != nil {
				t.Fatalf("failed to format expected content with gofmt standards: %v", err)
			}
			if diff := cmp.Diff(string(formattedExpected), string(got)); diff != "" {
				t.Errorf("content file mismatch (-want +got)\n%s", diff)
			}
		})
	}
}
