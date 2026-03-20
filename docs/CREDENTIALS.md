# Jenkins CLI - Authentication & Credentials Guide

Complete guide to authenticating with Jenkins servers using the `jenkins` CLI. Covers interactive login, environment variables, API tokens, CSRF crumbs, TLS, RBAC, and every edge case you will encounter in production.

---

## Table of Contents

- [Quick Start](#quick-start)
- [Getting Your Credentials](#getting-your-credentials)
- [Understanding Jenkins Authentication](#understanding-jenkins-authentication)
- [Minimum Required Permissions](#minimum-required-permissions)
- [Configuration](#configuration)
- [TLS / SSL Configuration](#tls--ssl-configuration)
- [Read-Only Mode](#read-only-mode)
- [CI / Automation / Agent Usage](#ci--automation--agent-usage)
- [Edge Cases & Troubleshooting](#edge-cases--troubleshooting)
- [Security Best Practices](#security-best-practices)

---

## Quick Start

### Option 1: Interactive Login (Recommended)

```bash
jenkins login
```

You will be prompted for four values:

```
Jenkins URL: https://jenkins.example.com
Username: admin
API Token: 11a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6
Profile name [default]: default
Testing connection... OK (authenticated as Admin User)
Profile "default" saved to /home/you/.config/jenkins-cli/config.yaml
```

The CLI tests the connection before saving. If the test fails, nothing is written to disk.

### Option 2: Environment Variables

```bash
export JENKINS_URL=https://jenkins.example.com
export JENKINS_USER=admin
export JENKINS_TOKEN=11a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6
```

Then run any command directly:

```bash
jenkins status
jenkins job list
```

No `jenkins login` is needed when environment variables are set. This is the preferred method for CI pipelines, Docker containers, and automation scripts.

### Option 3: Command-Line Flags

```bash
jenkins status \
  --server https://jenkins.example.com \
  --user admin \
  --token 11a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6
```

Useful for one-off commands or testing. Flags override both environment variables and config file values.

---

## Getting Your Credentials

### Step 1: Find Your Jenkins URL

The Jenkins URL is the address you use to access the Jenkins web UI in your browser. The CLI needs the **base URL** -- no trailing slash, no path to a specific page.

| Deployment | Typical URL | Notes |
|---|---|---|
| Self-hosted (default) | `http://localhost:8080` | Jenkins default port is 8080 |
| Self-hosted (custom port) | `http://jenkins.internal:9090` | Check `--httpPort` in startup args |
| Behind reverse proxy | `https://jenkins.example.com` | Use the external-facing URL |
| With context path | `https://ci.example.com/jenkins` | Include the context path |
| Kubernetes (Ingress) | `https://jenkins.k8s.example.com` | Use the Ingress hostname |
| Kubernetes (port-forward) | `http://localhost:8080` | After `kubectl port-forward` |

**How to verify your URL:**

```bash
# The URL should return JSON when you append /api/json
curl -s https://jenkins.example.com/api/json | head -c 100

# If you see JSON output starting with {"_class":"hudson.model...", the URL is correct.
# If you see HTML or a redirect, adjust the URL.
```

**Context path gotcha:** If Jenkins is deployed at `https://ci.example.com/jenkins`, you must include `/jenkins` in the URL. The CLI will not discover it automatically.

```bash
# Wrong -- will fail with 404 or redirect to login page
jenkins login   # URL: https://ci.example.com

# Correct
jenkins login   # URL: https://ci.example.com/jenkins
```

### Step 2: Generate an API Token

Jenkins API tokens are the recommended authentication method. They replace your password for API and CLI access.

**Step-by-step instructions:**

1. **Log in** to the Jenkins web UI at `https://jenkins.example.com`
2. **Click your username** in the top-right corner of the page
3. **Click "Configure"** in the left sidebar (or go directly to `https://jenkins.example.com/me/configure`)
4. **Scroll down** to the **"API Token"** section
5. **Click "Add new Token"**
6. **Enter a descriptive name** (e.g., `cli-access`, `ci-pipeline-token`, `my-laptop`)
7. **Click "Generate"**
8. **Copy the token immediately** -- it is displayed only once and cannot be retrieved later

The token will look something like: `11a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6`

**Important details about API tokens:**

- An API token is **not** your Jenkins password. It is a separate credential generated specifically for API access.
- You can create **multiple tokens** per user. Use separate tokens for different tools (CLI, CI pipelines, scripts) so you can revoke them independently.
- Tokens can be **individually revoked** from the same "API Token" section where you generated them.
- Tokens are stored as **SHA-256 hashes** on the Jenkins server. If the Jenkins data directory is compromised, the raw tokens cannot be recovered from the hashes.
- Tokens **inherit the permissions** of the user who created them. The token has the same access level as your user account.
- Tokens **do not expire** by default. Some Jenkins administrators configure token expiration policies via plugins.

### Step 3: Use with the CLI

**Interactive login (saves credentials to config file):**

```bash
jenkins login
# Enter: URL, username, API token, profile name
```

**Environment variables (no config file needed):**

```bash
export JENKINS_URL=https://jenkins.example.com
export JENKINS_USER=admin
export JENKINS_TOKEN=11a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6

jenkins job list
```

**Inline flags (single command, nothing saved):**

```bash
jenkins job list \
  --server https://jenkins.example.com \
  --user admin \
  --token 11a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6
```

**Verify your setup:**

```bash
# Check connection and server info
jenkins status

# Confirm which user you are authenticated as
jenkins whoami
```

---

## Understanding Jenkins Authentication

### How the CLI Authenticates

The CLI uses **HTTP Basic Authentication** on every request. It sends your username and API token as the Basic Auth credentials. This is the standard Jenkins API authentication mechanism.

```
Authorization: Basic base64(username:token)
```

The CLI handles this automatically -- you never need to construct the header yourself.

### API Token vs Password

| | API Token | Password |
|---|---|---|
| **Recommended** | Yes | No |
| **CSRF crumb required** | No (for most Jenkins versions) | Yes |
| **Can be revoked independently** | Yes | No (changing password invalidates everything) |
| **Stored on Jenkins server** | SHA-256 hash | Depends on security realm |
| **Works with all auth backends** | Yes (LDAP, SAML, local) | Depends on the security realm |
| **Multiple per user** | Yes | No |
| **Risk if leaked** | Limited (can revoke just that token) | Full account compromise |

**Why API tokens are preferred:** API tokens decouple API access from your login password. You can revoke a single token without affecting your ability to log in to the Jenkins UI or other tools. Tokens also bypass the CSRF crumb requirement for most POST requests, which simplifies programmatic access.

### CSRF Crumb Tokens

Jenkins protects against Cross-Site Request Forgery (CSRF) attacks using a crumb-based system. Here is what you need to know:

**What they are:** A CSRF crumb is a short-lived token that Jenkins issues to prove that a request originated from a trusted source. The crumb is a header (typically `Jenkins-Crumb`) that must be included on POST requests.

**When they are needed:** POST requests (build triggers, job creation, configuration changes, restarts, etc.) require a crumb when CSRF protection is enabled -- which it is by default in modern Jenkins installations.

**How the CLI handles crumbs automatically:** You never need to manage crumbs manually. The CLI:

1. Fetches a crumb from `/crumbIssuer/api/json` before the first POST request
2. Caches the crumb for 5 minutes to avoid redundant network calls
3. Injects the crumb header into every POST request automatically
4. Handles the case where CSRF protection is disabled (404 from crumb endpoint)

**When crumbs cause problems:** If you see `No valid crumb` errors, it usually means you are using password authentication with a misconfigured Jenkins proxy. The fix is to switch to API token authentication.

---

## Minimum Required Permissions

The permissions your Jenkins user (and therefore your API token) needs depend on which CLI commands you use.

### Permission Matrix

| CLI Command | Required Jenkins Permission | Notes |
|---|---|---|
| `jenkins status` | Overall/Read | Basic server connectivity |
| `jenkins whoami` | Overall/Read | View own user info |
| `jenkins job list` | Job/Read | On the target folder |
| `jenkins job get <path>` | Job/Read | On the specific job |
| `jenkins job config <path>` | Job/ExtendedRead or Job/Configure | ExtendedRead for viewing config |
| `jenkins job build <path>` | Job/Build | On the specific job |
| `jenkins job create <path>` | Job/Create | On the target folder |
| `jenkins job update <path>` | Job/Configure | On the specific job |
| `jenkins job copy <src> <dst>` | Job/Create + Job/Read | Create in target, Read on source |
| `jenkins job rename <old> <new>` | Job/Configure | On the specific job |
| `jenkins job delete <path>` | Job/Delete | On the specific job |
| `jenkins job enable <path>` | Job/Configure | On the specific job |
| `jenkins job disable <path>` | Job/Configure | On the specific job |
| `jenkins job wipe-workspace <path>` | Job/WipeOut | On the specific job |
| `jenkins build list <path>` | Job/Read | On the specific job |
| `jenkins build get <path> <num>` | Job/Read | On the specific job |
| `jenkins build log <path> <num>` | Job/Read | On the specific job |
| `jenkins build stop <path> <num>` | Job/Cancel | On the specific job |
| `jenkins build delete <path> <num>` | Run/Delete | On the specific job |
| `jenkins build artifacts <path> <num>` | Job/Read | On the specific job |
| `jenkins build test-report <path> <num>` | Job/Read | On the specific job |
| `jenkins build env <path> <num>` | Job/Read | May require Run/Artifacts |
| `jenkins build stages <path> <num>` | Job/Read | Requires Pipeline plugin |
| `jenkins build replay <path> <num>` | Job/Build + Job/Configure | Replay requires configure |
| `jenkins queue list` | Overall/Read | View the build queue |
| `jenkins queue cancel <id>` | Overall/Cancel | Cancel queued items |
| `jenkins node list` | Computer/Read (Overall/Read) | List agents |
| `jenkins node get <name>` | Computer/Read | Specific agent |
| `jenkins node create <name>` | Computer/Create | Create agents |
| `jenkins node delete <name>` | Computer/Delete | Delete agents |
| `jenkins node enable <name>` | Computer/Connect | Bring agent online |
| `jenkins node disable <name>` | Computer/Disconnect | Take agent offline |
| `jenkins node log <name>` | Computer/Read | View agent logs |
| `jenkins view list` | Overall/Read | List views |
| `jenkins view get <name>` | Overall/Read | View details |
| `jenkins view create <name>` | View/Create | Create views |
| `jenkins view delete <name>` | View/Delete | Delete views |
| `jenkins view add-job` | View/Configure | Modify view membership |
| `jenkins view remove-job` | View/Configure | Modify view membership |
| `jenkins plugin list` | Overall/Read | List plugins |
| `jenkins plugin get <name>` | Overall/Read | Plugin details |
| `jenkins plugin install <name>` | Overall/Administer | Install plugins |
| `jenkins plugin uninstall <name>` | Overall/Administer | Uninstall plugins |
| `jenkins plugin check-updates` | Overall/Administer | Check for updates |
| `jenkins credential list` | Credentials/View | On the credential domain |
| `jenkins credential get <id>` | Credentials/View | On the specific credential |
| `jenkins credential create` | Credentials/Create | On the credential domain |
| `jenkins credential update <id>` | Credentials/Update | On the specific credential |
| `jenkins credential delete <id>` | Credentials/Delete | On the specific credential |
| `jenkins user list` | Overall/Read | List users |
| `jenkins user get <id>` | Overall/Read | User details |
| `jenkins pipeline validate` | Overall/Read | Validate Jenkinsfile |
| `jenkins pipeline input-list` | Job/Read | View pending inputs |
| `jenkins pipeline input-submit` | Job/Build | Approve/submit input |
| `jenkins pipeline input-abort` | Job/Cancel or Job/Build | Abort pending input |
| `jenkins system info` | Overall/Read | System information |
| `jenkins system restart` | Overall/Administer | Restart Jenkins |
| `jenkins system quiet-down` | Overall/Administer | Enter quiet mode |
| `jenkins system cancel-quiet-down` | Overall/Administer | Exit quiet mode |
| `jenkins system run-script` | Overall/Administer | Execute Groovy scripts |

### Role-Based Access Control (RBAC)

Jenkins supports several authorization strategies. The permissions above apply regardless of which strategy is in use.

**Matrix-based Security (built-in):**

The default fine-grained authorization. An administrator assigns permissions to individual users or groups in **Manage Jenkins > Security > Authorization**. Each permission (Job/Read, Job/Build, etc.) is toggled per user.

**Role Strategy Plugin:**

Adds role-based access control. Administrators define roles (e.g., `developer`, `viewer`, `admin`) with specific permission sets, then assign users to roles. This is the most common RBAC approach in larger organizations.

To use the CLI with Role Strategy:
1. Ask your Jenkins administrator which role you have
2. Verify the role includes the permissions for the commands you need (see table above)
3. If you get 403 errors, request the specific missing permission

**Project-based Matrix Authorization:**

Extends matrix authorization to individual jobs and folders. A user might have Job/Build on `team-a/*` jobs but not on `team-b/*` jobs. If you can see a job in the Jenkins UI but get a 403 from the CLI, check the project-level permissions.

**Folder-level Permissions:**

When using the Folders plugin, permissions can be scoped to specific folders. Your token inherits your user's permissions, including folder-level grants. If `jenkins job list` works but `jenkins job list --folder restricted-folder` returns 403, you lack permissions on that folder.

---

## Configuration

### Config File

The CLI stores profiles in a YAML config file.

**Default location:**

```
~/.config/jenkins-cli/config.yaml
```

If the `XDG_CONFIG_HOME` environment variable is set, the config file is located at:

```
$XDG_CONFIG_HOME/jenkins-cli/config.yaml
```

**Full config file structure:**

```yaml
# The profile to use when --profile is not specified
current_profile: production

# Connection profiles
profiles:
  production:
    url: https://jenkins.example.com
    username: admin
    token: 11a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6
    insecure: false
    read_only: false

  staging:
    url: https://jenkins-staging.example.com
    username: deployer
    token: 22b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7
    insecure: false
    read_only: false

  local:
    url: http://localhost:8080
    username: admin
    token: 33c3d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8
    insecure: false
    read_only: false

  dev-selfsigned:
    url: https://jenkins-dev.internal:8443
    username: developer
    token: 44d4e5f6a7b8c9d0e1f2a3b4c5d6e7f8a9
    insecure: true       # Skip TLS verification for self-signed certs
    read_only: true      # Block all write operations

# Default settings
defaults:
  output: table          # Default output format (table, json, yaml)
```

**File permissions:** The config file is created with mode `0600` (read/write for owner only). The config directory is created with mode `0700`. The CLI enforces these permissions when writing.

### Environment Variables

Environment variables override config file values. CLI flags override environment variables.

| Environment Variable | Description | Example |
|---|---|---|
| `JENKINS_URL` | Jenkins server URL | `https://jenkins.example.com` |
| `JENKINS_USER` | Jenkins username | `admin` |
| `JENKINS_TOKEN` | Jenkins API token | `11a1b2c3d4e5f6...` |
| `JENKINS_INSECURE` | Skip TLS verification | `true` or `false` |
| `JENKINS_READ_ONLY` | Block write operations | `true` or `false` |
| `JENKINS_NO_INPUT` | Disable interactive prompts | `true` or `1` |
| `JENKINS_QUIET` | Suppress informational output | `true` or `1` |
| `XDG_CONFIG_HOME` | Override config directory base | `/custom/config/path` |

**Priority order (highest to lowest):**

1. **CLI flags** (`--server`, `--user`, `--token`, `--insecure`)
2. **Environment variables** (`JENKINS_URL`, `JENKINS_USER`, `JENKINS_TOKEN`, `JENKINS_INSECURE`)
3. **Config file profile** (selected by `--profile` flag or `current_profile` in config)

This means you can have a config file with default profiles and override individual values with environment variables or flags for specific invocations:

```bash
# Use the "production" profile from config, but override the token
JENKINS_TOKEN=temporary-token jenkins job list --profile production

# Use environment variables for URL/user, but override insecure via flag
export JENKINS_URL=https://jenkins.example.com
export JENKINS_USER=admin
export JENKINS_TOKEN=mytoken
jenkins status --insecure
```

### Multiple Profiles

Profiles let you manage connections to multiple Jenkins instances. Each profile stores a complete set of connection parameters.

**Creating profiles:**

```bash
# Create the first profile (becomes the active profile automatically)
jenkins login --name production
# Prompts: URL, username, token

# Create additional profiles
jenkins login --name staging
jenkins login --name local-dev
```

**Using profiles:**

```bash
# Use the currently active profile (set during login or manually)
jenkins status

# Explicitly select a profile for a single command
jenkins job list --profile staging
jenkins build log my-pipeline 42 --profile production

# Mix profile with flag overrides
jenkins status --profile staging --insecure
```

**Example workflow with multiple Jenkins instances:**

```bash
# Set up profiles for your environments
jenkins login --name prod      # https://jenkins.example.com
jenkins login --name staging   # https://jenkins-staging.example.com
jenkins login --name local     # http://localhost:8080

# Check build status across environments
jenkins job list --recursive --status FAILURE --profile prod -o json
jenkins job list --recursive --status FAILURE --profile staging -o json

# Promote a build: check staging, then trigger prod
jenkins build get deploy-pipeline 42 --profile staging -o json
jenkins job build deploy-pipeline --param VERSION=1.2.3 --profile prod --wait --follow
```

---

## TLS / SSL Configuration

### Self-Signed Certificates

Development and internal Jenkins instances often use self-signed TLS certificates. By default, the CLI rejects connections to servers with untrusted certificates. You have three ways to handle this:

**Option 1: `--insecure` flag (per command)**

```bash
jenkins status --insecure
jenkins job list --insecure
jenkins job build my-pipeline --insecure --wait
```

**Option 2: `JENKINS_INSECURE` environment variable (session-wide)**

```bash
export JENKINS_INSECURE=true
jenkins status
jenkins job list
# All commands in this session skip TLS verification
```

**Option 3: `insecure: true` in profile (permanent)**

```bash
# During login
jenkins login --insecure
# The --insecure flag is saved to the profile

# Or edit the config file manually
# ~/.config/jenkins-cli/config.yaml
```

```yaml
profiles:
  dev:
    url: https://jenkins-dev.internal:8443
    username: developer
    token: mytoken
    insecure: true  # Skip TLS verification for this profile
```

**Security warning:** The `--insecure` flag disables all TLS certificate validation, including hostname verification. This makes the connection vulnerable to man-in-the-middle attacks. Only use this for development/testing environments on trusted networks.

### Jenkins with HTTPS

**Reverse proxy (nginx/Apache) with SSL termination:**

This is the most common production setup. An nginx or Apache reverse proxy handles TLS, and Jenkins runs on HTTP internally.

```
Client (CLI) --HTTPS--> nginx/Apache --HTTP--> Jenkins (:8080)
```

Use the external HTTPS URL with the CLI:

```bash
jenkins login
# URL: https://jenkins.example.com
```

No special TLS configuration is needed on the CLI side as long as the reverse proxy uses a certificate from a trusted CA (e.g., Let's Encrypt, DigiCert).

**Jenkins native HTTPS (keystore):**

Jenkins can serve HTTPS directly using a Java keystore. This is configured with `--httpsPort`, `--httpsKeyStore`, and `--httpsKeyStorePassword` on the Jenkins startup command.

```bash
# If using a trusted CA certificate
jenkins login
# URL: https://jenkins.example.com:8443

# If using a self-signed certificate
jenkins login --insecure
# URL: https://jenkins.example.com:8443
```

**Corporate CA / internal PKI:**

If your organization uses an internal Certificate Authority, add the CA certificate to your system's trust store so the CLI trusts it automatically:

```bash
# macOS: Add to system keychain
sudo security add-trusted-cert -d -r trustRoot \
  -k /Library/Keychains/System.keychain corp-ca.pem

# Ubuntu/Debian: Add to system certificates
sudo cp corp-ca.pem /usr/local/share/ca-certificates/corp-ca.crt
sudo update-ca-certificates

# RHEL/CentOS/Fedora
sudo cp corp-ca.pem /etc/pki/ca-trust/source/anchors/corp-ca.pem
sudo update-ca-trust
```

After adding the CA certificate to your system trust store, the CLI will accept certificates signed by that CA without needing `--insecure`.

---

## Read-Only Mode

The CLI supports a read-only safety mode that blocks all write operations (job creation, builds, deletes, restarts, etc.). This is useful for monitoring dashboards, coding agents, and any context where accidental mutations would be dangerous.

**Enable read-only mode:**

```bash
# Per command
jenkins job list --read-only

# Via environment variable
export JENKINS_READ_ONLY=true
jenkins job list    # works
jenkins job build my-pipeline  # blocked with error

# In config profile
```

```yaml
profiles:
  monitoring:
    url: https://jenkins.example.com
    username: monitor-user
    token: mytoken
    read_only: true
```

**What gets blocked:**

Any command annotated as a write operation, including: `job build`, `job create`, `job update`, `job delete`, `job copy`, `job rename`, `job enable`, `job disable`, `job wipe-workspace`, `build stop`, `build delete`, `build replay`, `queue cancel`, `node create`, `node delete`, `node enable`, `node disable`, `view create`, `view delete`, `view add-job`, `view remove-job`, `plugin install`, `plugin uninstall`, `credential create`, `credential update`, `credential delete`, `system restart`, `system quiet-down`, `system cancel-quiet-down`, `system run-script`.

**Override read-only per command:**

```bash
# Profile has read_only: true, but you need to trigger one build
jenkins job build my-pipeline --read-only=false
```

---

## CI / Automation / Agent Usage

### Non-Interactive Mode

When running in CI pipelines, Docker containers, or as part of automated tooling, use environment variables and the `--no-input` flag to prevent the CLI from prompting for input:

```bash
export JENKINS_URL=https://jenkins.example.com
export JENKINS_USER=ci-bot
export JENKINS_TOKEN=$JENKINS_API_TOKEN  # from CI secrets
export JENKINS_NO_INPUT=true

jenkins job build deploy-pipeline --param VERSION=1.2.3 --wait
```

If `--no-input` is set and a command would normally prompt (like `jenkins login`), it fails immediately with a descriptive error message instead of hanging.

### GitHub Actions Example

```yaml
jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - name: Trigger Jenkins build
        env:
          JENKINS_URL: ${{ secrets.JENKINS_URL }}
          JENKINS_USER: ${{ secrets.JENKINS_USER }}
          JENKINS_TOKEN: ${{ secrets.JENKINS_TOKEN }}
        run: |
          jenkins job build deploy-pipeline \
            --param BRANCH=${{ github.ref_name }} \
            --param COMMIT=${{ github.sha }} \
            --wait --follow --timeout 1h
```

### GitLab CI Example

```yaml
trigger-jenkins:
  script:
    - export JENKINS_URL=$JENKINS_URL
    - export JENKINS_USER=$JENKINS_USER
    - export JENKINS_TOKEN=$JENKINS_TOKEN
    - jenkins job build deploy-pipeline --param VERSION=$CI_COMMIT_TAG --wait
  variables:
    JENKINS_URL: https://jenkins.example.com
    JENKINS_USER: gitlab-ci
    # JENKINS_TOKEN should be set as a CI/CD masked variable
```

### Docker Usage

```dockerfile
FROM alpine:latest
RUN apk add --no-cache curl && \
    curl -sSL https://raw.githubusercontent.com/piyush-gambhir/jenkins-cli/main/install.sh | bash

ENV JENKINS_NO_INPUT=true
ENTRYPOINT ["jenkins"]
```

```bash
docker run --rm \
  -e JENKINS_URL=https://jenkins.example.com \
  -e JENKINS_USER=admin \
  -e JENKINS_TOKEN=mytoken \
  jenkins-cli job list -o json
```

### Quiet Mode

Use `--quiet` or `JENKINS_QUIET=true` to suppress informational output (connection messages, profile save confirmations, update notices). Only command output and errors are printed:

```bash
# In a script that parses JSON output
result=$(JENKINS_QUIET=true jenkins build get my-pipeline 42 -o json)
```

---

## Edge Cases & Troubleshooting

### Jenkins with LDAP / Active Directory

If your Jenkins uses LDAP or Active Directory for authentication:

- **API tokens work normally.** Generate the token through the Jenkins UI while logged in via LDAP. The token is stored in Jenkins, not in LDAP.
- **Your username is your LDAP username.** Use the same username you type into the Jenkins login page.
- **The user must exist in LDAP.** If your LDAP account is disabled or deleted, the API token stops working.
- **Password changes in LDAP do not affect API tokens.** Tokens are independent of your LDAP password.

```bash
# Use your LDAP username (not email, not full DN)
export JENKINS_USER=jsmith
export JENKINS_TOKEN=<token-generated-in-jenkins-ui>
jenkins whoami
```

### Jenkins Behind Reverse Proxy

**Context path issues:**

If Jenkins is deployed at a subpath (e.g., `https://ci.example.com/jenkins`), the URL must include the context path:

```bash
# This is the --prefix=/jenkins setting in Jenkins configuration
export JENKINS_URL=https://ci.example.com/jenkins
```

Common symptom: you get an HTML login page or a 404 instead of a JSON API response. If `jenkins status` returns an error about HTML content, check that your URL includes the context path.

**Proxy headers:**

Reverse proxies must forward the correct headers for Jenkins to work. If you experience authentication issues, verify the proxy configuration includes:

```nginx
# nginx example
location /jenkins {
    proxy_pass http://jenkins:8080/jenkins;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
}
```

### Jenkins in Kubernetes

**Using `kubectl port-forward` for local access:**

```bash
# Forward local port 8080 to the Jenkins pod
kubectl port-forward svc/jenkins 8080:8080 -n jenkins

# In another terminal
export JENKINS_URL=http://localhost:8080
jenkins login
```

**Using an Ingress URL:**

```bash
# Use the Ingress hostname directly
export JENKINS_URL=https://jenkins.k8s.example.com
jenkins login
```

**Service account token (in-cluster):**

If running the CLI inside a Kubernetes pod in the same cluster as Jenkins, you can use the internal service URL:

```bash
export JENKINS_URL=http://jenkins.jenkins.svc.cluster.local:8080
export JENKINS_USER=automation
export JENKINS_TOKEN=<api-token>
jenkins status
```

### Jenkins with SSO (SAML / OAuth / OpenID Connect)

When Jenkins uses SSO (via the SAML plugin, OAuth plugin, or OpenID Connect plugin):

- **API tokens still work.** SSO controls how you log in to the Jenkins web UI, but API tokens are a Jenkins-native mechanism that bypasses SSO.
- **Generate the token while authenticated via SSO.** Log in to Jenkins through your SSO provider, then navigate to `<jenkins-url>/me/configure` to create an API token.
- **The username may differ from your SSO identity.** Check your Jenkins username by clicking your name in the Jenkins UI. It might be your email, a short username, or a UUID depending on the SSO plugin configuration.

```bash
# Your Jenkins username might be your email when using SAML
export JENKINS_USER=jane.doe@example.com
export JENKINS_TOKEN=<token-from-jenkins-ui>
jenkins whoami
```

### Jenkins with Two-Factor Authentication (2FA)

If Jenkins has 2FA enabled (via a plugin):

- **API tokens bypass 2FA.** 2FA applies to interactive web logins, not to API token authentication.
- **Generate your token before enabling 2FA**, or authenticate with 2FA in the web UI first, then generate a token.

### Common Errors

| Error | Cause | Fix |
|---|---|---|
| `jenkins API error: 401 Unauthorized` | Invalid credentials | Verify username and regenerate API token |
| `jenkins API error: 403 Forbidden` | Valid credentials but insufficient permissions | Request the required permission from your Jenkins admin (see permission matrix above) |
| `No valid crumb was included in the request` | CSRF crumb mismatch | Usually caused by a proxy stripping headers. Switch to API token auth if using password. Check proxy configuration. |
| `jenkins API error: 404 Not Found` | Wrong URL, missing context path, or nonexistent resource | Verify Jenkins URL includes context path. Check job path spelling. |
| HTML error page returned instead of JSON | CLI hit a web page instead of the API | URL is likely wrong. Verify with `curl <url>/api/json`. Include context path if applicable. |
| `connection refused` | Jenkins is not running on the specified host:port | Verify Jenkins is running and listening on the expected port. Check firewall rules. |
| `TLS handshake failure` / `certificate signed by unknown authority` | Self-signed or untrusted TLS certificate | Use `--insecure` flag, or add the CA certificate to your system trust store |
| `no such host` / `DNS resolution failed` | Hostname cannot be resolved | Check DNS, VPN connection, or use IP address directly |
| `context deadline exceeded` / `timeout` | Jenkins server is slow or unreachable | Check network connectivity. Increase timeout with `--timeout` for build commands. |
| `Jenkins URL not configured` | No URL in config, env, or flags | Run `jenkins login` or set `JENKINS_URL` environment variable |
| `command 'X' is blocked in read-only mode` | Read-only mode is active | Use `--read-only=false` to override, or remove `read_only: true` from your profile |
| `interactive input required but --no-input is set` | Login command with `--no-input` | Use environment variables instead of interactive login |

### Debugging Authentication Issues

Use the `--verbose` flag to see the HTTP requests and responses:

```bash
jenkins status --verbose
```

This prints request/response details to stderr:

```
--> GET https://jenkins.example.com/api/json
    Accept: application/json
    Authorization: [REDACTED]
<-- 200 200 OK (145ms)
```

The `Authorization` header is automatically redacted in verbose output for security. If you see a `401` response, your credentials are invalid. If you see a `403`, your credentials are valid but you lack permissions.

**Testing raw connectivity:**

```bash
# Test if the URL is reachable
curl -s -o /dev/null -w "%{http_code}" https://jenkins.example.com/api/json

# Test authentication directly
curl -s -u "admin:YOUR_API_TOKEN" https://jenkins.example.com/api/json | head -c 200
```

---

## Security Best Practices

### Use API Tokens, Not Passwords

API tokens can be individually revoked without changing your password. They do not grant access to the Jenkins UI login, and they are stored as SHA-256 hashes on the server. Always prefer tokens over passwords.

### Use Least-Privilege Permissions

Create dedicated Jenkins users for CLI access with only the permissions needed for the intended operations. A CI bot that only triggers builds should not have Overall/Administer permissions.

```
# Example: minimum permissions for a CI build bot
- Overall/Read
- Job/Read (on target jobs/folders)
- Job/Build (on target jobs/folders)
```

### Store Credentials Securely

**Do:**
- Use environment variables sourced from a secrets manager (Vault, AWS Secrets Manager, 1Password CLI)
- Use CI/CD platform secret variables (GitHub Actions secrets, GitLab CI masked variables)
- Use the CLI config file (stored with `0600` permissions in `~/.config/jenkins-cli/`)

**Do not:**
- Hard-code tokens in scripts, Dockerfiles, or source code
- Pass tokens as command-line arguments in shared environments (they appear in `ps` output and shell history)
- Commit `.config/jenkins-cli/config.yaml` to version control

### Use Read-Only Mode for Safety

When using the CLI in contexts where accidental mutations are dangerous (monitoring, coding agents, exploratory scripts), enable read-only mode:

```bash
export JENKINS_READ_ONLY=true
```

Or set it in the profile:

```yaml
profiles:
  monitoring:
    url: https://jenkins.example.com
    username: monitor
    token: mytoken
    read_only: true
```

### Revoke Unused Tokens

Periodically review your API tokens at `<jenkins-url>/me/configure` and revoke any that are no longer in use. Use descriptive token names so you can identify which tool or machine each token belongs to.

### Use HTTPS

Always use HTTPS for Jenkins connections, especially when credentials are transmitted over the network. HTTP transmits Basic Auth credentials in base64 (which is trivially decoded, not encrypted).

If you must use HTTP (e.g., `localhost` development), ensure the connection does not traverse untrusted networks.

### Separate Tokens per Context

Create separate API tokens for different use cases:

| Token Name | Used By | Permissions |
|---|---|---|
| `cli-laptop` | Your personal CLI | Full user permissions |
| `ci-github-actions` | GitHub Actions | Job/Build on specific jobs |
| `monitoring-dashboard` | Grafana/monitoring | Overall/Read only |
| `deploy-bot` | Deployment automation | Job/Build on deploy jobs |

If any single token is compromised, you can revoke it without disrupting other integrations.
