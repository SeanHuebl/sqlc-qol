# sqlc-qol
CLI Tool to improve the quality of life for devs, who are using SQLC, with automation.\
Contribution is welcome if you would like to submit any other features!

# Table of Contents:
- [Installating the CLI Tool](#installation)
- [qualify-models](#qualify-models)
- [add-nosec](#add-nosec)

## Directory Tree:
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

## Installation:

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
 
## qualify-models:
The `qualify-models` command is designed to streamline your SQLC workflow when you decouple your models from your generated query files. After you move SQLC-generated models into a dedicated external models package (e.g., `internal/models`) for enhanced modularity and decoupling, the SQLC-generated query files still reference model types without a package qualifier. Moreover, every time you run `sqlc generate`, these query files are overwritten, meaning any manual modifications are lost.

This command automates the process by:

- **Extracting Model Names:**  
  It parses your specified models file (e.g., `internal/models/database.go`) to collect all struct names defined in your external models package.

- **Qualifying Model References:**  
  It scans your SQLC-generated query files (using a glob pattern like `internal/database/*.sql.go`) and rewrites any bare model identifiers by prefixing them with your models package name (e.g., converting `Transaction` to `models.Transaction`). This ensures that all references explicitly point to your dedicated models package.

- **Adding Necessary Imports:**  
  If the query file does not already import your models package, the command automatically inserts the required import statement (e.g., `import "internal/models"`).

### Flags:

Below is an explanation of each flag used in the **qualify-models** command:

- **`-model`**  
  This flag specifies the path to the file that contains your external models.  
  **Purpose:** The tool parses this file to extract all the struct names (model types) defined in your dedicated models package.  
  **Example:** If your models are located in `internal/models/database.go`, you would pass:  
  ```bash
  -model=internal/models/database.go
  ```

- **`-queryglob`**  
  This flag defines a glob pattern that matches all the SQLC-generated query files you want to process.  
  **Purpose:** Since SQLC generates one file per table (or query group), this pattern tells the tool which files to search for unqualified model references.  
  **Example:** If your generated files are in `internal/database/` and all end with `.sql.go`, you might use:  
  ```bash
  -queryglob="internal/database/*.sql.go"
  ```

- **`-modelimport`**  
  This flag specifies the import path for your external models package.  
  **Purpose:** When the tool qualifies model references (e.g., turning `Transaction` into `models.Transaction`), it also checks and, if needed, inserts the correct import statement in the query files.  
  **Example:** If your models package is `internal/models`, then you would pass:  
  ```bash
  -modelimport=internal/models
  ```

### How These Flags Work Together:

When you run the command, the tool performs the following steps:

1. **Model Extraction:**  
   It reads the file specified by `-model` (e.g., `internal/models/database.go`) to collect all struct names that represent your models.

2. **File Processing:**  
   It uses the `-queryglob` pattern to identify all SQLC-generated query files. For each file, it searches for bare model references.

3. **Qualifying References and Adding Imports:**  
   For every unqualified model identifier found (e.g., `Transaction`), the tool rewrites it to use the fully qualified form (e.g., `models.Transaction`) and makes sure the file has an import for the models package specified by `-modelimport`.

This integration ensures that after every `sqlc generate` run, your query files remain consistent with your decoupled, external models package—automatically maintaining a modular and scalable codebase.

### **Why Use qualify-models?**

Because `sqlc generate` overwrites your query files each time it's run, manual adjustments are not sustainable. Integrating `qualify-models` into your workflow script ensures that every new generation of code is automatically updated to maintain the decoupled, modular structure of your project. This saves time, reduces errors, and keeps your codebase consistent.

**Usage Example:**

```bash
sqlc-qol qualify-models -model=internal/models/database.go \
                        -queryglob="internal/database/*.sql.go" \
                        -modelimport=internal/models
```

In this example, the command:
- Reads model definitions from `internal/models/database.go`.
- Processes all SQLC-generated files in `internal/database/` matching the `.sql.go` pattern.
- Rewrites the files to qualify model references with `models.` and adds the `internal/models` import if it's missing.

Integrate this tool into your build or post-generation script to ensure your codebase remains clean and modular every time you run `sqlc generate`.

---

## add-nosec:

The `add-nosec` command automates the process of inserting `// #nosec` comments into your SQLC-generated files. This is particularly useful when gosec flags parts of the generated code—such as query constants—for hardcoded credentials. After you've verified that no sensitive credentials are hardcoded in your code, you can safely run `add-nosec` to suppress these false-positive warnings. The usefullness also extends to CI/CD pipelines where running gosec is part of the pull request actions, and will avoid the test failing due to false flagging.

### Overview

- **Purpose:**  
  Automatically add `// #nosec` comments to specified constant declarations in your SQLC-generated files, ensuring that gosec warnings (e.g., for hardcoded credentials) are suppressed.

- **Why It's Needed:**  
  SQLC-generated query code can sometimes be flagged by gosec as containing hardcoded credentials—even though these warnings are false positives because the queries don't contain actual sensitive information. Since every run of `sqlc generate` overwrites your files, any manual changes to suppress these warnings are lost. This command re-applies the necessary annotations as part of your automated workflow.

- **Method:**  
  Leveraging Go’s AST-based processing, the tool locates constant declarations (e.g., for SQL queries) and appends a `// #nosec` comment. This ensures that only the intended sections are annotated, reducing the risk of inadvertently ignoring other security issues.

### Flags

- **`-file`**  
  - **Description:** Specifies the path to the SQLC-generated file you want to process.  
  - **Example:**  
    ```bash
    -file=internal/database/refresh_tokens.sql.go
    ```

- **`-target`**  
  - **Description:** A comma-separated list of constant names (or identifiers) that should have `// #nosec` appended to their declarations.  
  - **Example:**  
    ```bash
    -target=createRefreshToken,revokeToken
    ```

### Usage Example

After verifying that your generated queries do not contain actual hardcoded credentials, run:

```bash
sqlc-qol add-nosec -file=internal/database/refresh_tokens.sql.go -target=createRefreshToken,revokeToken
```

In this example, the command will:

- Open the file `internal/database/refresh_tokens.sql.go`.
- Search for constant declarations named `createRefreshToken` and `revokeToken`.
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

### Workflow Integration

Because `sqlc generate` overwrites your query files each time it's run, manual fixes for gosec warnings are unsustainable. Integrate `add-nosec` into your build or post-generation script to ensure that after every SQLC generation, your files are automatically updated. For example, add the following target to your Makefile:

```makefile
.PHONY: generate
generate:
	sqlc generate
	sqlc-qol add-nosec -file=internal/database/refresh_tokens.sql.go -target=createRefreshToken,revokeToken
```

This integration ensures that your codebase remains modular and consistent—automatically suppressing false-positive gosec warnings every time you generate new code.

### Summary

The `add-nosec` command is an essential tool for maintaining security annotations in a dynamic, auto-generated codebase. By automating the insertion of `// #nosec` comments, it saves you from manual intervention and ensures that, once you’ve verified that your SQLC-generated queries are safe, your files remain compliant with your security standards—even after each run of `sqlc generate`.

---
