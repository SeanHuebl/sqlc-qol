# sqlc-qol
CLI Tool to improve the quality of life for devs using SQLC 
## Installation:

You have several options to install **sqlc-qol**:

### Installing via Go Install (recommended)

If you have Go installed (version 1.16 or higher is recommended), you can install **sqlc-qol** directly from the source repository using:

```bash
go install github.com/yourusername/sqlc-qol@latest
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
git clone https://github.com/yourusername/sqlc-qol.git
cd sqlc-qol
go build -o sqlc-qol
```

This creates an executable named `sqlc-qol` that you can run.

### Verifying Your Installation

After installation, verify that **sqlc-qol** is correctly installed by running:

```bash
sqlc-qol --help
```

## Current Features:
  ### Post Processing:
  - [qualify-models](#qualify-models)
  - addnosec
  
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
