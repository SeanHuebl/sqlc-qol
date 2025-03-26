# sqlc-qol

CLI Tool to improve the quality of life for devs, who are using SQLC, with automation.\
Contribution is welcome if you would like to submit any other features!

## Table of Contents

- [Installating the CLI Tool](#installation)
- [qualify-models](#qualify-models)
- [add-nosec](#add-nosec)

## Directory Tree

```plaintext
sqlc-qol/
├── cmd/
│   ├── root.go              # Root command that initializes Cobra.
│   ├── add-nosec.go         # Implements the 'add-nosec' subcommand.
│   └── qualify-models.go    # Implements the 'qualify-models' subcommand.
├── internal/
│   ├── addnosec/
│   │   └── addnosec.go      # Business logic for adding // #nosec comments.
│   └── qualifymodels/
│       └── qualifymodels.go # Business logic for qualifying model references.
├── go.mod                   
├── go.sum                   
├── main.go                  
└── README.md                
```

## Installation

You have several options to install **sqlc-qol**:

### Installing via Go Install (recommended)

If you have Go installed (version 1.16 or higher is recommended), you can install **sqlc-qol** directly from the source repository using:

```bash
go install github.com/seanhuebl/sqlc-qol@latest
```

This command downloads, builds, and installs the tool into your Go bin directory (usually `$GOPATH/bin` or `$HOME/go/bin`).

### Installing a Pre-built Binary

Alternatively, you can download a pre-built binary for your platform:

1. Visit the [Releases](https://github.com/seanhuebl/sqlc-qol/releases) page.
2. Download the binary that matches your operating system and architecture.
3. Place the binary in a directory that's part of your system's `PATH` (for example, `/usr/local/bin` on Unix-like systems).

### Building from Source

If you prefer to build the tool yourself, follow these steps:

```bash
git clone https://github.com/seanhuebl/sqlc-qol.git
cd sqlc-qol
go build -o sqlc-qol
```

This creates an executable named `sqlc-qol` that you can run.

### Verifying Your Installation

After installation, verify that **sqlc-qol** is correctly installed by running:

```bash
sqlc-qol --help
```

---

Below is a refined version of the **qualify-models** section that combines the "Why It's Needed" and "Why Use qualify-models" parts into one cohesive overview:

---

## qualify-models

The `qualify-models` command streamlines your SQLC workflow when you decouple your models from your generated query files. Once you move SQLC-generated models into a dedicated external models package (e.g., `internal/models`) for enhanced modularity, the query files continue to reference model types without a package qualifier. Since every run of `sqlc generate` overwrites your query files, any manual changes are lost.

### qualify-models Overview

**Purpose & Rationale:**  
The primary goal of `qualify-models` is to ensure that your SQLC-generated query files consistently reference your external models package. By automating this process, the tool:

- **Eliminates Manual Adjustments:**  
  Avoid the unsustainable task of manually qualifying model references every time `sqlc generate` is run.
  
- **Prevents Ambiguity:**  
  Unqualified model identifiers (e.g., `Transaction`) can lead to confusion or errors. The tool converts these into fully qualified names (e.g., `models.Transaction`), ensuring that all references clearly point to the dedicated models package.
  
- **Maintains Modular Code:**  
  The command enforces a clean separation between your query logic and your models, which is particularly beneficial in CI/CD pipelines and collaborative projects where consistency is critical.

**Method:**  
The tool operates in three key steps:

1. **Extracting Model Names:**  
   It parses your specified models file (via the `-model` flag, e.g., `internal/models/database.go`) to collect all struct names defined in your external models package.

2. **Qualifying Model References:**  
   It uses the `-queryglob` flag to locate all SQLC-generated query files (e.g., `"internal/database/*.sql.go"`) and rewrites any bare model identifiers by prefixing them with your models package name (e.g., turning `Transaction` into `models.Transaction`).

3. **Adding Necessary Imports:**  
   If a query file does not already import your models package, the tool automatically inserts the required import statement (using the value provided via the `-modelimport` flag, e.g., `internal/models`).

### qualify-models Flags

- **`-model`**  
  - **Description:** Specifies the path to the file containing your external models.  
  - **Example:**  

    ```bash
    -model=internal/models/database.go
    ```

- **`-queryglob`**  
  - **Description:** Defines a glob pattern to match all SQLC-generated query files to process.  
  - **Example:**  

    ```bash
    -queryglob="internal/database/*.sql.go"
    ```

- **`-modelimport`**  
  - **Description:** Specifies the import path for your external models package.  
  - **Example:**  

    ```bash
    -modelimport=internal/models
    ```

### qualify-models Usage Example

After setting up your dedicated models package and ensuring your SQLC-generated queries are ready, run:

```bash
sqlc-qol qualify-models -model=internal/models/database.go \
                        -queryglob="internal/database/*.sql.go" \
                        -modelimport=internal/models
```

In this example, the command:

- Reads model definitions from `internal/models/database.go`.
- Processes all SQLC-generated files in `internal/database/` that match the `.sql.go` pattern.
- Rewrites the files to qualify model references with `models.` and adds the `internal/models` import if it's missing.

### qualify models Workflow Integration

Because `sqlc generate` overwrites your query files each time it runs, integrating `qualify-models` into your build or post-generation script is essential for maintaining a decoupled, modular codebase. For example, add the following target to your Makefile:

```makefile
.PHONY: generate
generate:
 sqlc generate
 sqlc-qol qualify-models -model=internal/models/database.go -queryglob="internal/database/*.sql.go" -modelimport=internal/models
```

This integration ensures that your codebase remains consistent and modular—automatically updating model references every time you regenerate your SQLC code.

---

## add-nosec

The `add-nosec` command automates the process of inserting `// #nosec` comments into your SQLC-generated files. This is particularly useful when gosec flags parts of the generated code—such as query constants—for hardcoded credentials. After you've verified that no sensitive credentials are hardcoded in your code, you can safely run `add-nosec` to suppress these false-positive warnings. This tool is especially beneficial in CI/CD pipelines where running gosec is part of your pull request checks, helping avoid test failures due to false flags.

### add-nosec Overview

- **Purpose:**  
  Automatically add `// #nosec` comments to specified constant declarations in your SQLC-generated files, ensuring that gosec warnings (e.g., for hardcoded credentials) are suppressed.

- **Why It's Needed:**  
  SQLC-generated query code can sometimes be flagged by gosec as containing hardcoded credentials—even though these warnings are false positives because the queries don't contain actual sensitive information. Since every run of `sqlc generate` overwrites your files, any manual changes to suppress these warnings are lost. This command re-applies the necessary annotations as part of your automated workflow.

- **Method:**  
  Leveraging Go’s AST-based processing, the tool locates constant declarations (e.g., for SQL queries) and appends a `// #nosec` comment. This ensures that only the intended sections are annotated, reducing the risk of inadvertently ignoring other security issues.

### add-nosec Flags

- **`-queryglob`**  
  - **Description:** Specifies a glob pattern that matches all the SQLC-generated files you want to process.  
  - **Example:**  

    ```bash
    -queryglob="internal/database/*.sql.go"
    ```

- **`-target`**  
  - **Description:** A comma-separated list of constant names (or identifiers) that should have `// #nosec` appended to their declarations.  
  - **Example:**  

    ```bash
    -target=createRefreshToken,revokeToken
    ```

### add-nosec Usage Example

After verifying that your generated queries do not contain actual hardcoded credentials, run:

```bash
sqlc-qol add-nosec -queryglob="internal/database/*.sql.go" -target=createRefreshToken,revokeToken
```

In this example, the command will:

- Traverse all files matching the glob pattern (e.g., all `.sql.go` files in the `internal/database/` directory).
- Search for constant declarations named `createRefreshToken` and `revokeToken` in each file.
- Append `// #nosec` to these constants to suppress gosec warnings.

### Example of a Processed File

Before running `add-nosec`, your SQLC-generated code might look like this:

```go
const revokeToken = `-- name: RevokeToken :exec
UPDATE refresh_tokens
SET revoked_at = ?1
WHERE user_id = ?2
    AND device_info_id = ?3
    AND revoked_at IS NULL
`
```

After processing, the tool automatically appends the comment, resulting in:

```go
const revokeToken = `-- name: RevokeToken :exec
UPDATE refresh_tokens
SET revoked_at = ?1
WHERE user_id = ?2
    AND device_info_id = ?3
    AND revoked_at IS NULL
` // #nosec
```

### add-nosec Workflow Integration

Because `sqlc generate` overwrites your query files each time it's run, manual fixes for gosec warnings are unsustainable. Integrate `add-nosec` into your build or post-generation script so that after every SQLC generation, your files are automatically updated. For example, add the following target to your Makefile:

```makefile
.PHONY: generate
generate:
  sqlc generate
  sqlc-qol add-nosec -queryglob="internal/database/*.sql.go" -target=createRefreshToken,revokeToken
```

This integration ensures that your codebase remains modular and consistent—automatically suppressing false-positive gosec warnings every time you generate new code.

### Summary

The `add-nosec` command is an essential tool for maintaining security annotations in a dynamic, auto-generated codebase. By automating the insertion of `// #nosec` comments across all relevant SQLC-generated files (using a glob pattern), it saves you from manual intervention and ensures that, once you’ve verified your SQLC-generated queries are safe, your files remain compliant with your security standards—even after each run of `sqlc generate`.

---
