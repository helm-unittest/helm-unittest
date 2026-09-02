#!/bin/sh
set -eu

expression=""
key=""
value=""

for arg in "$@"; do
  if printf '%s\n' "$arg" | grep -E '^\.metadata\.annotations\.[A-Za-z0-9._-]+="[^"]*"$' > /dev/null 2>&1; then
    expression="$arg"
    key=$(printf '%s\n' "$arg" | sed 's/^\.metadata\.annotations\.\([A-Za-z0-9._-]*\)="\([^"]*\)"$/\1/')
    value=$(printf '%s\n' "$arg" | sed 's/^\.metadata\.annotations\.\([A-Za-z0-9._-]*\)="\([^"]*\)"$/\2/')
    break
  fi
done

input="$(cat)"

if [ -z "$expression" ]; then
  printf '%s' "$input"
  exit 0
fi

has_key=0
if printf '%s\n' "$input" | grep -E "^[[:space:]]*${key}:[[:space:]]*" > /dev/null 2>&1; then
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
