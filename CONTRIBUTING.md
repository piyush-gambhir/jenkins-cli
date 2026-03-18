# Contributing to Jenkins CLI

Thank you for your interest in contributing! This guide will help you get started.

## Development Setup

### Prerequisites

- Go 1.22 or later
- Make
- Git

### Clone and Build

```bash
git clone https://github.com/piyush-gambhir/jenkins-cli.git
cd jenkins-cli
make build
```

### Run Locally

```bash
./bin/jenkins --help
./bin/jenkins version
```

### Run Tests

```bash
make test
```

### Lint

```bash
make lint    # requires golangci-lint
make vet     # go vet
make fmt     # gofmt
```

## Project Structure

```
.
├── main.go                 # Entry point
├── cmd/                    # Cobra command definitions (flat layout)
│   ├── root.go             # Root command, global flags
│   ├── login.go            # Auth commands
│   ├── status.go           # Server status
│   ├── whoami.go           # Current user
│   ├── job.go              # Job parent command
│   ├── job_list.go         # Job list subcommand
│   ├── job_get.go          # Job get subcommand
│   ├── job_build.go        # Trigger builds (--param, --wait, --follow)
│   ├── job_config.go       # Get raw config.xml
│   ├── build.go            # Build parent command
│   ├── build_list.go       # Build list subcommand
│   ├── build_log.go        # Console output (--follow for streaming)
│   ├── build_stages.go     # Pipeline stage breakdown
│   ├── credential.go       # Credential parent command
│   ├── credential_list.go  # Credential CRUD
│   ├── node.go             # Node/agent commands
│   ├── plugin.go           # Plugin management
│   ├── view.go             # View management
│   ├── pipeline.go         # Jenkinsfile validation, input handling
│   ├── system.go           # System admin (restart, groovy scripts)
│   └── ...
├── internal/
│   ├── client/             # HTTP API client
│   │   ├── client.go       # Base client (auth, headers, errors)
│   │   ├── crumb.go        # CSRF crumb token handling
│   │   ├── jobs.go         # Job API methods
│   │   ├── builds.go       # Build API methods
│   │   ├── credentials.go  # Credential API methods
│   │   ├── nodes.go        # Node API methods
│   │   ├── plugins.go      # Plugin API methods
│   │   ├── views.go        # View API methods
│   │   ├── pipeline.go     # Pipeline validation and input actions
│   │   ├── system.go       # System admin API methods
│   │   ├── queue.go        # Build queue API methods
│   │   ├── users.go        # User API methods
│   │   ├── request.go      # HTTP request helpers
│   │   └── errors.go       # Error types
│   ├── config/             # Config file and auth resolution
│   ├── output/             # JSON/YAML/Table formatters
│   ├── path/               # Job path translation (slash notation to Jenkins API URL segments)
│   ├── version/            # Build version info
│   └── update/             # Self-update logic
├── Makefile
├── .goreleaser.yaml
└── .github/workflows/
    ├── ci.yml              # Build + test on every push/PR
    └── release.yml         # GoReleaser on tag push
```

### Flat Command Layout

Unlike many Cobra projects that use subdirectories, this project uses a flat file layout in `cmd/`. Each file is named `<resource>_<action>.go` (e.g., `job_list.go`, `build_log.go`). The parent command file (e.g., `job.go`) registers all subcommands.

### CSRF Crumb Handling

Jenkins requires a CSRF crumb token for all state-changing (POST) requests. The client in `internal/client/crumb.go` automatically fetches and attaches crumb tokens. If you add a new write operation, the crumb is handled transparently by the base request layer.

### Job Path Translation

Jenkins encodes folder paths as `/job/folder/job/subfolder/job/name` in its API URLs. The `internal/path/` package translates the user-friendly slash notation (`team/project/pipeline`) to the Jenkins API URL format. Always use this package when constructing job URLs.

## Adding a New Command

1. **Add the API method** in `internal/client/<resource>.go`:
   ```go
   func (c *Client) ListWidgets(params ...) ([]Widget, error) {
       // HTTP call to the Jenkins API
   }
   ```

2. **Create the command** in `cmd/<resource>_list.go`:
   ```go
   func newJobListCmd() *cobra.Command {
       // Define flags, run function, help text with examples
   }
   ```

3. **Register** the command in the parent command file (e.g., `cmd/<resource>.go`) inside its `init()` or constructor function.

4. **Add a test** in the corresponding `_test.go` file using `httptest.NewServer`.

5. **Update documentation**:
   - Add a `Long` description with examples to the command
   - Update `README.md` with the new command
   - Update `CLAUDE.md` if it's a commonly-used command
   - Update the skill's `references/commands.md`

## Code Style

- Follow standard Go conventions (`gofmt`, `go vet`)
- Use meaningful variable names
- Every command must have:
  - `Short` description (one line)
  - `Long` description with usage examples
  - Proper flag definitions with descriptions
- Use `-o json` output in all examples for agent-friendliness
- Table output should have meaningful column headers

## Commit Messages

Follow conventional commits:
```
feat: add widget list command
fix: correct pagination in dashboard search
docs: update README with new alert commands
test: add tests for credential CRUD
chore: update dependencies
```

## Pull Requests

1. Fork the repo and create a feature branch
2. Make your changes with tests
3. Run `make test` and `make vet` to ensure everything passes
4. Commit with a clear message
5. Open a PR against `main`

## Releasing

Releases are automated via GoReleaser. To create a release:

```bash
git tag v0.2.0
git push origin v0.2.0
```

This triggers GitHub Actions to:
1. Build binaries for all platforms
2. Create a GitHub Release with assets
3. Generate a changelog

## Reporting Issues

- Use GitHub Issues
- Include: CLI version (`jenkins version`), OS/arch, command that failed, error output
- For feature requests, describe the use case

## License

This project is licensed under the MIT License — see [LICENSE](LICENSE) for details.
