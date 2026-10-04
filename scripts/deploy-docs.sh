#!/usr/bin/env bash
# Build the Next.js site in web/ and deploy it as the jenkins-cli-web Worker
# (static assets, routed at projects.piyushgambhir.com/jenkins-cli).
# Run as `bash scripts/deploy-docs.sh`.
#
# Uses CLOUDFLARE_API_TOKEN from .env.deploy.production when present, otherwise
# the local `wrangler login` session. The account is pinned in web/wrangler.jsonc.
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT_DIR/web"

DEPLOY_ENV_FILE="$ROOT_DIR/.env.deploy.production"
if [[ -f "$DEPLOY_ENV_FILE" ]]; then
  set -a
  # shellcheck disable=SC1090
  source "$DEPLOY_ENV_FILE"
  set +a
fi

pnpm install --frozen-lockfile

if [[ -z "${CLOUDFLARE_API_TOKEN:-}" ]] && ! pnpm exec wrangler whoami >/dev/null 2>&1; then
  echo "error: not logged in to wrangler. Run \`pnpm exec wrangler login\` in web/ first (or set CLOUDFLARE_API_TOKEN in $DEPLOY_ENV_FILE)." >&2
  exit 1
fi

echo "==> Building the site in web/"
pnpm build:cloudflare

echo "==> Deploying the jenkins-cli-web Worker"
pnpm deploy:cloudflare
