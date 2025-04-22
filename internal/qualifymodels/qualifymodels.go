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
	"gopkg.in/yaml.v3"
)

var (
	parseFile  = parser.ParseFile
	glob       = filepath.Glob
	createFile = os.Create
	formatNode = format.Node
)

type sqlcConfig struct {
	Gen struct {
		Go struct {
			OutputModelsPackage     string `yaml:"output_models_package"`
			ModelsPackageImportPath string `yaml:"models_package_import_path"`
		}
	}
}

// Detects if SQLC is modern v2 post PR #3874 on March 6, 2025 to determine if qualify models is needed.
func isSQLCModern() bool {
	data, err := os.ReadFile("sqlc.yaml")
	if err != nil {
		return false
	}
	var cfg sqlcConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return false
	}
	return cfg.Gen.Go.OutputModelsPackage != "" &&
		cfg.Gen.Go.ModelsPackageImportPath != ""
}

// Run processes SQLC-generated query files and qualifies bare model type
// references by prefixing them with a package alias and injecting the
// corresponding import. If SQLC v2+ native model qualification is detected
// via isSQLCModern(), it logs a message and exits without modifying any files.
//
// Workflow:
//  1. Check for native SQLC qualification support; if present, skip processing.
//  2. Parse the models file at modelPath and collect all struct type names.
//  3. Derive the package alias from modelImport (last path element).
//  4. Glob for all query files matching queryGlob.
//  5. For each file:
//     a) Parse its AST and traverse all identifiers.
//     b) When an identifier matches a model name and is not already
//     part of a selector, replace it with `alias.Identifier`.
//     c) Ensure the import for modelImport is present.
//     d) Rewrite the file in place with `go/format`.
//
// Parameters:
//   - modelPath:       Path to the Go source file defining your models.
//   - queryGlob:       Glob pattern to match SQLC‑generated `.sql.go` files.
//   - modelImport:     Import path for your external models package.
//
// Returns:
//   - error: Any error encountered while parsing, globbing, or writing files.
//     Returns nil if native SQLC qualification is enabled or if all
//     files are successfully processed.
func Run(modelPath, queryGlob, modelImport string) error {
	// Check if the SQLC config is modern v2
	if isSQLCModern() {
		fmt.Println("Detected native SQLC model qualification, skipping qualify-models")
		return nil
	}
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
