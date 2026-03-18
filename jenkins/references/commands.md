# Jenkins CLI -- Full Command Reference

Complete reference for every command, subcommand, flag, and argument in the `jenkins` CLI.

## Global Flags

These flags are available on every command:

| Flag | Short | Description |
|---|---|---|
| `--output <format>` | `-o` | Output format: `table` (default), `json`, `yaml` |
| `--profile <name>` | | Configuration profile to use |
| `--server <url>` | `-s` | Jenkins server URL override |
| `--user <name>` | `-u` | Jenkins username override |
| `--token <token>` | `-t` | Jenkins API token override |
| `--insecure` | `-k` | Skip TLS certificate verification |
| `--no-color` | | Disable color output |
| `--verbose` | `-v` | Enable verbose output |

---

## Top-Level Commands

### `jenkins login`

Interactively authenticate with a Jenkins server. Prompts for URL, username, API token, and profile name. Tests the connection before saving.

```bash
jenkins login
jenkins login --name staging
jenkins login --insecure
```

| Flag | Description |
|---|---|
| `--name <profile>` | Profile name (default: `default`) |

### `jenkins status`

Show Jenkins server connection status and information (version, mode, executors).

```bash
jenkins status
jenkins status -o json
jenkins status --profile staging
```

### `jenkins whoami`

Show the currently authenticated user.

```bash
jenkins whoami
jenkins whoami -o json
```

### `jenkins version`

Print the CLI version.

```bash
jenkins version
```

### `jenkins update`

Check for and install CLI updates.

```bash
jenkins update
jenkins update --check
```

| Flag | Description |
|---|---|
| `--check` | Only check for updates, don't install |

---

## Job Commands (`jenkins job`)

Aliases: `jobs`

### `jenkins job list`

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

| Flag | Short | Description |
|---|---|---|
| `--folder <path>` | `-f` | Folder path to list jobs from |
| `--recursive` | `-r` | List jobs recursively through all subfolders |
| `--status <status>` | | Filter by job status: SUCCESS, FAILURE, UNSTABLE, DISABLED, ABORTED, NOT_BUILT, RUNNING |

### `jenkins job get <path>`

Get detailed information about a job.

```bash
jenkins job get my-pipeline
jenkins job get my-folder/my-pipeline
jenkins job get my-pipeline -o json
```

### `jenkins job build <path>`

Trigger a build for a job.

```bash
jenkins job build my-pipeline
jenkins job build my-pipeline --param BRANCH=main --param ENV=staging
jenkins job build my-pipeline --wait
jenkins job build my-pipeline --follow
jenkins job build my-pipeline --param BRANCH=main --wait --follow --timeout 1h
jenkins job build my-folder/my-pipeline --param VERSION=1.2.3
```

| Flag | Short | Description |
|---|---|---|
| `--param <KEY=VALUE>` | `-p` | Build parameters (repeatable) |
| `--wait` | `-w` | Wait for the build to complete |
| `--follow` | `-F` | Stream the build console output in real time |
| `--timeout <duration>` | | Timeout for --wait/--follow (default: 30m) |

### `jenkins job create <path>`

Create a new job from an XML configuration file.

```bash
jenkins job create my-new-job --from-file config.xml
jenkins job create my-new-job --from-file config.xml --folder my-folder
```

| Flag | Short | Description |
|---|---|---|
| `--from-file <path>` | | Path to XML config file (required) |
| `--folder <path>` | `-f` | Folder to create the job in |

### `jenkins job update <path>`

Update a job's configuration from an XML file.

```bash
jenkins job update my-pipeline --from-file new-config.xml
```

| Flag | Description |
|---|---|
| `--from-file <path>` | Path to XML config file (required) |

### `jenkins job copy <source> <target>`

Copy an existing job.

```bash
jenkins job copy my-pipeline my-pipeline-copy
jenkins job copy my-pipeline new-pipeline --folder my-folder
```

| Flag | Short | Description |
|---|---|---|
| `--folder <path>` | `-f` | Folder context for the copy operation |

### `jenkins job rename <old-path> <new-name>`

Rename a job.

```bash
jenkins job rename old-name new-name
jenkins job rename my-folder/old-name new-name
```

### `jenkins job delete <path>`

Delete a job permanently.

```bash
jenkins job delete my-pipeline --confirm
jenkins job delete my-folder/my-pipeline --confirm
```

| Flag | Description |
|---|---|
| `--confirm` | Confirm deletion (required) |

### `jenkins job enable <path>`

Enable a disabled job.

```bash
jenkins job enable my-pipeline
jenkins job enable my-folder/my-pipeline
```

### `jenkins job disable <path>`

Disable a job.

```bash
jenkins job disable my-pipeline
jenkins job disable my-folder/my-pipeline
```

### `jenkins job config <path>`

Get the raw config.xml of a job.

```bash
jenkins job config my-pipeline
jenkins job config my-pipeline > config.xml
jenkins job config my-folder/my-pipeline
```

### `jenkins job wipe-workspace <path>`

Wipe a job's workspace directory.

```bash
jenkins job wipe-workspace my-pipeline --confirm
```

| Flag | Description |
|---|---|
| `--confirm` | Confirm wipe (required) |

---

## Build Commands (`jenkins build`)

Aliases: `builds`

### `jenkins build list <job-path>`

List builds for a job.

```bash
jenkins build list my-pipeline
jenkins build list my-pipeline --limit 10
jenkins build list my-pipeline --status FAILURE
jenkins build list my-pipeline --status SUCCESS --limit 50
jenkins build list my-folder/my-pipeline
jenkins build list my-pipeline -o json
```

| Flag | Short | Description |
|---|---|---|
| `--limit <n>` | `-l` | Maximum number of builds to list (default: 25) |
| `--status <status>` | | Filter by status: SUCCESS, FAILURE, UNSTABLE, ABORTED, NOT_BUILT, RUNNING |

### `jenkins build get <job-path> <build-number>`

Get detailed information about a specific build.

```bash
jenkins build get my-pipeline 42
jenkins build get my-folder/my-pipeline 10
jenkins build get my-pipeline 42 -o json
```

### `jenkins build log <job-path> <build-number>`

View or stream build console output.

```bash
jenkins build log my-pipeline 42
jenkins build log my-pipeline 42 --follow
jenkins build log my-pipeline 42 > build-42.log
```

| Flag | Short | Description |
|---|---|---|
| `--follow` | `-f` | Stream the log output in real-time |

### `jenkins build stop <job-path> <build-number>`

Stop a running build.

```bash
jenkins build stop my-pipeline 42
```

### `jenkins build delete <job-path> <build-number>`

Delete a build record permanently.

```bash
jenkins build delete my-pipeline 42 --confirm
```

| Flag | Description |
|---|---|
| `--confirm` | Confirm deletion (required) |

### `jenkins build artifacts <job-path> <build-number>`

List or download build artifacts.

```bash
jenkins build artifacts my-pipeline 42
jenkins build artifacts my-pipeline 42 -o json
jenkins build artifacts my-pipeline 42 --download
jenkins build artifacts my-pipeline 42 --download --output-dir ./artifacts
```

| Flag | Short | Description |
|---|---|---|
| `--download` | `-d` | Download artifacts |
| `--output-dir <path>` | | Directory to download artifacts to |

### `jenkins build test-report <job-path> <build-number>`

View test results for a build.

```bash
jenkins build test-report my-pipeline 42
jenkins build test-report my-pipeline 42 -o json
```

### `jenkins build env <job-path> <build-number>`

View injected environment variables for a build.

```bash
jenkins build env my-pipeline 42
jenkins build env my-pipeline 42 -o json
```

### `jenkins build stages <job-path> <build-number>`

View pipeline stage breakdown for a build.

```bash
jenkins build stages my-pipeline 42
jenkins build stages my-pipeline 42 -o json
```

### `jenkins build replay <job-path> <build-number>`

Replay a pipeline build with the same script.

```bash
jenkins build replay my-pipeline 42
```

### `jenkins build open <job-path> <build-number>`

Open a build page in the default web browser.

```bash
jenkins build open my-pipeline 42
```

---

## Queue Commands (`jenkins queue`)

### `jenkins queue list`

List all items in the build queue.

```bash
jenkins queue list
jenkins queue list -o json
```

### `jenkins queue cancel <queue-id>`

Cancel a queued build by its queue ID.

```bash
jenkins queue cancel 123
```

---

## Pipeline Commands (`jenkins pipeline`)

### `jenkins pipeline validate`

Validate a declarative Jenkinsfile. Only declarative pipelines are supported; scripted pipelines cannot be validated via this endpoint.

```bash
jenkins pipeline validate --from-file Jenkinsfile
jenkins pipeline validate -f ./ci/Jenkinsfile
```

| Flag | Short | Description |
|---|---|---|
| `--from-file <path>` | `-f` | Path to Jenkinsfile (required) |

### `jenkins pipeline input-list <job-path> <build-number>`

List pending input actions for a pipeline build.

```bash
jenkins pipeline input-list my-pipeline 42
jenkins pipeline input-list my-pipeline 42 -o json
```

### `jenkins pipeline input-submit <job-path> <build-number> <input-id>`

Submit (proceed with) a pending input action.

```bash
jenkins pipeline input-submit my-pipeline 42 my-input-id
jenkins pipeline input-submit my-pipeline 42 my-input-id --param APPROVE=yes
jenkins pipeline input-submit my-pipeline 42 deploy-approval --param ENV=prod --param VERSION=1.0
```

| Flag | Short | Description |
|---|---|---|
| `--param <KEY=VALUE>` | `-p` | Input parameters (repeatable) |

### `jenkins pipeline input-abort <job-path> <build-number> <input-id>`

Abort a pending input action.

```bash
jenkins pipeline input-abort my-pipeline 42 my-input-id
```

---

## Node Commands (`jenkins node`)

Aliases: `nodes`, `agent`, `agents`

### `jenkins node list`

List all nodes.

```bash
jenkins node list
jenkins node list --offline
jenkins node list --online
jenkins node list -o json
```

| Flag | Description |
|---|---|
| `--offline` | Show only offline nodes |
| `--online` | Show only online nodes |

### `jenkins node get <name>`

Get detailed information about a node.

```bash
jenkins node get my-agent
jenkins node get "(built-in)"
jenkins node get my-agent -o json
```

### `jenkins node create <name>`

Create a new permanent agent node.

```bash
jenkins node create my-agent --remote-fs /home/jenkins --executors 2
jenkins node create build-agent --remote-fs /opt/jenkins --labels "linux docker"
```

| Flag | Description |
|---|---|
| `--remote-fs <path>` | Remote filesystem root (required) |
| `--executors <n>` | Number of executors (default: 1) |
| `--labels <labels>` | Node labels (space-separated) |

### `jenkins node delete <name>`

Delete a node.

```bash
jenkins node delete my-agent --confirm
```

| Flag | Description |
|---|---|
| `--confirm` | Confirm deletion (required) |

### `jenkins node enable <name>`

Bring an offline node back online.

```bash
jenkins node enable my-agent
```

### `jenkins node disable <name>`

Take a node offline.

```bash
jenkins node disable my-agent
jenkins node disable my-agent --message "Maintenance window"
```

| Flag | Short | Description |
|---|---|---|
| `--message <reason>` | `-m` | Offline reason message |

### `jenkins node log <name>`

View the agent log for a node.

```bash
jenkins node log my-agent
jenkins node log my-agent > agent.log
```

---

## View Commands (`jenkins view`)

Aliases: `views`

### `jenkins view list`

List all views.

```bash
jenkins view list
jenkins view list -o json
```

### `jenkins view get <name>`

Get details about a view and its jobs.

```bash
jenkins view get "My View"
jenkins view get "All" -o json
```

### `jenkins view create <name>`

Create a new view.

```bash
jenkins view create "My Team"
jenkins view create "Dashboard" --type hudson.model.ListView
```

| Flag | Description |
|---|---|
| `--type <class>` | View type class name (default: hudson.model.ListView) |

### `jenkins view delete <name>`

Delete a view (does not delete the jobs in it).

```bash
jenkins view delete "My View" --confirm
```

| Flag | Description |
|---|---|
| `--confirm` | Confirm deletion (required) |

### `jenkins view add-job <view-name> <job-path>`

Add a job to a view.

```bash
jenkins view add-job "My View" my-pipeline
```

### `jenkins view remove-job <view-name> <job-path>`

Remove a job from a view.

```bash
jenkins view remove-job "My View" my-pipeline
```

---

## Plugin Commands (`jenkins plugin`)

Aliases: `plugins`

### `jenkins plugin list`

List installed plugins.

```bash
jenkins plugin list
jenkins plugin list --active
jenkins plugin list --enabled
jenkins plugin list -o json
```

| Flag | Description |
|---|---|
| `--active` | Show only active and enabled plugins |
| `--enabled` | Show only enabled plugins |

### `jenkins plugin get <short-name>`

Get details about an installed plugin.

```bash
jenkins plugin get git
jenkins plugin get workflow-aggregator
jenkins plugin get git -o json
```

### `jenkins plugin install <short-name>`

Install a plugin.

```bash
jenkins plugin install git
jenkins plugin install git --version 5.2.0
jenkins plugin install blueocean
```

| Flag | Description |
|---|---|
| `--version <ver>` | Specific plugin version to install |

### `jenkins plugin uninstall <short-name>`

Uninstall a plugin (requires Jenkins restart to take effect).

```bash
jenkins plugin uninstall git --confirm
```

| Flag | Description |
|---|---|
| `--confirm` | Confirm uninstallation (required) |

### `jenkins plugin check-updates`

Check for available plugin updates.

```bash
jenkins plugin check-updates
jenkins plugin check-updates -o json
```

---

## Credential Commands (`jenkins credential`)

Aliases: `credentials`, `cred`, `creds`

Credentials live in a store (default: `system`) and domain (default: `_` for global).

### `jenkins credential list`

List credentials.

```bash
jenkins credential list
jenkins credential list --store system --domain my-domain
jenkins credential list --type "SSH"
jenkins credential list --type "Username with password"
jenkins credential list -o json
```

| Flag | Description |
|---|---|
| `--store <name>` | Credential store (default: `system`) |
| `--domain <name>` | Credential domain (default: `_` for global) |
| `--type <filter>` | Filter by credential type name (case-insensitive substring match) |

### `jenkins credential get <id>`

Get details about a credential.

```bash
jenkins credential get my-ssh-key
jenkins credential get my-cred --store system --domain my-domain
jenkins credential get my-cred -o json
```

| Flag | Description |
|---|---|
| `--store <name>` | Credential store (default: `system`) |
| `--domain <name>` | Credential domain (default: `_`) |

### `jenkins credential create`

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

| Flag | Description |
|---|---|
| `--from-file <path>` | Path to XML config file (required) |
| `--store <name>` | Credential store (default: `system`) |
| `--domain <name>` | Credential domain (default: `_`) |

### `jenkins credential update <id>`

Update an existing credential.

```bash
jenkins credential update my-cred-id --from-file updated-cred.xml
jenkins credential update my-cred-id --from-file cred.xml --store system --domain my-domain
```

| Flag | Description |
|---|---|
| `--from-file <path>` | Path to XML config file (required) |
| `--store <name>` | Credential store (default: `system`) |
| `--domain <name>` | Credential domain (default: `_`) |

### `jenkins credential delete <id>`

Delete a credential.

```bash
jenkins credential delete my-cred-id --confirm
jenkins credential delete my-cred-id --store system --domain my-domain --confirm
```

| Flag | Description |
|---|---|
| `--confirm` | Confirm deletion (required) |
| `--store <name>` | Credential store (default: `system`) |
| `--domain <name>` | Credential domain (default: `_`) |

---

## User Commands (`jenkins user`)

Aliases: `users`

### `jenkins user list`

List all known users.

```bash
jenkins user list
jenkins user list -o json
```

### `jenkins user get <user-id>`

Get details about a user.

```bash
jenkins user get admin
jenkins user get admin -o json
```

---

## System Commands (`jenkins system`)

### `jenkins system info`

Show Jenkins system information (version, mode, executors).

```bash
jenkins system info
jenkins system info -o json
```

### `jenkins system restart`

Restart the Jenkins server.

```bash
jenkins system restart --confirm
jenkins system restart --safe --confirm
```

| Flag | Description |
|---|---|
| `--safe` | Wait for running builds to finish before restarting |
| `--confirm` | Confirm restart (required) |

### `jenkins system quiet-down`

Enter quiet-down mode (no new builds will start).

```bash
jenkins system quiet-down
```

### `jenkins system cancel-quiet-down`

Exit quiet-down mode.

```bash
jenkins system cancel-quiet-down
```

### `jenkins system run-script`

Execute a Groovy script on the Jenkins controller. This runs with full controller access.

```bash
jenkins system run-script --script 'println Jenkins.instance.numExecutors'
jenkins system run-script --from-file my-script.groovy
jenkins system run-script --script 'Jenkins.instance.allItems.each { println it.fullName }'
jenkins system run-script --script 'System.getProperties().each { k, v -> println "$k=$v" }'
```

| Flag | Description |
|---|---|
| `--script <code>` | Groovy script to execute inline |
| `--from-file <path>` | Path to Groovy script file |
