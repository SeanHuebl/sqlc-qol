package qualifymodels

import (
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/ast/astutil"
)

var (
	parseFile  = parser.ParseFile
	createFile = os.Create
	formatNode = format.Node
	walkDir    = filepath.WalkDir
)

// Run processes Go source files under a given directory and qualifies bare
// model type references by prefixing them with a package alias and injecting
// the corresponding import.

// Workflow:
//   1. Check for native SQLC qualification support; if present, skip processing.
//   2. Parse the models file at modelPath and collect all struct type names.
//   3. Derive the package alias from modelImport (last path element).
//   4. Recursively walk all `.go` files under rootDir, skipping the model file
//      itself and any vendor or hidden directories.
//   5. For each discovered file:
//      a) Parse its AST and traverse all identifiers.
//      b) When an identifier matches a model name and is not already
//         part of a selector, replace it with `alias.Identifier`.
//      c) Ensure the import for modelImport is present.
//      d) Overwrite the file in place using `go/format`.
//
// Parameters:
//   - modelPath:   Path to the Go source file defining your models.
//   - rootDbDir:     Directory root in which to search for `.go` files to update.
//   - modelImport: Import path for your external models package.
//
// Returns:
//   - error: Any error encountered while parsing, walking the directory, or
//     writing files. Returns nil if native SQLC qualification is enabled or
//     if all files are successfully processed.

func Run(modelPath, rootDbDir, modelImport string) error {
	// Create new file set and parse the models file.
	fset := token.NewFileSet()
	modelFile, err := parseFile(fset, modelPath, nil, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("failed to parse model file: %w", err)
	}

	// Extract all struct names defined in the models file.
	modelNames := make(map[string]bool)
	for _, decl := range modelFile.Decls {
		genericDecl, ok := decl.(*ast.GenDecl)
		if !ok || genericDecl.Tok != token.TYPE {
			continue
		}
		for _, spec := range genericDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}
			if _, ok := typeSpec.Type.(*ast.StructType); ok {
				modelNames[typeSpec.Name.Name] = true
			}
		}
	}
	// Create package alias from the modelImport path
	pkgAlias := path.Base(modelImport)

	var files []string
	if err := walkDir(rootDbDir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(p, ".go") {
			return nil
		}
		if filepath.Clean(p) == filepath.Clean(modelPath) {
			return nil
		}
		files = append(files, p)
		return nil
	}); err != nil {
		return fmt.Errorf("failed to walkDir %s: %w", rootDbDir, err)
	}

	// Process the files
	for _, file := range files {
		fsetQuery := token.NewFileSet()
		queryFile, err := parseFile(fsetQuery, file, nil, parser.ParseComments)
		if err != nil {
			return fmt.Errorf("failed to parse query file %s: %w", file, err)
		}

		// Traverse AST to find bare identifiers that match the model names.
		astutil.Apply(queryFile, func(c *astutil.Cursor) bool {
			ident, ok := c.Node().(*ast.Ident)
			if !ok {
				return true
			}

			// Check if ident matches one of the model names
			if modelNames[ident.Name] {
				// If ident is already part of selector expression skip
				if _, ok := c.Parent().(*ast.SelectorExpr); ok {
					return true
				}
				// Replace bare ident with qualified selector expression (e.g, models.Transaction)
				newNode := &ast.SelectorExpr{
					X:   ast.NewIdent(pkgAlias),
					Sel: ast.NewIdent(ident.Name),
				}
				c.Replace(newNode)
			}
			return true
		}, nil)

		astutil.AddImport(fsetQuery, queryFile, modelImport)

		// This is so the defer happens after each file is processed
		// and not after all files are processed
		if err := func() error {

			outFile, err := createFile(file)

			if err != nil {
				return fmt.Errorf("failed to open file %s for writing: %w", file, err)
			}
			defer outFile.Close()

			return formatNode(outFile, fsetQuery, queryFile)
		}(); err != nil {
			return fmt.Errorf("failed to write updated file %s: %w", file, err)
		}
	}
	return nil
}
