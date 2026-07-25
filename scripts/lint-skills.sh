#!/bin/sh

set -eu

strict=false
if [ "${1:-}" = "--strict" ]; then
  strict=true
  shift
fi

if [ "$#" -gt 0 ]; then
  files="$*"
  explicit=true
else
  files=$(find skills -mindepth 2 -maxdepth 2 -name SKILL.md -print | sort)
  explicit=false
fi

required_sections='触发条件
非触发条件
输入
Capability 前置检查
读取范围
写入范围
执行生命周期
授权策略
冲突、重试与幂等
结果摘要'

errors=0
warnings=0
checked=0

report_error() {
  printf 'ERROR %s: %s\n' "$1" "$2" >&2
  errors=$((errors + 1))
}

for file in $files; do
  if [ ! -f "$file" ]; then
    report_error "$file" "文件不存在"
    continue
  fi

  if ! grep -q '^## Capability 前置检查$' "$file"; then
    if [ "$strict" = true ] || [ "$explicit" = true ] || [ "$file" = "skills/_template/SKILL.md" ]; then
      report_error "$file" "未采用 V2 Skill 契约（缺少 Capability 前置检查）"
    else
      printf 'WARN  %s: V1 Skill，等待 P3 迁移\n' "$file" >&2
      warnings=$((warnings + 1))
    fi
    continue
  fi

  checked=$((checked + 1))

  frontmatter=$(awk '
    NR == 1 && $0 == "---" { in_frontmatter = 1; next }
    in_frontmatter && $0 == "---" { exit }
    in_frontmatter { print }
  ' "$file")

  if [ -z "$frontmatter" ]; then
    report_error "$file" "缺少 YAML frontmatter"
  else
    printf '%s\n' "$frontmatter" | grep -q '^name: [a-z0-9][a-z0-9-]*$' ||
      report_error "$file" "name 必须使用小写字母、数字和连字符"
    printf '%s\n' "$frontmatter" | grep -q '^description: .\+' ||
      report_error "$file" "缺少非空 description"
    extra_keys=$(printf '%s\n' "$frontmatter" | sed -n 's/^\([A-Za-z_][A-Za-z0-9_-]*\):.*/\1/p' | grep -Ev '^(name|description)$' || true)
    if [ -n "$extra_keys" ]; then
      report_error "$file" "frontmatter 仅允许 name 和 description"
    fi
  fi

  old_ifs=$IFS
  IFS='
'
  for section in $required_sections; do
    grep -q "^## $section$" "$file" || report_error "$file" "缺少章节：$section"
  done
  IFS=$old_ifs

  grep -q 'obs-cli capabilities --output json --require' "$file" ||
    report_error "$file" "缺少 capability 检查命令"
  grep -q 'discover.*read.*plan.*authorize.*apply.*verify' "$file" ||
    report_error "$file" "缺少完整执行生命周期"
  grep -q -- '--if-match' "$file" ||
    report_error "$file" "缺少 revision 前置条件规则"
  grep -q 'REVISION_CONFLICT' "$file" ||
    report_error "$file" "缺少 revision 冲突规则"
  grep -qi 'dry-run' "$file" ||
    report_error "$file" "缺少 dry-run 规则"
  grep -q 'stdin' "$file" ||
    report_error "$file" "缺少长文本 stdin/文件规则"
  grep -q 'verify' "$file" ||
    report_error "$file" "缺少写后验证规则"
done

printf 'Skill lint: %s V2 checked, %s legacy warning(s), %s error(s)\n' "$checked" "$warnings" "$errors"
test "$errors" -eq 0
