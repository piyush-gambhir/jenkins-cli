# Jenkins CLI

A command-line interface for managing Jenkins CI/CD servers -- jobs, builds, nodes, plugins, credentials, pipelines, views, and system administration.

Designed for both human operators and coding agents (LLMs). All list and get commands support `-o json` and `-o yaml` for machine-readable output.

## Installation

### From Source

```bash
git clone https://github.com/piyush-gambhir/jenkins-cli.git
cd jenkins-cli
go build -o jenkins .
sudo mv jenkins /usr/local/bin/
```

### Using `go install`

```bash
go install github.com/piyush-gambhir/jenkins-cli@latest
```

### Using the Install Script

```bash
curl -sSL https://raw.githubusercontent.com/piyush-gambhir/jenkins-cli/main/install.sh | bash
```

### From GitHub Releases

Download the latest binary for your platform from [GitHub Releases](https://github.com/piyush-gambhir/jenkins-cli/releases) and place it in your `PATH`.

## Authentication

Jenkins uses **username + API token** (not your password). Generate an API token at `<jenkins-url>/user/<username>/configure` in the "API Token" section.

### Interactive Login

```bash
jenkins login
```

This prompts for:
1. Jenkins URL (e.g., `https://jenkins.example.com`)
2. Username
3. API Token
4. Profile name (default: `default`)

The connection is tested before saving.

### Environment Variables

```bash
export JENKINS_URL=https://jenkins.example.com
export JENKINS_USER=admin
export JENKINS_TOKEN=11a1b2c3d4e5f6...
```

### Command-Line Flags

```bash
jenkins status --server https://jenkins.example.com --user admin --token 11a1b2c3d4...
```

### Multiple Profiles

```bash
# Save different profiles
jenkins login --name production
jenkins login --name staging

# Use a specific profile
jenkins status --profile staging
jenkins job list --profile production
```

### Insecure Mode (Skip TLS Verification)

```bash
jenkins login --insecure
jenkins status --insecure
```

## Job Path Notation

Jenkins organizes jobs in folders. This CLI uses slash-separated paths:

```bash
jenkins job get my-job                     # root-level job
jenkins job get my-folder/my-job           # job in a folder
jenkins job get team/project/pipeline      # nested folders
```

This notation is used consistently across all commands that accept a `<job-path>` argument.

## Output Formats

All list and get commands support three output formats via the `-o` / `--output` flag:

| Flag       | Description                                    |
|------------|------------------------------------------------|
| (default)  | Human-readable table                           |
| `-o json`  | JSON (ideal for parsing with `jq` or by LLMs) |
| `-o yaml`  | YAML                                           |

```bash
jenkins job list -o json
jenkins build get my-pipeline 42 -o yaml
jenkins node list -o json
```

## Global Flags

These flags are available on every command:

| Flag                   | Description                        |
|------------------------|------------------------------------|
| `-o, --output <fmt>`   | Output format: table, json, yaml  |
| `--profile <name>`     | Configuration profile to use      |
| `-s, --server <url>`   | Jenkins server URL                |
| `-u, --user <name>`    | Jenkins username                  |
| `-t, --token <token>`  | Jenkins API token                 |
| `-k, --insecure`       | Skip TLS certificate verification |
| `--no-color`           | Disable color output              |
| `-v, --verbose`        | Verbose output                    |

## Commands

### login

Interactively authenticate with a Jenkins server.

```bash
jenkins login
jenkins login --name staging
```

### status

Show Jenkins server connection status and info.

```bash
jenkins status
jenkins status -o json
jenkins status --profile staging
```

### whoami

Show the currently authenticated user.

```bash
jenkins whoami
jenkins whoami -o json
```

---

### Job Commands

Manage Jenkins jobs (list, create, build, configure, etc.).

#### job list

List jobs at the root level or in a specific folder.

```bash
jenkins job list
jenkins job list --folder my-folder
jenkins job list --recursive
jenkins job list --folder my-team --recursive
jenkins job list --status FAILURE
jenkins job list --recursive --status SUCCESS
jenkins job list -o json
```

**Flags:**

| Flag                | Description                                                                 |
|---------------------|-----------------------------------------------------------------------------|
| `-f, --folder`      | Folder path to list jobs from                                              |
| `-r, --recursive`   | List jobs recursively through all subfolders                               |
| `--status`          | Filter by job status (SUCCESS, FAILURE, UNSTABLE, DISABLED, ABORTED, NOT_BUILT, RUNNING) |

#### job get

Get detailed information about a job.

```bash
jenkins job get my-pipeline
jenkins job get my-folder/my-pipeline
jenkins job get my-pipeline -o json
```

#### job build

Trigger a build for a job.

```bash
# Simple build
jenkins job build my-pipeline

# Parameterized build
jenkins job build my-pipeline --param BRANCH=main --param ENV=staging

# Wait for completion
jenkins job build my-pipeline --wait

# Stream console output in real-time
jenkins job build my-pipeline --follow

# Wait + follow + timeout
jenkins job build my-pipeline --param BRANCH=main --wait --follow --timeout 1h

# Build a job in a folder
jenkins job build my-folder/my-pipeline --param VERSION=1.2.3
```

**Flags:**

| Flag                  | Description                                          |
|-----------------------|------------------------------------------------------|
| `-p, --param`         | Build parameters as KEY=VALUE (repeatable)           |
| `-w, --wait`          | Wait for the build to complete                       |
| `-F, --follow`        | Follow (stream) the build console output             |
| `--timeout`           | Timeout for --wait/--follow (default: 30m)           |

#### job create

Create a new job from an XML configuration file.

```bash
jenkins job create my-new-job --from-file config.xml
jenkins job create my-new-job --from-file config.xml --folder my-folder
```

**Flags:**

| Flag             | Description                                |
|------------------|--------------------------------------------|
| `--from-file`    | Path to XML config file (required)         |
| `-f, --folder`   | Folder to create the job in                |

#### job update

Update a job's configuration from an XML file.

```bash
jenkins job update my-pipeline --from-file new-config.xml
```

**Flags:**

| Flag           | Description                        |
|----------------|------------------------------------|
| `--from-file`  | Path to XML config file (required) |

#### job copy

Copy an existing job.

```bash
jenkins job copy my-pipeline my-pipeline-copy
jenkins job copy my-pipeline new-pipeline --folder my-folder
```

**Flags:**

| Flag            | Description                             |
|-----------------|-----------------------------------------|
| `-f, --folder`  | Folder context for the copy operation   |

#### job rename

Rename a job.

```bash
jenkins job rename old-name new-name
jenkins job rename my-folder/old-name new-name
```

#### job delete

Delete a job permanently.

```bash
jenkins job delete my-pipeline --confirm
jenkins job delete my-folder/my-pipeline --confirm
```

**Flags:**

| Flag        | Description        |
|-------------|--------------------|
| `--confirm` | Confirm deletion   |

#### job enable

Enable a disabled job.

```bash
jenkins job enable my-pipeline
jenkins job enable my-folder/my-pipeline
```

#### job disable

Disable a job.

```bash
jenkins job disable my-pipeline
jenkins job disable my-folder/my-pipeline
```

#### job config

Get the raw config.xml of a job.

```bash
jenkins job config my-pipeline
jenkins job config my-pipeline > config.xml
jenkins job config my-folder/my-pipeline
```

#### job wipe-workspace

Wipe a job's workspace directory.

```bash
jenkins job wipe-workspace my-pipeline --confirm
```

**Flags:**

| Flag        | Description     |
|-------------|-----------------|
| `--confirm` | Confirm wipe    |

---

### Build Commands

Inspect and manage build runs.

#### build list

List builds for a job.

```bash
jenkins build list my-pipeline
jenkins build list my-pipeline --limit 10
jenkins build list my-pipeline --status FAILURE
jenkins build list my-pipeline --status SUCCESS --limit 50
jenkins build list my-folder/my-pipeline
jenkins build list my-pipeline -o json
```

**Flags:**

| Flag              | Description                                                                     |
|-------------------|---------------------------------------------------------------------------------|
| `-l, --limit`     | Maximum number of builds to list (default: 25)                                  |
| `--status`        | Filter by build status (SUCCESS, FAILURE, UNSTABLE, ABORTED, NOT_BUILT, RUNNING) |

#### build get

Get detailed information about a specific build.

```bash
jenkins build get my-pipeline 42
jenkins build get my-folder/my-pipeline 10
jenkins build get my-pipeline 42 -o json
```

#### build log

View or stream build console output.

```bash
# Full log (after build completes)
jenkins build log my-pipeline 42

# Stream log in real-time (like tail -f)
jenkins build log my-pipeline 42 --follow

# Save log to file
jenkins build log my-pipeline 42 > build-42.log
```

**Flags:**

| Flag            | Description                        |
|-----------------|------------------------------------|
| `-f, --follow`  | Stream the log output in real-time |

#### build stop

Stop a running build.

```bash
jenkins build stop my-pipeline 42
```

#### build delete

Delete a build record permanently.

```bash
jenkins build delete my-pipeline 42 --confirm
```

**Flags:**

| Flag        | Description        |
|-------------|--------------------|
| `--confirm` | Confirm deletion   |

#### build artifacts

List or download build artifacts.

```bash
jenkins build artifacts my-pipeline 42
jenkins build artifacts my-pipeline 42 --download
jenkins build artifacts my-pipeline 42 --download --output-dir ./artifacts
jenkins build artifacts my-pipeline 42 -o json
```

**Flags:**

| Flag              | Description                              |
|-------------------|------------------------------------------|
| `-d, --download`  | Download artifacts                       |
| `--output-dir`    | Directory to download artifacts to       |

#### build test-report

View test results for a build.

```bash
jenkins build test-report my-pipeline 42
jenkins build test-report my-pipeline 42 -o json
```

#### build env

View injected environment variables for a build.

```bash
jenkins build env my-pipeline 42
jenkins build env my-pipeline 42 -o json
```

#### build stages

View pipeline stage breakdown for a build.

```bash
jenkins build stages my-pipeline 42
jenkins build stages my-pipeline 42 -o json
```

#### build replay

Replay a pipeline build with the same script.

```bash
jenkins build replay my-pipeline 42
```

#### build open

Open a build page in the default web browser.

```bash
jenkins build open my-pipeline 42
```

---

### Queue Commands

Manage the Jenkins build queue.

#### queue list

List all items in the build queue.

```bash
jenkins queue list
jenkins queue list -o json
```

#### queue cancel

Cancel a queued build by its queue ID.

```bash
jenkins queue cancel 123
```

---

### Node Commands

Manage Jenkins nodes/agents.

#### node list

List all nodes.

```bash
jenkins node list
jenkins node list --offline
jenkins node list --online
jenkins node list -o json
```

**Flags:**

| Flag        | Description              |
|-------------|--------------------------|
| `--offline` | Show only offline nodes  |
| `--online`  | Show only online nodes   |

#### node get

Get detailed information about a node.

```bash
jenkins node get my-agent
jenkins node get "(built-in)"
jenkins node get my-agent -o json
```

#### node create

Create a new permanent agent node.

```bash
jenkins node create my-agent --remote-fs /home/jenkins --executors 2
jenkins node create build-agent --remote-fs /opt/jenkins --labels "linux docker"
```

**Flags:**

| Flag           | Description                            |
|----------------|----------------------------------------|
| `--executors`  | Number of executors (default: 1)       |
| `--remote-fs`  | Remote filesystem root (required)      |
| `--labels`     | Node labels (space-separated)          |

#### node delete

Delete a node.

```bash
jenkins node delete my-agent --confirm
```

**Flags:**

| Flag        | Description        |
|-------------|--------------------|
| `--confirm` | Confirm deletion   |

#### node enable

Bring an offline node back online.

```bash
jenkins node enable my-agent
```

#### node disable

Take a node offline.

```bash
jenkins node disable my-agent
jenkins node disable my-agent --message "Maintenance window"
```

**Flags:**

| Flag             | Description               |
|------------------|---------------------------|
| `-m, --message`  | Offline reason message    |

#### node log

View the agent log for a node.

```bash
jenkins node log my-agent
jenkins node log my-agent > agent.log
```

---

### View Commands

Manage Jenkins views (dashboard collections of jobs).

#### view list

List all views.

```bash
jenkins view list
jenkins view list -o json
```

#### view get

Get details about a view and its jobs.

```bash
jenkins view get "My View"
jenkins view get "All" -o json
```

#### view create

Create a new view.

```bash
jenkins view create "My Team"
jenkins view create "Dashboard" --type hudson.model.ListView
```

**Flags:**

| Flag     | Description                                         |
|----------|-----------------------------------------------------|
| `--type` | View type class name (default: hudson.model.ListView) |

#### view delete

Delete a view (does not delete the jobs in it).

```bash
jenkins view delete "My View" --confirm
```

**Flags:**

| Flag        | Description        |
|-------------|--------------------|
| `--confirm` | Confirm deletion   |

#### view add-job

Add a job to a view.

```bash
jenkins view add-job "My View" my-pipeline
```

#### view remove-job

Remove a job from a view.

```bash
jenkins view remove-job "My View" my-pipeline
```

---

### Plugin Commands

Manage Jenkins plugins.

#### plugin list

List installed plugins.

```bash
jenkins plugin list
jenkins plugin list --active
jenkins plugin list --enabled
jenkins plugin list -o json
```

**Flags:**

| Flag        | Description                          |
|-------------|--------------------------------------|
| `--active`  | Show only active and enabled plugins |
| `--enabled` | Show only enabled plugins            |

#### plugin get

Get details about an installed plugin.

```bash
jenkins plugin get git
jenkins plugin get workflow-aggregator
jenkins plugin get git -o json
```

#### plugin install

Install a plugin.

```bash
jenkins plugin install git
jenkins plugin install git --version 5.2.0
jenkins plugin install blueocean
```

**Flags:**

| Flag        | Description                      |
|-------------|----------------------------------|
| `--version` | Specific plugin version to install |

#### plugin uninstall

Uninstall a plugin (requires Jenkins restart).

```bash
jenkins plugin uninstall git --confirm
```

**Flags:**

| Flag        | Description              |
|-------------|--------------------------|
| `--confirm` | Confirm uninstallation   |

#### plugin check-updates

Check for available plugin updates.

```bash
jenkins plugin check-updates
jenkins plugin check-updates -o json
```

---

### Credential Commands

Manage Jenkins credentials. Credentials live in a store (default: `system`) and domain (default: `_` for global).

#### credential list

List credentials.

```bash
jenkins credential list
jenkins credential list --store system --domain my-domain
jenkins credential list --type "SSH"
jenkins credential list --type "Username with password"
jenkins credential list -o json
```

**Flags:**

| Flag       | Description                                              |
|------------|----------------------------------------------------------|
| `--store`  | Credential store (default: system)                       |
| `--domain` | Credential domain (default: _ for global)                |
| `--type`   | Filter by credential type name (case-insensitive substring match) |

#### credential get

Get details about a credential.

```bash
jenkins credential get my-ssh-key
jenkins credential get my-cred --store system --domain my-domain
jenkins credential get my-cred -o json
```

**Flags:**

| Flag       | Description                          |
|------------|--------------------------------------|
| `--store`  | Credential store (default: system)   |
| `--domain` | Credential domain (default: _)       |

#### credential create

Create a credential from an XML configuration file.

```bash
jenkins credential create --from-file cred.xml
jenkins credential create --from-file cred.xml --store system --domain my-domain
```

Example XML for a username/password credential:

```xml
<com.cloudbees.plugins.credentials.impl.UsernamePasswordCredentialsImpl>
  <scope>GLOBAL</scope>
  <id>my-cred-id</id>
  <username>admin</username>
  <password>secret</password>
  <description>My credential</description>
</com.cloudbees.plugins.credentials.impl.UsernamePasswordCredentialsImpl>
```

**Flags:**

| Flag          | Description                          |
|---------------|--------------------------------------|
| `--from-file` | Path to XML config file (required)   |
| `--store`     | Credential store (default: system)   |
| `--domain`    | Credential domain (default: _)       |

#### credential update

Update an existing credential.

```bash
jenkins credential update my-cred-id --from-file updated-cred.xml
jenkins credential update my-cred-id --from-file cred.xml --store system --domain my-domain
```

**Flags:**

| Flag          | Description                          |
|---------------|--------------------------------------|
| `--from-file` | Path to XML config file (required)   |
| `--store`     | Credential store (default: system)   |
| `--domain`    | Credential domain (default: _)       |

#### credential delete

Delete a credential.

```bash
jenkins credential delete my-cred-id --confirm
jenkins credential delete my-cred-id --store system --domain my-domain --confirm
```

**Flags:**

| Flag        | Description                          |
|-------------|--------------------------------------|
| `--confirm` | Confirm deletion                     |
| `--store`   | Credential store (default: system)   |
| `--domain`  | Credential domain (default: _)       |

---

### User Commands

List and inspect Jenkins users.

#### user list

List all known users.

```bash
jenkins user list
jenkins user list -o json
```

#### user get

Get details about a user.

```bash
jenkins user get admin
jenkins user get admin -o json
```

---

### Pipeline Commands

Validate Jenkinsfiles and manage pipeline input actions.

#### pipeline validate

Validate a declarative Jenkinsfile.

```bash
jenkins pipeline validate --from-file Jenkinsfile
jenkins pipeline validate -f ./ci/Jenkinsfile
```

**Flags:**

| Flag             | Description                       |
|------------------|-----------------------------------|
| `-f, --from-file` | Path to Jenkinsfile (required)  |

Note: Only declarative pipelines are supported. Scripted pipelines cannot be validated via this endpoint.

#### pipeline input-list

List pending input actions for a pipeline build.

```bash
jenkins pipeline input-list my-pipeline 42
jenkins pipeline input-list my-pipeline 42 -o json
```

#### pipeline input-submit

Submit (proceed with) a pending input action.

```bash
jenkins pipeline input-submit my-pipeline 42 my-input-id
jenkins pipeline input-submit my-pipeline 42 my-input-id --param APPROVE=yes
jenkins pipeline input-submit my-pipeline 42 deploy-approval --param ENV=prod --param VERSION=1.0
```

**Flags:**

| Flag          | Description                                    |
|---------------|------------------------------------------------|
| `-p, --param` | Input parameters as KEY=VALUE (repeatable)     |

#### pipeline input-abort

Abort a pending input action.

```bash
jenkins pipeline input-abort my-pipeline 42 my-input-id
```

---

### System Commands

Jenkins server administration.

#### system info

Show Jenkins system information.

```bash
jenkins system info
jenkins system info -o json
```

#### system restart

Restart the Jenkins server.

```bash
# Immediate restart
jenkins system restart --confirm

# Safe restart (wait for running builds to finish)
jenkins system restart --safe --confirm
```

**Flags:**

| Flag        | Description                        |
|-------------|------------------------------------|
| `--safe`    | Wait for builds before restarting  |
| `--confirm` | Confirm restart                    |

#### system quiet-down

Enter quiet-down mode (no new builds will start).

```bash
jenkins system quiet-down
```

#### system cancel-quiet-down

Exit quiet-down mode.

```bash
jenkins system cancel-quiet-down
```

#### system run-script

Execute a Groovy script on the Jenkins controller.

```bash
# Inline script
jenkins system run-script --script 'println Jenkins.instance.numExecutors'

# Script from file
jenkins system run-script --from-file my-script.groovy

# List all jobs
jenkins system run-script --script 'Jenkins.instance.allItems.each { println it.fullName }'

# Get system properties
jenkins system run-script --script 'System.getProperties().each { k, v -> println "$k=$v" }'
```

**Flags:**

| Flag          | Description                      |
|---------------|----------------------------------|
| `--script`    | Groovy script to execute inline  |
| `--from-file` | Path to Groovy script file       |

---

### Utility Commands

#### version

Print the CLI version.

```bash
jenkins version
```

#### update

Check for and install CLI updates.

```bash
# Check and install
jenkins update

# Check only (don't install)
jenkins update --check
```

**Flags:**

| Flag      | Description                          |
|-----------|--------------------------------------|
| `--check` | Only check, don't install            |

## Common Workflows

### Trigger a Build and Monitor It

```bash
# Trigger a parameterized build and stream the console output
jenkins job build my-pipeline --param BRANCH=main --follow

# Or trigger, wait for completion, then check the result
jenkins job build my-pipeline --wait
```

### Export and Re-import a Job Configuration

```bash
jenkins job config my-pipeline > config.xml
# Edit config.xml...
jenkins job update my-pipeline --from-file config.xml
```

### Check Build Health Across All Jobs

```bash
# List all jobs recursively and output as JSON for analysis
jenkins job list --recursive -o json

# List only failing jobs
jenkins job list --recursive --status FAILURE
```

### Debug a Failed Build

```bash
# Check the build status
jenkins build get my-pipeline 42

# View the console log
jenkins build log my-pipeline 42

# Check test results
jenkins build test-report my-pipeline 42

# View pipeline stages to find which stage failed
jenkins build stages my-pipeline 42
```

### Manage Nodes for Maintenance

```bash
# Check which nodes are offline
jenkins node list --offline

# Take a node offline for maintenance
jenkins node disable my-agent --message "Patching OS"

# Bring it back online
jenkins node enable my-agent
```

### Validate a Jenkinsfile Before Committing

```bash
jenkins pipeline validate --from-file Jenkinsfile
```

### Approve a Pipeline Deployment

```bash
# List pending inputs
jenkins pipeline input-list my-pipeline 42

# Submit approval
jenkins pipeline input-submit my-pipeline 42 deploy-approval --param ENV=production
```

## License

MIT
