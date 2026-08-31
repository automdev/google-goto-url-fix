#!/usr/bin/env bash
# Resolve a Google /goto URL by reading the Location header.
#
# Google Search results often wrap destinations as:
#   https://www.google.com/goto?url=CAES...
# The CAES blob is opaque — you cannot decode it. Google only reveals the
# real URL in the HTTP Location header. Do not follow the redirect (-L).
#
# Usage:
#   ./examples/curl/head.sh 'https://www.google.com/goto?url=CAES...'

set -euo pipefail

GOTO="${1:-}"
if [[ -z "$GOTO" || "$GOTO" != https://www.google.com/goto* ]]; then
  echo "Usage: $0 'https://www.google.com/goto?url=CAES...'" >&2
  exit 1
fi

# HEAD: headers only, no body. --max-redirs 0 refuses any follow.
echo "== HEAD =="
curl -sI --max-redirs 0 \
  -H 'Referer: https://www.google.com/search' \
  "$GOTO" | grep -iE '^(HTTP/|location:)'

echo
# GET without -L: same Location read, no follow.
# Google sometimes returns 402 (or another non-3xx) *with* Location —
# still read the header.
echo "== GET (no follow) =="
curl -sD - -o /dev/null --max-redirs 0 \
  -H 'Referer: https://www.google.com/search' \
  "$GOTO" | grep -iE '^(HTTP/|location:)'
