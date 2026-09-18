#!/usr/bin/env bash

set -euo pipefail

commit_message_file="$1"
commit_subject="$(sed -n '/^[^#]/ { p; q; }' "$commit_message_file")"
pattern='^(feat|fix|docs|style|refactor|perf|test|build|ci|chore|revert)(\([a-z0-9._/-]+\))?!?: .{1,72}$'

if [[ ! "$commit_subject" =~ $pattern ]]; then
    echo "Invalid commit message: $commit_subject" >&2
    echo "Expected: <type>[optional scope][!]: <description>" >&2
    echo "Example: feat(cache): add expiration" >&2
    exit 1
fi
