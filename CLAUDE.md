# Jenkins CLI - Agent Guide

## Quick Reference

- **Binary:** `jenkins`
- **Config file:** `~/.config/jenkins-cli/config.yaml`
- **Env vars:** `JENKINS_URL`, `JENKINS_USER`, `JENKINS_TOKEN`, `JENKINS_INSECURE`
- **Auth method:** Username + API token (not password; generate at `<jenkins-url>/user/<username>/configure`)
- **Config priority:** CLI flags > environment variables > profile config

## Setup

```bash
# Interactive login (prompts for URL, username, API token, profile name)
jenkins login

# Or set environment variables for non-interactive use
export JENKINS_URL=https://jenkins.example.com
export JENKINS_USER=admin
export JENKINS_TOKEN=11xxxxxxxxxxxxxxxxxx
```

## Job Path Notation

Jenkins organizes jobs in folders using slash-separated paths:

```
my-job                        # root-level job
my-folder/my-job              # job inside a folder
team/project/deploy-pipeline  # nested folders
```

All job commands accept this path format. Use `jenkins job list --recursive` to discover paths.

## Output Formats

All list/get commands support three output formats via `-o`:

- `-o table` (default) -- human-readable tabular output
- `-o json` -- JSON, ideal for programmatic parsing with jq
- `-o yaml` -- YAML, useful for config management

**For agents:** Always use `-o json` when you need to parse or process output programmatically.

## Common Workflows

### Check server connectivity

```bash
# Show server status (version, mode, executors, etc.)
jenkins status -o json

# Show current authenticated user
jenkins whoami -o json
```

### List and inspect jobs

```bash
# List all root-level jobs
jenkins job list -o json

# List jobs in a specific folder
jenkins job list --folder my-team -o json

# List all jobs recursively across all folders
jenkins job list --recursive -o json

# Filter by status (SUCCESS, FAILURE, UNSTABLE, DISABLED, ABORTED, NOT_BUILT, RUNNING)
jenkins job list --recursive --status FAILURE -o json

# Get detailed info about a specific job
jenkins job get my-folder/deploy-pipeline -o json

# Get the raw config.xml for a job
jenkins job config my-folder/deploy-pipeline
```

### Trigger a build, wait, and follow logs

```bash
# Simple trigger (fire and forget)
jenkins job build my-folder/deploy-pipeline

# Trigger with parameters
jenkins job build my-folder/deploy-pipeline --param BRANCH=main --param ENV=staging

# Trigger and wait for completion
jenkins job build my-folder/deploy-pipeline --wait

# Trigger and stream console output in real time
jenkins job build my-folder/deploy-pipeline --follow

# Full workflow: trigger with params, wait, and stream output
jenkins job build my-folder/deploy-pipeline --param BRANCH=main --wait --follow

# Set a custom timeout (default is 30m)
jenkins job build my-folder/deploy-pipeline --wait --timeout 1h
```

### View build details and logs

```bash
# List builds for a job
jenkins build list my-folder/deploy-pipeline -o json

# Get details for a specific build
jenkins build get my-folder/deploy-pipeline 42 -o json

# View the full console log
jenkins build log my-folder/deploy-pipeline 42

# Stream the log of a running build in real time
jenkins build log my-folder/deploy-pipeline 42 --follow

# View pipeline stage breakdown
jenkins build stages my-folder/deploy-pipeline 42 -o json

# View test results
jenkins build test-report my-folder/deploy-pipeline 42 -o json

# List build artifacts
jenkins build artifacts my-folder/deploy-pipeline 42 -o json

# View build environment variables
jenkins build env my-folder/deploy-pipeline 42 -o json

# Stop a running build
jenkins build stop my-folder/deploy-pipeline 42

# Replay a pipeline build
jenkins build replay my-folder/deploy-pipeline 42

# Open a build in the browser
jenkins build open my-folder/deploy-pipeline 42
```

### Manage the build queue

```bash
# List items waiting in the build queue
jenkins queue list -o json

# Cancel a pending build by queue ID
jenkins queue cancel <queue-id>
```

### Pipeline validation and input handling

```bash
# Validate a declarative Jenkinsfile (scripted pipelines not supported)
jenkins pipeline validate --from-file Jenkinsfile

# List pending input actions for a pipeline build
jenkins pipeline input-list my-pipeline 42 -o json

# Submit (proceed with) a pending input
jenkins pipeline input-submit my-pipeline 42 <input-id>

# Submit with parameters
jenkins pipeline input-submit my-pipeline 42 deploy-approval --param APPROVE=yes --param ENV=prod

# Abort a pending input
jenkins pipeline input-abort my-pipeline 42 <input-id>
```

### Credential management

```bash
# List all system credentials (default store=system, domain=_ for global)
jenkins credential list -o json

# List credentials in a specific store and domain
jenkins credential list --store system --domain my-domain -o json

# Filter by credential type (case-insensitive substring match)
jenkins credential list --type "SSH" -o json

# Get details about a specific credential
jenkins credential get <credential-id> -o json
jenkins credential get <credential-id> --store system --domain _ -o json

# Create a credential from XML config
jenkins credential create --store system --domain _ -f credential.xml

# Update a credential
jenkins credential update <credential-id> --store system --domain _ -f credential.xml

# Delete a credential
jenkins credential delete <credential-id> --store system --domain _
```

### Node/agent management

```bash
# List all nodes
jenkins node list -o json

# List only offline or online nodes
jenkins node list --offline -o json
jenkins node list --online -o json

# Get details about a specific node
jenkins node get my-agent -o json

# Create a new permanent agent node
jenkins node create -f node.xml

# Take a node offline with a reason
jenkins node disable my-agent --message "Maintenance window"

# Bring a node back online
jenkins node enable my-agent

# View node agent log
jenkins node log my-agent

# Delete a node
jenkins node delete my-agent
```

### Plugin management

```bash
# List installed plugins
jenkins plugin list -o json

# List only active/enabled plugins
jenkins plugin list --active -o json
jenkins plugin list --enabled -o json

# Get details about a specific plugin
jenkins plugin get <plugin-short-name> -o json

# Install a plugin by name
jenkins plugin install <plugin-short-name>

# Uninstall a plugin
jenkins plugin uninstall <plugin-short-name>

# Check for available plugin updates
jenkins plugin check-updates -o json
```

### View management

```bash
# List all views
jenkins view list -o json

# Get view details including its jobs
jenkins view get <view-name> -o json

# Create a new view
jenkins view create <view-name>

# Add a job to a view
jenkins view add-job <view-name> <job-path>

# Remove a job from a view
jenkins view remove-job <view-name> <job-path>

# Delete a view
jenkins view delete <view-name>
```

### User management

```bash
# List all known users
jenkins user list -o json

# Get details about a specific user
jenkins user get <user-id> -o json
```

### System administration

```bash
# Show Jenkins system info (version, mode, executors, etc.)
jenkins system info -o json

# Restart Jenkins (graceful -- waits for running builds)
jenkins system restart --safe

# Restart Jenkins immediately
jenkins system restart

# Enter quiet-down mode (no new builds start)
jenkins system quiet-down

# Cancel quiet-down mode
jenkins system cancel-quiet-down

# Execute a Groovy script on the Jenkins controller
jenkins system run-script --script 'println Jenkins.instance.numExecutors'

# Execute a Groovy script from a file
jenkins system run-script --from-file maintenance-script.groovy
```

### Job lifecycle management

```bash
# Create a new job from XML config
jenkins job create my-folder/new-pipeline -f config.xml

# Copy an existing job
jenkins job copy my-folder/source-pipeline my-folder/target-pipeline

# Rename a job
jenkins job rename my-folder/old-name my-folder/new-name

# Disable a job
jenkins job disable my-folder/deploy-pipeline

# Enable a disabled job
jenkins job enable my-folder/deploy-pipeline

# Update a job's XML config
jenkins job update my-folder/deploy-pipeline -f config.xml

# Wipe the workspace directory
jenkins job wipe-workspace my-folder/deploy-pipeline

# Delete a job permanently
jenkins job delete my-folder/deploy-pipeline
```

## Tips for Agents

- Always use `-o json` when you need to parse output programmatically.
- Job paths use `/` as separator for folders: `team/project/pipeline`.
- Use `jenkins job list --recursive -o json` to discover all job paths across all folders.
- For CI/CD workflows, the typical pattern is: `job build --wait --follow --param KEY=VALUE`.
- The `--follow` flag on `job build` streams console output in real time -- great for monitoring.
- The `--timeout` flag on `job build` defaults to 30 minutes; increase it for long-running builds.
- Build numbers are integers. Use `build list <job> -o json` to find recent build numbers.
- Credentials use `--store` (default: `system`) and `--domain` (default: `_` for global).
- Credential configs and job configs use Jenkins XML format, not JSON.
- `pipeline validate` only works with declarative pipelines, not scripted pipelines.
- `system run-script` executes Groovy with full controller access -- use with caution.
- The CLI supports multiple profiles via `--profile` flag. Profiles are created during `jenkins login`.

## Common CI/CD Agent Patterns

### Trigger build, wait for result, check status

```bash
# Trigger and wait
jenkins job build team/deploy-pipeline --param BRANCH=main --wait -o json

# If you need to check the latest build result separately
jenkins build list team/deploy-pipeline --limit 1 -o json
jenkins build get team/deploy-pipeline <build-number> -o json
```

### Find failing jobs across the instance

```bash
jenkins job list --recursive --status FAILURE -o json
```

### Handle pipeline approval gates

```bash
# Check for pending inputs
jenkins pipeline input-list team/deploy-pipeline 42 -o json

# Approve the deployment
jenkins pipeline input-submit team/deploy-pipeline 42 deploy-approval --param APPROVE=yes
```

## Complete Command Reference

### Top-level commands

| Command | Description |
|---------|-------------|
| `jenkins login` | Interactively authenticate with a Jenkins server |
| `jenkins status` | Show Jenkins server status |
| `jenkins whoami` | Show current authenticated user |
| `jenkins version` | Print CLI version |
| `jenkins update` | Check for and install CLI updates |

### `jenkins job` (alias: `jobs`) -- Manage jobs

| Command | Description |
|---------|-------------|
| `jenkins job list` | List jobs (--folder, --recursive, --status) |
| `jenkins job get <path>` | Get detailed info about a job |
| `jenkins job create <path> -f <xml>` | Create a new job from XML config |
| `jenkins job update <path> -f <xml>` | Update a job's XML config |
| `jenkins job copy <source> <target>` | Copy an existing job |
| `jenkins job rename <old> <new>` | Rename a job |
| `jenkins job delete <path>` | Permanently delete a job |
| `jenkins job enable <path>` | Enable a disabled job |
| `jenkins job disable <path>` | Disable a job |
| `jenkins job config <path>` | Retrieve the raw config.xml |
| `jenkins job wipe-workspace <path>` | Wipe the workspace directory |
| `jenkins job build <path>` | Trigger a build (--param, --wait, --follow, --timeout) |

### `jenkins build` (alias: `builds`) -- Manage builds

| Command | Description |
|---------|-------------|
| `jenkins build list <path>` | List builds for a job (--status) |
| `jenkins build get <path> <number>` | Get detailed info about a build |
| `jenkins build log <path> <number>` | View console output (--follow) |
| `jenkins build stop <path> <number>` | Stop a running build |
| `jenkins build delete <path> <number>` | Delete a build record |
| `jenkins build artifacts <path> <number>` | List or download build artifacts |
| `jenkins build test-report <path> <number>` | View test results |
| `jenkins build env <path> <number>` | View injected environment variables |
| `jenkins build stages <path> <number>` | View pipeline stage breakdown |
| `jenkins build replay <path> <number>` | Replay a pipeline build |
| `jenkins build open <path> <number>` | Open a build in the browser |

### `jenkins queue` -- Manage the build queue

| Command | Description |
|---------|-------------|
| `jenkins queue list` | List items in the build queue |
| `jenkins queue cancel <queue-id>` | Cancel a pending build |

### `jenkins pipeline` -- Pipeline operations

| Command | Description |
|---------|-------------|
| `jenkins pipeline validate` | Validate a declarative Jenkinsfile (--from-file) |
| `jenkins pipeline input-list <path> <number>` | List pending input actions |
| `jenkins pipeline input-submit <path> <number> <id>` | Proceed with a pending input (--param) |
| `jenkins pipeline input-abort <path> <number> <id>` | Abort a pending input |

### `jenkins credential` (aliases: `credentials`, `cred`, `creds`) -- Manage credentials

| Command | Description |
|---------|-------------|
| `jenkins credential list` | List credentials (--store, --domain, --type) |
| `jenkins credential get <id>` | Get credential details (--store, --domain) |
| `jenkins credential create -f <xml>` | Create a credential (--store, --domain) |
| `jenkins credential update <id> -f <xml>` | Update a credential (--store, --domain) |
| `jenkins credential delete <id>` | Delete a credential (--store, --domain) |

### `jenkins node` (aliases: `nodes`, `agent`, `agents`) -- Manage nodes/agents

| Command | Description |
|---------|-------------|
| `jenkins node list` | List all nodes (--offline, --online) |
| `jenkins node get <name>` | Get node details |
| `jenkins node create -f <xml>` | Create a permanent agent node |
| `jenkins node delete <name>` | Delete a node |
| `jenkins node enable <name>` | Bring a node online |
| `jenkins node disable <name>` | Take a node offline (--message) |
| `jenkins node log <name>` | View agent log |

### `jenkins plugin` (alias: `plugins`) -- Manage plugins

| Command | Description |
|---------|-------------|
| `jenkins plugin list` | List installed plugins (--active, --enabled) |
| `jenkins plugin get <name>` | Get plugin details |
| `jenkins plugin install <name>` | Install a plugin |
| `jenkins plugin uninstall <name>` | Uninstall a plugin |
| `jenkins plugin check-updates` | Check for plugin updates |

### `jenkins view` (alias: `views`) -- Manage views

| Command | Description |
|---------|-------------|
| `jenkins view list` | List all views |
| `jenkins view get <name>` | Get view details including jobs |
| `jenkins view create <name>` | Create a new view |
| `jenkins view delete <name>` | Delete a view |
| `jenkins view add-job <view> <job>` | Add a job to a view |
| `jenkins view remove-job <view> <job>` | Remove a job from a view |

### `jenkins user` (alias: `users`) -- Manage users

| Command | Description |
|---------|-------------|
| `jenkins user list` | List all known users |
| `jenkins user get <id>` | Get user details |

### `jenkins system` -- System administration

| Command | Description |
|---------|-------------|
| `jenkins system info` | Show server info (version, mode, executors) |
| `jenkins system restart` | Restart Jenkins (--safe for graceful) |
| `jenkins system quiet-down` | Enter quiet-down mode |
| `jenkins system cancel-quiet-down` | Exit quiet-down mode |
| `jenkins system run-script` | Execute a Groovy script (--script or --from-file) |

## Global Flags

| Flag | Description |
|------|-------------|
| `-o, --output <format>` | Output format: table (default), json, yaml |
| `--profile <name>` | Configuration profile to use |
| `-s, --server <url>` | Jenkins server URL override |
| `-u, --user <name>` | Jenkins username override |
| `-t, --token <token>` | Jenkins API token override |
| `-k, --insecure` | Skip TLS certificate verification |
| `--no-color` | Disable color output |
| `-v, --verbose` | Enable verbose output |
