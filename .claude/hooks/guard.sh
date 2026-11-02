#!/usr/bin/env bash
# PreToolUse guard. Blocks secret access, deploy/destructive commands and edits
# to landed migrations. Exit 2 = block, stderr is shown to Claude.
set -uo pipefail
input="$(cat)"
tool="$(jq -r '.tool_name // empty' <<<"$input")"
root="${CLAUDE_PROJECT_DIR:-$(git rev-parse --show-toplevel)}"

block() { echo "BLOCKED by .claude/hooks/guard.sh: $1" >&2; exit 2; }

# .env, .env.infra, *.tfvars, terraform state (but not *.example)
is_secret_path() {
  case "$(basename "$1")" in
    *.example) return 1 ;;
    .env|.env.*|*.tfvars|terraform.tfstate*) return 0 ;;
  esac
  return 1
}

case "$tool" in
  Bash)
    cmd="$(jq -r '.tool_input.command // empty' <<<"$input")"

    # Secret files via the shell (permissions.deny only covers file tools).
    if grep -Eq '(^|[[:space:]/"'"'"'=])\.env(\.[A-Za-z0-9_-]+)?([[:space:]"'"'"';|&)]|$)|\.tfvars|terraform\.tfstate' <<<"$cmd" \
       && ! grep -Eq '\.env\.example|\.env\.infra\.example' <<<"$cmd"; then
      block "command touches a secret file (.env*, *.tfvars, tfstate). Ask the user to run it."
    fi

    grep -Eq '(^|[[:space:];&|(])(terraform|tofu)[[:space:]]+(apply|destroy)' <<<"$cmd" \
      && block "terraform apply/destroy is owner-run only."
    grep -Eq '(^|[[:space:];&|(])(fly|flyctl)[[:space:]]+(deploy|secrets|machine|apps[[:space:]]+destroy)' <<<"$cmd" \
      && block "fly deploy/secrets/machine changes are owner-run; CI deploys on merge to main."
    grep -Eq 'git[[:space:]]+push([^;&|]*)[[:space:]](-f|--force|--force-with-lease)' <<<"$cmd" \
      && block "force-push is not allowed."
    grep -Eq 'git[[:space:]]+push([^;&|]*)[[:space:]](origin[[:space:]]+)?(HEAD:)?(refs/heads/)?main([[:space:]]|$)' <<<"$cmd" \
      && block "never push to main; push a branch and open a PR."
    grep -Eq 'gh[[:space:]]+pr[[:space:]]+merge' <<<"$cmd" \
      && block "merging PRs is a human step."
    if grep -Eq 'migrate[[:space:]].*(down|drop|force)' <<<"$cmd" \
       && ! grep -Eq 'localhost|127\.0\.0\.1|\$\{?DATABASE_URL' <<<"$cmd"; then
      block "migrate down/drop/force only against a localhost database."
    fi
    if grep -Eq 'migrate[[:space:]].*(down|drop|force)' <<<"$cmd" && grep -Eq 'neon\.tech|fly\.dev' <<<"$cmd"; then
      block "migrate down/drop/force against a remote database."
    fi
    ;;

  Read|Edit|Write|MultiEdit|NotebookEdit)
    path="$(jq -r '.tool_input.file_path // .tool_input.notebook_path // empty' <<<"$input")"
    [ -z "$path" ] && exit 0
    is_secret_path "$path" && block "$path is a secret file."

    # Landed migrations are append-only: block edits to any migration present in HEAD.
    if [ "$tool" != "Read" ]; then
      rel="${path#"$root"/}"
      case "$rel" in
        backend/migrations/*.sql)
          if git -C "$root" cat-file -e "HEAD:$rel" 2>/dev/null; then
            block "$rel is already committed. Migrations are append-only: add the next-numbered migration instead."
          fi ;;
      esac
    fi
    ;;
esac
exit 0
