package qualifymodels

import (
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"

	"go/token"
	"os"
	"path"
	"path/filepath"

	"golang.org/x/tools/go/ast/astutil"
)

var (
	createFile = os.Create
	parseFile  = parser.ParseFile
	glob       = filepath.Glob
	formatNode = format.Node
)

// Run processes SQLC-generated query files to qualify model references.
// It takes three parameters:
//
// - modelPath: path to the file containing the external models (e.g., "internal/models/database.go")
//
// - queryGlob: a glob pattern to match SQLC-generated query files (e.g, "internal/database/*.sql.go")
//
// - modelImport: the import path for the external models package (e.g, "internal/models")
func Run(modelPath, queryGlob, modelImport string) error {
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

	// Find all query files that match the provided glob pattern.
	files, err := glob(queryGlob)
	if err != nil {
		return fmt.Errorf("failed to glob query files: %w", err)
	}

	// Process the query files
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

		outFile, err := createFile(file)
		if err != nil {
			return fmt.Errorf("failed to open file %s for writing: %w", file, err)
		}
		defer outFile.Close()

		if err := formatNode(outFile, fsetQuery, queryFile); err != nil {
			return fmt.Errorf("failed to write updated file %s: %w", file, err)
		}
	}
	return nil
}
