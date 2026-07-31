#!/usr/bin/env bash

set -euo pipefail

repo_root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
cd "$repo_root"

bash -n scripts/bootstrap.sh scripts/install.sh scripts/install-skills.sh
latest_plan=$(
  scripts/install.sh \
    --install-dir /tmp/obs-cli-install-eval-bin \
    --dry-run
)
grep -Fq 'release:     latest' <<<"$latest_plan"

scripts/install.sh \
  --version v1.0.0-rc.1 \
  --install-dir /tmp/obs-cli-install-eval-bin \
  --dry-run \
  >/dev/null

eval_root=$(mktemp -d "${TMPDIR:-/tmp}/obs-cli-install-eval.XXXXXX")
trap 'rm -rf "$eval_root"' EXIT
skills_root="$eval_root/skills"

mkdir -p "$eval_root/fake-bin"
printf '%s\n' \
  '#!/usr/bin/env bash' \
  'printf "%s" "https://github.com/andy-neoaira/obs-cli/releases/tag/v9.8.7"' \
  >"$eval_root/fake-bin/curl"
chmod 0755 "$eval_root/fake-bin/curl"
latest_bootstrap_plan=$(
  PATH="$eval_root/fake-bin:$PATH" scripts/bootstrap.sh \
    --agent cursor \
    --install-dir "$eval_root/latest-bootstrap-bin" \
    --skills-dir "$eval_root/latest-bootstrap-skills" \
    --source . \
    --dry-run
)
grep -Fq 'Resolved latest obs-cli Release: v9.8.7' \
  <<<"$latest_bootstrap_plan"
grep -Fq 'release:     v9.8.7' <<<"$latest_bootstrap_plan"
grep -Fq 'version:     v9.8.7' <<<"$latest_bootstrap_plan"

scripts/install-skills.sh \
  --agent custom \
  --dest "$skills_root" \
  --source . \
  --version v1.0.0-rc.1 \
  >/dev/null

installed=$(
  find "$skills_root" -mindepth 2 -maxdepth 2 -name SKILL.md -print |
    wc -l |
    tr -d ' '
)
[[ "$installed" == "11" ]]
[[ ! -e "$skills_root/_template" ]]
[[ ! -e "$skills_root/evals" ]]

scripts/install-skills.sh \
  --agent custom \
  --dest "$skills_root" \
  --source . \
  --version v1.0.0 \
  --upgrade \
  --dry-run \
  >/dev/null

scripts/install-skills.sh \
  --agent custom \
  --dest "$skills_root" \
  --source . \
  --version v1.0.0 \
  --upgrade \
  >/dev/null

grep -Fq '"version": "v1.0.0"' \
  "$skills_root/obsidian-capture/.obs-cli-managed.json"

printf '\nlocal edit\n' >>"$skills_root/obsidian-capture/SKILL.md"
if scripts/install-skills.sh \
  --agent custom \
  --dest "$skills_root" \
  --source . \
  --version v1.0.1 \
  --upgrade \
  >"$eval_root/conflict.out" 2>"$eval_root/conflict.err"; then
  printf 'install eval: modified Skill upgrade unexpectedly succeeded\n' >&2
  exit 1
fi
grep -Fq 'local modifications' "$eval_root/conflict.err"

agent_home="$eval_root/agent-home"
mkdir -p "$agent_home"
HOME="$agent_home" CODEX_HOME="$agent_home/codex-home" \
  scripts/install-skills.sh --agent codex --source . --version v1.0.0 >/dev/null
HOME="$agent_home" \
  scripts/install-skills.sh --agent claude-code --source . --version v1.0.0 >/dev/null
HOME="$agent_home" XDG_CONFIG_HOME="$agent_home/xdg" \
  scripts/install-skills.sh --agent opencode --source . --version v1.0.0 >/dev/null
HOME="$agent_home" \
  scripts/install-skills.sh --agent cursor --source . --version v1.0.0 >/dev/null
HOME="$agent_home" KIMI_CODE_HOME="$agent_home/kimi-home" \
  scripts/install-skills.sh --agent kimi-code --source . --version v1.0.0 >/dev/null

[[ -f "$agent_home/codex-home/skills/obsidian-capture/SKILL.md" ]]
[[ -f "$agent_home/.claude/skills/obsidian-capture/SKILL.md" ]]
[[ -f "$agent_home/xdg/opencode/skills/obsidian-capture/SKILL.md" ]]
[[ -f "$agent_home/.cursor/skills/obsidian-capture/SKILL.md" ]]
[[ -f "$agent_home/kimi-home/skills/obsidian-capture/SKILL.md" ]]

scripts/bootstrap.sh \
  --agent cursor \
  --version v1.0.0-rc.1 \
  --install-dir "$eval_root/bootstrap-bin" \
  --skills-dir "$eval_root/bootstrap-skills" \
  --source . \
  --dry-run \
  >/dev/null

printf 'installer evals passed\n'
