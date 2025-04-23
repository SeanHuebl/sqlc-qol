package addnosec

import (
	"encoding/csv"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/seanhuebl/sqlc-qol/internal/config"
	"golang.org/x/tools/go/ast/astutil"
)

var (
	parseFile  = parser.ParseFile
	glob       = filepath.Glob
	createFile = os.Create
	formatNode = format.Node

	openFile  = os.Open
	pathAbs   = filepath.Abs
	baseAbs   = filepath.Abs
	hasPrefix = strings.HasPrefix
)

// Run scans all Go source files matching queryGlob and appends a “// #nosec” comment
// to any const declarations whose names you’ve specified via targets or csvPath.
// You must supply exactly one of targets (a comma‑separated list) or csvPath
// (pointing to a CSV file under config.AllowedBaseDir); otherwise Run returns an error.
//
// It works by:
//  1. Building a map of target names (from CSV or comma list).
//  2. Globbing for files via queryGlob.
//  3. Parsing each file’s AST, finding ast.ValueSpec nodes whose names match targets,
//     and injecting a `// #nosec` comment if one isn’t already present.
//  4. Rewriting each file in place with go/format.
//
// Parameters:
//   - queryGlob: glob pattern for selecting .go files (e.g. "internal/database/*.sql.go")
//   - targets: comma‑separated const names (mutually exclusive with csvPath)
//   - csvPath: path to a no‑header CSV listing const names (mutually exclusive with targets)
//   - config: holds AllowedBaseDir for sanitizing CSV paths
//
// Returns an error if:
//   - both or neither of targets/csvPath are provided,
//   - the CSV cannot be read/parsed or lies outside AllowedBaseDir,
//   - globbing fails,
//   - any file can’t be parsed, opened, or written.
func Run(queryGlob, targets, csvPath string, config config.Config) error {
	fmt.Fprintf(os.Stderr, "DEBUG: loading targets from %q or %q\n", targets, csvPath)

	var targetMap map[string]bool
	var err error

	if csvPath != "" && targets != "" {
		return fmt.Errorf("cannot specify both targets and csvPath")
	} else if targets == "" && csvPath == "" {
		return fmt.Errorf("must specify either targets or csvPath")
	}

	if csvPath != "" {
		targetMap, err = parseTargetsCSV(csvPath, config.AllowedBaseDir)
		if err != nil {
			return fmt.Errorf("error parsing CSV file: %w", err)
		}
	} else {
		targetMap = parseTargets(targets)
	}
	fmt.Fprintf(os.Stderr, "DEBUG: targetMap = %#v\n", targetMap)
	cwd, _ := os.Getwd()
	fmt.Fprintf(os.Stderr, "DEBUG: cwd=%q, glob=%q\n", cwd, queryGlob)

	files, err := glob(queryGlob)
	fmt.Fprintf(os.Stderr, "DEBUG: glob found %d files, err=%v\n", len(files), err)
	if err != nil {
		return fmt.Errorf("failed to glob files with pattern %q: %w", queryGlob, err)
	}

	for _, file := range files {
		fmt.Printf("Processing file: %s\n", file)

		fset := token.NewFileSet()
		f, err := parseFile(fset, file, nil, parser.ParseComments)
		if err != nil {
			return fmt.Errorf("failed to parse file %s: %w", file, err)
		}

		astutil.Apply(f, func(c *astutil.Cursor) bool {
			valSpec, ok := c.Node().(*ast.ValueSpec)
			if !ok {
				return true
			}
			for _, name := range valSpec.Names {
				if targetMap[name.Name] {
					fmt.Printf("  → tagging const %q in %s\n", name.Name, file)
					hasNoSec := false
					if valSpec.Comment != nil {
						for _, comment := range valSpec.Comment.List {
							if strings.Contains(comment.Text, "#nosec") {
								hasNoSec = true
								break
							}
						}
					}
					if !hasNoSec {
						newComment := &ast.Comment{
							Slash: name.End(),
							Text:  "// #nosec",
						}
						if valSpec.Comment == nil {
							valSpec.Comment = &ast.CommentGroup{
								List: []*ast.Comment{newComment},
							}
						} else {
							valSpec.Comment.List = append(valSpec.Comment.List, newComment)
						}

					}
				}
			}
			return true
		}, nil)
		fmt.Fprintf(os.Stderr, "DEBUG: opening %s for writing\n", file)
		outFile, err := createFile(file)
		if err != nil {
			return fmt.Errorf("failed to open file %s for writing: %w", file, err)
		}
		defer outFile.Close()
		fmt.Fprintf(os.Stderr, "DEBUG: writing formatted AST to %s\n", file)
		if err := formatNode(outFile, fset, f); err != nil {
			return fmt.Errorf("failed to write formatted file %s: %w", file, err)
		}
		fmt.Fprintf(os.Stderr, "DEBUG: write complete for %s\n", file)
	}
	return nil
}

func parseTargetsCSV(csvPath, allowedBaseDir string) (map[string]bool, error) {
	// while low risk in CLI, sanitizing to protect users as much as possible from security risk
	safePath, err := sanitizePath(csvPath, allowedBaseDir)
	if err != nil {
		return nil, err
	}
	// after sanitizing the path to make sure it is safe to open
	// we can tell the security analyzer that it is safe to ignore
	f, err := openFile(safePath) // #nosec
	if err != nil {
		return nil, fmt.Errorf("failed to open CSV file: %w", err)
	}
	defer f.Close()
	reader := csv.NewReader(f)
	targets, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to parse CSV file: %w", err)
	}
	targetMap := make(map[string]bool)

	for _, target := range targets {
		for _, name := range target {
			trimmed := strings.TrimSpace(name)
			if trimmed != "" {
				targetMap[name] = true
			}
		}
	}
	return targetMap, nil
}

func parseTargets(targets string) map[string]bool {
	targetMap := make(map[string]bool)
	for _, target := range strings.Split(targets, ",") {
		trimmed := strings.TrimSpace(target)
		if trimmed != "" {
			targetMap[trimmed] = true
		}
	}
	return targetMap
}

func sanitizePath(csvPath, baseDir string) (string, error) {
	absPath, err := pathAbs(csvPath)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %w", err)
	}
	baseAbs, err := baseAbs(baseDir)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute base directory: %w", err)
	}
	if !hasPrefix(absPath, baseAbs) {
		return "", fmt.Errorf("invalid path: %q is not within the allowed directory %q", absPath, baseAbs)
	}
	return absPath, nil
}
