# sqlc-qol

A command‑line utility that streamlines common post‑processing tasks for SQLC‑generated Go code. It provides two subcommands:

- `qualify-models`: Automatically qualify model type references in SQLC output to point at your external models package.
- `add-nosec`: Append `// #nosec` comments to specified constant declarations to suppress gosec false positives.

Whether you integrate it into a CI pipeline or run it locally, **sqlc‑qol** saves you time by automating repetitive edits and keeping your codebase consistent.

## Table of Contents

1. [Features](#features)
2. [Installation](#installation)
3. [Usage](#usage)
   - [Global Usage](#global-usage)
   - [Commands](#commands)
     - [qualify-models](#qualify-models)
     - [add-nosec](#add-nosec)
4. [Directory Structure](#directory-structure)
5. [Configuration & Requirements](#configuration--requirements)
6. [Integration Examples](#integration-examples)
7. [Contributing](#contributing)
8. [License](#license)

---

## Features

- `qualify-models`: Parses your external models file to collect struct names, then rewrites all SQLC‑generated query files to qualify bare references (e.g., `Transaction` → `models.Transaction`) and inject the necessary import.
- `add-nosec`: Scans for constant declarations matching a glob and a list of names (or a CSV), and appends `// #nosec` to each to suppress gosec warnings about hardcoded values.
- **No manual editing**: Automate repetitive maintenance tasks that would otherwise be lost whenever you re‑run `sqlc generate`.

---

## Installation

### Go Install (recommended)

Requires Go 1.16 or later:

```bash
go install github.com/seanhuebl/sqlc-qol@latest
```

This installs the `sqlc-qol` binary to your `$GOPATH/bin` (or `$HOME/go/bin`).

### Pre‑built Binaries

1. Browse the [Releases page](https://github.com/seanhuebl/sqlc-qol/releases).
2. Download the binary for your OS and architecture.
3. Move it into a directory on your `$PATH`, e.g.:

   ```bash
   mv sqlc-qol /usr/local/bin/
   ```

### Build from Source

```bash
git clone https://github.com/seanhuebl/sqlc-qol.git
cd sqlc-qol
go build -o sqlc-qol .
```

This produces a local `./sqlc-qol` executable.

### Verify Installation

```bash
sqlc-qol --help
```

You should see the global usage and available subcommands.

---

## Usage

### Global Usage

```text
Usage:
  sqlc-qol [command]

Available Commands:
  qualify-models  Qualify model types in SQLC query files
  add-nosec       Add // #nosec comments to specified constants
  help            Help about any command
  completion      Generate shell completion scripts

Flags:
  -h, --help   help for sqlc-qol

Use "sqlc-qol [command] --help" for more information about a command.
```

### Commands

#### qualify-models

Parses your external models file to discover all struct names, then rewrites SQLC‑generated query files to fully qualify those types and inject the import.

Starting with modern SQLC v2 configurations (as of PR #3874 on March 6, 2025) that include output_models_package and models_package_import_path, this tool will detect SQLC's native qualification support and skip processing, preserving the default SQLC behavior.

```bash
sqlc-qol qualify-models \
  --models   internal/models/database.go \
  --queries  "internal/database/*.sql.go" \
  --import   internal/models
```

**Flags**:

- `--models`, `-m` (required): Path to your Go source file containing model definitions (e.g., `internal/models/database.go`).
- `--dir`, `-d` (required): root directory where your database files live (e.g. `internal/database`).
- `--import`, `-i` (required): Import path for your models package (e.g., `internal/models`).

#### add-nosec

Appends `// #nosec` comments to constant declarations you specify, preventing gosec false positives after each run of `sqlc generate`.

```bash
# By targets list:
sqlc-qol add-nosec \
  "internal/database/*.sql.go" \
  --targets=createRefreshToken,revokeToken

# Or from a CSV file (no headers, located in ./data):
sqlc-qol add-nosec \
  "internal/database/*.sql.go" \
  --csv=./data/targets.csv
```

**Flags**:

- `--targets`, `-t`: Comma‑separated list of constant names to annotate.
- `--csv`, `-c`: Path to a CSV (no headers) listing one or more constant names; files **must** live under `./data`.

> **Note:** You must specify exactly one of `--targets` or `--csv`.

---

## Directory Structure

```plaintext
sqlc-qol/
├── cmd/
│   ├── root.go           # Cobra entrypoint and global setup
│   ├── add-nosec.go      # CLI wiring for add-nosec
│   └── qualify-models.go # CLI wiring for qualify-models
├── internal/
│   ├── addnosec/
│   │   └── addnosec.go   # Business logic for adding // #nosec
│   └── qualifymodels/
│       └── qualifymodels.go # Business logic for qualifying models
├── go.mod
├── go.sum
└── main.go               # Entrypoint: calls cmd.Execute()
```

---

## Configuration & Requirements

- **Go 1.16+**: Required for building and running.
- **Allowed CSV directory**: CSV files for the `--csv` flag must reside under `./data`.\
  The tool safeguards against directory traversal and only reads files within this directory.

---

## Integration Examples

To automate post‑generation tasks in a Makefile:

```makefile
.PHONY: all generate postprocess move rename qualify mocks

all: generate postprocess move rename qualify mocks

generate:
 sqlc generate

# suppress false positives
postprocess: generate
 sqlc-qol add-nosec \
 "internal/database/*.sql.go" \
 --csv=./data/targets.csv

# move the models file to the globel models pkg
move:
 mv ./internal/database/models.go ./internal/models/db.go

# rename the package name to models
rename: move
 sed -i "s/^package database$$/package models/" internal/models/db.go

# qualify all the models in all the files within database dir (-d)
qualify: rename move
 sqlc-qol qualify-models \
 -m internal/models/db.go \
 -d internal/database \
 -i github.com/seanhuebl/unity-wealth/internal/models
 goimports -w internal/database

# finish by regenerating mocks
mocks: qualify
 mockery --config mockery_database.yaml
 mockery --config mockery_auth.yaml
```

You can also integrate these commands into CI pipelines (GitHub Actions, GitLab CI, etc.) to ensure consistency on every build.

---

## Contributing

Contributions and suggestions are welcome! Please take the following steps:

1. Fork the repository.
2. Create a feature branch: `git checkout -b feature/YourFeature`.
3. Make your changes, including tests and documentation.
4. Open a pull request against `main`.

Please follow Go best practices and include table‑driven tests (using `go-cmp` for comparisons) for any logic changes.

---

## License

This project is licensed under the **MIT License**. See the [LICENSE](LICENSE) file for details.
