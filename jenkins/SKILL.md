---
name: jenkins
description: "Expert guide for using the jenkins CLI to manage Jenkins CI/CD servers. Use this skill whenever the user mentions Jenkins jobs, builds, pipelines, Jenkinsfile, build triggers, build logs, console output, Jenkins nodes, agents, executors, Jenkins plugins, Jenkins credentials, Jenkins views, build queue, build parameters, Jenkins system administration, or any Jenkins CI/CD automation. Also trigger when the user wants to trigger builds, check build status, stream build logs, manage Jenkins configuration, validate Jenkinsfiles, approve pipeline inputs, run Groovy scripts, or automate any Jenkins operations from the command line."
---

# Jenkins CLI -- Agent Skill Guide

Complete guide for coding agents to manage Jenkins CI/CD servers via the `jenkins` CLI. Covers jobs, builds, nodes, plugins, credentials, pipelines, views, and system administration.

## 1. Prerequisites & Setup

### Install

```bash
# Using go install
go install github.com/piyush-gambhir/jenkins-cli@latest

# Or from source
git clone https://github.com/piyush-gambhir/jenkins-cli.git
cd jenkins-cli && go build -o jenkins . && sudo mv jenkins /usr/local/bin/

# Or via install script
curl -sSL https://raw.githubusercontent.com/piyush-gambhir/jenkins-cli/main/install.sh | bash
```

### Authenticate

Jenkins uses **username + API token** (not password). Generate a token at `<jenkins-url>/user/<username>/configure`.

```bash
# Interactive login (saves a named profile)
jenkins login

# Or set environment variables for non-interactive / CI use
export JENKINS_URL=https://jenkins.example.com
export JENKINS_USER=admin
export JENKINS_TOKEN=11xxxxxxxxxxxxxxxxxx
```

Config priority: CLI flags > environment variables > profile config.

### Verify

```bash
jenkins status -o json    # Server version, mode, executors
jenkins whoami -o json     # Authenticated user info
```

## 2. Core Principles for Agents

- **ALWAYS use `-o json` for programmatic parsing.** All list/get commands support `-o json` and `-o yaml`. Default is `-o table` (human-readable). When you need to parse or process output, always pass `-o json` and pipe through `jq`.

- **Job paths use slash notation:** `folder/subfolder/job-name` (NOT Jenkins URL format like `job/folder/job/subfolder/job/name`). Use `jenkins job list --recursive -o json` to discover all paths.

- **The killer combo:** `jenkins job build <path> --follow` triggers a build AND streams console output in real time. This is the single most useful command for CI/CD automation.

- **For CI/CD automation:** `jenkins job build <path> --wait --timeout 30m` blocks until the build finishes and returns exit code 0 for SUCCESS, non-zero for FAILURE. Combine with `--follow` to also stream logs.

- **`--param KEY=VALUE` is repeatable** for parameterized builds. Pass as many as needed.

- **CSRF is handled automatically** -- no need to manually fetch crumbs.

- **Destructive operations require `--confirm`** -- deleting jobs, builds, nodes, views, plugins, and credentials all require the `--confirm` flag.

- **Credentials use `--store` and `--domain`** -- defaults are `system` and `_` (global). Credential and job configs use Jenkins XML format, not JSON.

- **`pipeline validate` only works with declarative pipelines** -- scripted pipelines cannot be validated.

- **`system run-script` executes Groovy with full controller access** -- use with caution; it can do anything on the Jenkins instance.

- **Multiple profiles** are supported via `--profile <name>`. Create profiles with `jenkins login --name <profile-name>`.

## 3. Common Workflows

### Trigger a build and watch it

```bash
jenkins job build my-pipeline --follow
```

### Trigger a parameterized build, wait for result

```bash
jenkins job build deploy/production --param BRANCH=main --param ENV=prod --wait --timeout 30m
echo "Exit code: $?"  # 0=SUCCESS, non-zero=FAILURE
```

### Full workflow: params + wait + stream logs

```bash
jenkins job build my-folder/deploy-pipeline --param BRANCH=main --param VERSION=1.2.3 --wait --follow --timeout 1h
```

### Check why a build failed

```bash
# Get last build info
jenkins build get my-job 42 -o json | jq '{result, timestamp, duration}'

# View the console log
jenkins build log my-job 42

# Check which pipeline stage failed
jenkins build stages my-job 42 -o json

# Check test results
jenkins build test-report my-job 42 -o json
```

### Find all failing jobs

```bash
jenkins job list --recursive --status FAILURE -o json
```

### Find all failing jobs (alternative with jq)

```bash
jenkins job list --recursive -o json | jq '.[] | select(.color | test("red"))'
```

### Manage jobs in folders

```bash
jenkins job list --folder team-alpha --recursive -o json
jenkins job get team-alpha/deploy/staging -o json
jenkins job build team-alpha/deploy/staging --param ENV=staging --follow
```

### Export and import a job config

```bash
# Export
jenkins job config my-job > my-job-config.xml

# Import to a new job
jenkins job create my-new-job --from-file my-job-config.xml

# Update an existing job
jenkins job update my-job --from-file my-job-config.xml
```

### Copy and rename jobs

```bash
jenkins job copy my-pipeline my-pipeline-copy
jenkins job rename my-folder/old-name new-name
```

### Enable and disable jobs

```bash
jenkins job disable my-folder/deploy-pipeline
jenkins job enable my-folder/deploy-pipeline
```

### List builds and filter by status

```bash
jenkins build list my-pipeline -o json
jenkins build list my-pipeline --status FAILURE --limit 10 -o json
jenkins build list my-pipeline --status SUCCESS --limit 5 -o json
```

### Stream a running build's log

```bash
jenkins build log my-pipeline 42 --follow
```

### Download build artifacts

```bash
jenkins build artifacts my-pipeline 42 -o json
jenkins build artifacts my-pipeline 42 --download --output-dir ./artifacts
```

### View build environment variables

```bash
jenkins build env my-pipeline 42 -o json
```

### Stop a running build

```bash
jenkins build stop my-pipeline 42
```

### Replay a pipeline build

```bash
jenkins build replay my-pipeline 42
```

### Check the build queue

```bash
jenkins queue list -o json
jenkins queue cancel <queue-id>
```

### Approve a pipeline input

```bash
# List pending inputs
jenkins pipeline input-list my-pipeline 42 -o json

# Submit approval with parameters
jenkins pipeline input-submit my-pipeline 42 deploy-approval --param APPROVE=yes --param ENV=prod

# Or abort the input
jenkins pipeline input-abort my-pipeline 42 deploy-approval
```

### Validate a Jenkinsfile before committing

```bash
jenkins pipeline validate --from-file Jenkinsfile
```

### Check node status

```bash
jenkins node list -o json | jq '.[] | {name: .displayName, offline, idle: (.numExecutors - .busyExecutors)}'
```

### Manage nodes for maintenance

```bash
# Take offline
jenkins node disable my-agent --message "Maintenance window"

# Bring back online
jenkins node enable my-agent

# Check agent log
jenkins node log my-agent
```

### Create a new agent node

```bash
jenkins node create build-agent --remote-fs /opt/jenkins --executors 2 --labels "linux docker"
```

### List offline nodes

```bash
jenkins node list --offline -o json
```

### Manage credentials

```bash
# List all global credentials
jenkins credential list -o json

# List credentials filtered by type
jenkins credential list --type "SSH" -o json

# Get a specific credential
jenkins credential get my-ssh-key -o json

# Create from XML
jenkins credential create --from-file cred.xml --store system --domain _

# Update
jenkins credential update my-cred-id --from-file updated-cred.xml --store system --domain _

# Delete
jenkins credential delete my-cred-id --store system --domain _ --confirm
```

### Manage plugins

```bash
# List installed plugins
jenkins plugin list -o json

# List only active plugins
jenkins plugin list --active -o json

# Install a plugin
jenkins plugin install git
jenkins plugin install blueocean --version 1.27.0

# Check for updates
jenkins plugin check-updates -o json

# Uninstall
jenkins plugin uninstall git --confirm
```

### Manage views

```bash
# List all views
jenkins view list -o json

# Get view details (including its jobs)
jenkins view get "My View" -o json

# Create a view and add jobs
jenkins view create "My Team"
jenkins view add-job "My Team" my-pipeline
jenkins view add-job "My Team" team/deploy-job

# Remove a job from a view
jenkins view remove-job "My Team" my-pipeline

# Delete a view (does not delete jobs)
jenkins view delete "My View" --confirm
```

### Run a Groovy script on the controller

```bash
# Inline script
jenkins system run-script --script 'println Jenkins.instance.numExecutors'

# From file
jenkins system run-script --from-file cleanup.groovy

# List all job names
jenkins system run-script --script 'Jenkins.instance.allItems.each { println it.fullName }'
```

### System administration

```bash
# Server info
jenkins system info -o json

# Safe restart (waits for running builds)
jenkins system restart --safe --confirm

# Immediate restart
jenkins system restart --confirm

# Enter quiet-down mode (no new builds)
jenkins system quiet-down

# Cancel quiet-down
jenkins system cancel-quiet-down
```

### List and inspect users

```bash
jenkins user list -o json
jenkins user get admin -o json
```

## 4. Command Reference (compact)

| Command Group | Key Commands |
|---|---|
| **Top-level** | `login`, `status`, `whoami`, `version`, `update` |
| **job** | `list`, `get`, `build`, `create`, `update`, `copy`, `rename`, `delete`, `enable`, `disable`, `config`, `wipe-workspace` |
| **build** | `list`, `get`, `log`, `stop`, `delete`, `artifacts`, `test-report`, `env`, `stages`, `replay`, `open` |
| **queue** | `list`, `cancel` |
| **pipeline** | `validate`, `input-list`, `input-submit`, `input-abort` |
| **credential** | `list`, `get`, `create`, `update`, `delete` |
| **node** | `list`, `get`, `create`, `delete`, `enable`, `disable`, `log` |
| **plugin** | `list`, `get`, `install`, `uninstall`, `check-updates` |
| **view** | `list`, `get`, `create`, `delete`, `add-job`, `remove-job` |
| **user** | `list`, `get` |
| **system** | `info`, `restart`, `quiet-down`, `cancel-quiet-down`, `run-script` |

### Global flags (available on all commands)

| Flag | Description |
|---|---|
| `-o, --output <fmt>` | Output format: `table` (default), `json`, `yaml` |
| `--profile <name>` | Configuration profile to use |
| `-s, --server <url>` | Jenkins server URL override |
| `-u, --user <name>` | Jenkins username override |
| `-t, --token <token>` | Jenkins API token override |
| `-k, --insecure` | Skip TLS certificate verification |
| `--no-color` | Disable color output |
| `-v, --verbose` | Enable verbose output |

### Key flags by command

| Command | Notable Flags |
|---|---|
| `job list` | `--folder`, `--recursive`, `--status` |
| `job build` | `--param KEY=VALUE` (repeatable), `--wait`, `--follow`, `--timeout` |
| `job create` | `--from-file` (XML config), `--folder` |
| `job delete` | `--confirm` |
| `build list` | `--limit`, `--status` |
| `build log` | `--follow` |
| `build artifacts` | `--download`, `--output-dir` |
| `node create` | `--remote-fs`, `--executors`, `--labels` |
| `node disable` | `--message` |
| `credential list` | `--store`, `--domain`, `--type` |
| `credential create` | `--from-file`, `--store`, `--domain` |
| `plugin install` | `--version` |
| `pipeline validate` | `--from-file` |
| `pipeline input-submit` | `--param KEY=VALUE` (repeatable) |
| `system restart` | `--safe`, `--confirm` |
| `system run-script` | `--script`, `--from-file` |

See [references/commands.md](references/commands.md) for the full command reference with all flags and examples.

## 5. Troubleshooting

| Symptom | Cause / Fix |
|---|---|
| **403 Forbidden** | Check credentials. Ensure you are using an API token (not password). Verify the user has permissions. CSRF is handled automatically, but the server may have additional restrictions. |
| **401 Unauthorized** | API token may be expired or revoked. Re-generate at `<jenkins-url>/user/<username>/configure` and run `jenkins login` again. |
| **Connection timeout / refused** | Check `JENKINS_URL` or `--server` value. Verify Jenkins is running. For self-signed certs, use `--insecure` or `-k`. |
| **Build stuck in queue** | Run `jenkins queue list -o json` to see what is waiting and why (e.g., no available executors, blocked by other builds). Cancel with `jenkins queue cancel <id>`. |
| **"No such job"** | Check the job path. Use `jenkins job list --recursive -o json` to discover the correct path. Remember: paths use `/` (e.g., `team/project/job`), not the Jenkins URL format. |
| **Pipeline validation fails on scripted pipeline** | `pipeline validate` only supports declarative pipelines. Scripted pipelines cannot be validated via this endpoint. |
| **Groovy script errors** | Check syntax. `system run-script` runs with full controller access. Use `--from-file` for complex scripts to avoid shell escaping issues. |
| **Plugin install has no effect** | Some plugin installs require a Jenkins restart. Use `jenkins system restart --safe --confirm` after installing. |
| **Credential XML errors** | Credentials use Jenkins XML format. Check the XML structure matches the credential type. See the README for example XML templates. |
| **Wrong profile** | Specify `--profile <name>` explicitly. Check `~/.config/jenkins-cli/config.yaml` for available profiles. |

## 6. References

- Full command reference with all flags, arguments, and examples: [references/commands.md](references/commands.md)
- CLI README: [README.md](../README.md)
- Agent-optimized guide: [CLAUDE.md](../CLAUDE.md)
- Config file location: `~/.config/jenkins-cli/config.yaml`
- Jenkins API token generation: `<jenkins-url>/user/<username>/configure`
