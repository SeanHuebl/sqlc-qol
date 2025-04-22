package helpers

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"os"
)

type BaseTestCase struct {
	Name              string
	ExpectedContent   string
	CreateErr         bool
	ParseErr          bool
	GlobErr           bool
	FormatErr         bool
	ExpectedErrSubStr string
}
type QualifymodelsTC struct {
	BaseTestCase
	ModelContent string
	QueryContent string
}
type AddnosecTC struct {
	BaseTestCase
	HasCsv      bool
	CsvTargets  string
	InitContent string
	Targets     string
	OpenErr     bool
	PathErr     bool
	BaseDirErr  bool
	PrefixErr   bool
}

// executeErrors returns a modified set of function dependencies that simulate error conditions
// for testing purposes. It accepts a testCase struct with various error flags and the original
// implementations of four functions: parseFile, glob, createFile, and formatNode. Depending on
// which error flag in testCase is set to true, it replaces the corresponding function with a stub
// that returns a simulated error.
//
// The testCase struct contains the following fields:
//   - name: a descriptive name for the test case.
//   - modelContent, queryContent, expectedContent: strings representing test input and expected output.
//   - createErr: if true, the createFile function is replaced to simulate a file creation error.
//   - parseErr: if true, the parseFile function is replaced to simulate a parsing error.
//   - globErr: if true, the glob function is replaced to simulate a glob matching error.
//   - formatErr: if true, the formatNode function is replaced to simulate a formatting error.
//   - expectedErrSubStr: a substring expected to be present in the error message.
//
// The functions provided are as follows:
//   - parseFile: parses a source file and returns its AST representation.
//   - glob: finds files matching a pattern.
//   - createFile: creates a file with the specified name.
//   - formatNode: formats an AST node and writes it to the provided writer.
//
// Only one error is simulated based on the order of evaluation:
//  1. parseErr
//  2. globErr
//  3. createErr
//  4. formatErr
//
// If none of the error flags are set, the original functions are returned unmodified.
//
// Returns a tuple of four functions corresponding to parseFile, glob, createFile, and formatNode,
// where each function may be the original or a stub that returns a simulated error.
func ExecuteBaseTCErrors(
	testCase BaseTestCase,
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

	if testCase.ParseErr {
		pf = func(fset *token.FileSet, filename string, src any, mode parser.Mode) (f *ast.File, err error) {
			return nil, fmt.Errorf("simulated parsing error")
		}
		return pf, gf, cf, fn

	}

	if testCase.GlobErr {
		gf = func(pattern string) (matches []string, err error) {
			return nil, fmt.Errorf("simluated glob error")
		}
		return pf, gf, cf, fn

	}

	if testCase.CreateErr {
		cf = func(name string) (*os.File, error) {
			return nil, fmt.Errorf("simulated create error")
		}
		return pf, gf, cf, fn

	}

	if testCase.FormatErr {
		fn = func(dst io.Writer, fset *token.FileSet, node any) error {
			return fmt.Errorf("simulated format error")
		}
		return pf, gf, cf, fn
	}

	return pf, gf, cf, fn
}

// ExecuteAddnosecErrors returns a modified set of function dependencies that simulate error conditions
// for testing purposes. It accepts a testCase struct with various error flags and the original
// implementations of four functions: openFile, pathAbs, baseAbs, and hasPrefix. Depending on
// which error flag in testCase is set to true, it replaces the corresponding function with a stub
// that returns a simulated error or failure behavior.
//
// The testCase struct contains the following fields:
//   - csvPath: the path to a CSV file used in the test.
//   - initContent: the content to be transformed
//   - targets: a string representing one or more (seperated by commas) const decl to be processed.
//   - openErr: if true, the openFile function is replaced to simulate a file opening error.
//   - pathErr: if true, the pathAbs function is replaced to simulate an absolute path resolution error.
//   - baseDirErr: if true, the baseAbs function is replaced to simulate a base directory resolution error.
//   - prefixErr: if true, the hasPrefix function is replaced to always return false.
//   - expectedErrSubStr: a substring expected to be present in the error message.
//
// The functions provided are as follows:
//   - openFile: opens a file by name and returns an os.File pointer.
//   - pathAbs: returns the absolute form of a given path.
//   - baseAbs: returns the base directory of a given path.
//   - hasPrefix: checks whether a string has a specific prefix.
//
// Only one error is simulated based on the order of evaluation:
//  1. openErr
//  2. pathErr
//  3. baseDirErr
//  4. prefixErr
//
// If none of the error flags are set, the original functions are returned unmodified.
//
// Returns a tuple of four functions corresponding to openFile, pathAbs, baseAbs, and hasPrefix,
// where each function may be the original or a stub that returns a simulated error or failure behavior.
func ExecuteAddnosecErrors(
	testCase AddnosecTC,
	openFile func(name string) (*os.File, error),
	pathAbs func(path string) (string, error),
	baseAbs func(path string) (string, error),
	hasPrefix func(s string, prefix string) bool,
) (
	func(name string) (*os.File, error),
	func(path string) (string, error),
	func(path string) (string, error),
	func(s string, prefix string) bool,
) {
	oF := openFile
	pA := pathAbs
	bA := baseAbs
	hP := hasPrefix

	if testCase.OpenErr {
		oF = func(name string) (*os.File, error) {
			return nil, fmt.Errorf("simulated open file error")
		}
		return oF, pA, bA, hP
	}
	if testCase.PathErr {
		pA = func(path string) (string, error) {
			return "", fmt.Errorf("simulated path error")
		}
		return oF, pA, bA, hP
	}
	if testCase.BaseDirErr {
		bA = func(path string) (string, error) {
			return "", fmt.Errorf("simulated base directory error")
		}
		return oF, pA, bA, hP
	}
	if testCase.PrefixErr {
		hP = func(s, prefix string) bool {
			return false
		}
		return oF, pA, bA, hP
	}
	return oF, pA, bA, hP
}
