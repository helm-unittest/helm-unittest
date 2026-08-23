#!/usr/bin/env bash
set -euo pipefail

expression=""
for arg in "$@"; do
  if [[ "$arg" =~ ^\.metadata\.annotations\.([A-Za-z0-9._-]+)=\"([^\"]*)\"$ ]]; then
    expression="$arg"
    key="${BASH_REMATCH[1]}"
    value="${BASH_REMATCH[2]}"
    break
  fi
done

input="$(cat)"

if [[ -z "$expression" ]]; then
  printf '%s' "$input"
  exit 0
fi

has_key=0
if printf '%s\n' "$input" | grep -Eq "^[[:space:]]*${key}:[[:space:]]*"; then
  has_key=1
fi

printf '%s\n' "$input" | awk -v key="$key" -v value="$value" -v has_key="$has_key" '
  {
    if ($0 ~ "^[[:space:]]*" key ":[[:space:]]*") {
      prefix = substr($0, 1, index($0, key ":") - 1)
      print prefix key ": \"" value "\""
      next
    }

    print

    if (has_key == 0 && $0 ~ /^[[:space:]]*annotations:[[:space:]]*$/) {
      match($0, /^[[:space:]]*/)
      print substr($0, RSTART, RLENGTH) "  " key ": \"" value "\""
    }
  }
'
