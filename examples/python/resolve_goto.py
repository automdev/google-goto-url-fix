#!/usr/bin/env python3
"""Resolve a Google /goto URL by reading the Location header.

Google Search results often wrap destinations as::

    https://www.google.com/goto?url=CAES...

The CAES blob is opaque — you cannot decode it. Google only reveals the
real URL in the HTTP ``Location`` header.

Do not follow the redirect (``allow_redirects=False``). Try HEAD first,
then GET. Google sometimes returns 402 (or another non-3xx) *with*
``Location``: still read the header.

Usage::

    python3 examples/python/resolve_goto.py 'https://www.google.com/goto?url=CAES...'
"""

from __future__ import annotations

import sys
import urllib.parse

import requests

GOOGLE_GOTO_HOSTS = {"google.com", "www.google.com"}


def is_google_goto(url: str) -> bool:
    """Return True if *url* is an https://www.google.com/goto… link."""
    try:
        parsed = urllib.parse.urlparse(url)
    except ValueError:
        return False
    host = (parsed.netloc or "").lower()
    path = parsed.path or ""
    return (
        parsed.scheme == "https"
        and host in GOOGLE_GOTO_HOSTS
        and (path == "/goto" or path.startswith("/goto/"))
    )


def resolve_goto(url: str, timeout: int = 10) -> str | None:
    """Return the destination from Location, or None if Google sent none."""
    if not is_google_goto(url):
        raise ValueError("expected https://www.google.com/goto?url=...")

    # Referer matches a real search click; some hops expect it.
    headers = {"Referer": "https://www.google.com/search"}
    for method in ("head", "get"):
        response = requests.request(
            method,
            url,
            allow_redirects=False,
            headers=headers,
            timeout=timeout,
        )
        location = response.headers.get("Location")
        if location:
            # Location may be relative; join against the request URL.
            return urllib.parse.urljoin(url, location)
    return None


def main() -> int:
    if len(sys.argv) != 2:
        print(
            "Usage: resolve_goto.py 'https://www.google.com/goto?url=CAES...'",
            file=sys.stderr,
        )
        return 2
    dest = resolve_goto(sys.argv[1])
    if not dest:
        print("no Location header", file=sys.stderr)
        return 1
    print(dest)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
