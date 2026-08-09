#!/usr/bin/env bash
# Stop hook: lint AND test only the packages this session actually touched.
#
# Why not plain `make lint && make test`: checking the whole module after every
# turn is slow, and a pre-existing failure in an untouched package would block
# completion in a way the session cannot fix by correcting its own work.
#
# Why tests and not just lint: the failure mode this repo has already hit is a
# passing-looking assertion that doesn't assert what it reads like
# (assert.ErrorIs against a freshly built fmt.Errorf). Lint cannot see that.
#
# -race is deliberately omitted here for turn latency. `make test` runs it and
# is the real definition-of-done gate.
#
# Exit 0 => allow completion. Exit 2 => block, stderr goes back to Claude.
set -uo pipefail

input="$(cat)"

# Stop hooks re-fire after they block. Without this guard, a failure the model
# cannot fix would loop forever.
if [ "$(printf '%s' "$input" | jq -r '.stop_hook_active // false' 2>/dev/null)" = "true" ]; then
  exit 0
fi

root="${CLAUDE_PROJECT_DIR:-$(git rev-parse --show-toplevel 2>/dev/null || pwd)}"
cd "$root" 2>/dev/null || exit 0

# Packages containing uncommitted .go changes (staged, unstaged, or untracked).
# Handles renames ("R  old -> new") by taking the destination path.
pkgs=()
while IFS= read -r line; do
  [ -n "$line" ] && pkgs+=("$line")
done < <(
  git status --porcelain --untracked-files=all 2>/dev/null \
    | sed 's/^...//; s/^.* -> //' \
    | grep -E '\.go$' \
    | xargs -r -n1 dirname \
    | sort -u \
    | awk '{ print ($0 == ".") ? "." : "./" $0 }'
)

# No Go changes this turn (docs-only edit, etc.) — nothing to check.
[ "${#pkgs[@]}" -eq 0 ] && exit 0

fail=""

if command -v gofmt >/dev/null 2>&1; then
  if unformatted="$(gofmt -l "${pkgs[@]}" 2>/dev/null)" && [ -n "$unformatted" ]; then
    fail+="Not gofmt-clean:"$'\n'"$unformatted"$'\n\n'
  fi
fi

if command -v golangci-lint >/dev/null 2>&1; then
  if ! out="$(golangci-lint run "${pkgs[@]}" 2>&1)"; then
    fail+="Lint failures:"$'\n'"$out"$'\n\n'
  fi
fi

if command -v go >/dev/null 2>&1; then
  if ! out="$(go test -count=1 "${pkgs[@]}" 2>&1)"; then
    fail+="Test failures:"$'\n'"$out"$'\n\n'
  fi
fi

if [ -n "$fail" ]; then
  {
    echo "pre-done: problems in the packages you changed. Fix these before finishing."
    echo
    printf '%s' "$fail"
  } >&2
  exit 2
fi

exit 0
