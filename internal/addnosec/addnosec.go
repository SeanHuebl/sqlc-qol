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

func Run(queryGlob, targets, csvPath string, config config.Config) error {
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

	files, err := glob(queryGlob)
	if err != nil {
		return fmt.Errorf("failed to glob files with pattern %q: %w", queryGlob, err)
	}

	for _, file := range files {
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
		outFile, err := createFile(file)
		if err != nil {
			return fmt.Errorf("failed to open file %s for writing: %w", file, err)
		}
		defer outFile.Close()

		if err := formatNode(outFile, fset, f); err != nil {
			return fmt.Errorf("failed to write formatted file %s: %w", file, err)
		}
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
