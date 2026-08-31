# google.com/goto: read `Location` (HEAD or GET)

Google Search no longer always exposes the final URL in the HTML. Results go
through `https://www.google.com/goto?url=CAES…`. That parameter is **not** an
encoded URL: you cannot decode it. Google only reveals the destination in the
HTTP **`Location`** header.

This repo explains the method, with **curl**, **Python**, **Node**, and **Go**
examples: `GET` or `HEAD` without following the redirect, then read
`Location`.

Autom article: [How to resolve google.com/goto](https://www.autom.dev/blog/google-goto-url-fix)

## The idea in one sentence

**Do not follow the redirect. Read `Location`.**

```
GET /goto?url=CAES…  HTTP/1.1
Host: www.google.com

HTTP/1.1 302 Found
Location: https://www.linkedin.com/in/satyanadella
```

`HEAD` asks for headers only (no body). `GET` with `redirect: "manual"` (or
`allow_redirects=False`) does the same read. Both work. Google sometimes
returns **402** (or another status) **with** a `Location`: read the header
even when it is not a 3xx.

## What this is not

The older wrapper `https://www.google.com/url?q=https://example.com` already
contains the target in `q=`. `goto` does not. Guessing a URL from a visible
name (“LinkedIn · Satya Nadella”) is a bad idea: you will get it wrong.
`Location` is the source of truth.

## curl: HEAD, do not follow

```bash
curl -sI --max-redirs 0 \
  -H 'Referer: https://www.google.com/search' \
  'https://www.google.com/goto?url=CAES...' \
  | grep -i '^location:'
```

`-I` = HEAD. Without `-L`, curl does not follow to the final page.
`--max-redirs 0` blocks any follow. The `Location:` line is the real URL.

Same thing with GET:

```bash
curl -sD - -o /dev/null --max-redirs 0 \
  -H 'Referer: https://www.google.com/search' \
  'https://www.google.com/goto?url=CAES...' \
  | grep -i '^location:'
```

Ready-to-run script: [`examples/curl/head.sh`](examples/curl/head.sh).

## Python

```python
import requests

goto = "https://www.google.com/goto?url=CAES..."
r = requests.head(
    goto,
    allow_redirects=False,
    headers={"Referer": "https://www.google.com/search"},
    timeout=10,
)
print(r.status_code, r.headers.get("Location"))
```

If `HEAD` returns nothing, retry with `GET` and do not follow:

```python
r = requests.get(goto, allow_redirects=False, timeout=10)
print(r.headers.get("Location"))
```

File: [`examples/python/resolve_goto.py`](examples/python/resolve_goto.py).

## Node

Use `fetch(..., { redirect: "manual" })` then
`response.headers.get("location")`. In Node, the headers are visible.

```javascript
const goto = "https://www.google.com/goto?url=CAES...";
const res = await fetch(goto, {
  method: "HEAD",
  redirect: "manual",
  headers: { Referer: "https://www.google.com/search" },
});
console.log(res.status, res.headers.get("location"));
```

File: [`examples/javascript/resolve_goto.js`](examples/javascript/resolve_goto.js).

## Go

`http.Client` follows redirects by default. Return
`http.ErrUseLastResponse` from `CheckRedirect` so the first hop stays
in place and `Location` can be read.

```go
client := &http.Client{
    CheckRedirect: func(*http.Request, []*http.Request) error {
        return http.ErrUseLastResponse
    },
}
req, _ := http.NewRequest(http.MethodHead, gotoURL, nil)
req.Header.Set("Referer", "https://www.google.com/search")
res, _ := client.Do(req)
fmt.Println(res.StatusCode, res.Header.Get("Location"))
```

File: [`examples/go/resolve_goto.go`](examples/go/resolve_goto.go).

## Why not decode CAES?

The `url=CAES…` blob is not a public URL protobuf. It is an opaque
reference. Only Google knows the destination, and it puts it in `Location`
at hop time. Do not decode it: query `/goto` and read the redirect.

## At Autom

Autom’s [Google Search](https://www.autom.dev/tools/google-search) API
resolves these `/goto` URLs on the server (same idea: no follow, read
`Location`) and already returns the final URL in `organic_results[].link`.

```bash
curl -s https://api.autom.dev/v1/google/search \
  -H "x-api-key: YOUR_KEY" \
  -H "Content-Type: application/json" \
  -d '{"query":"linkedin.com/in satya nadella","gl":"us","hl":"en"}'
```

1,000 free requests: [autom.dev/pricing](https://www.autom.dev/pricing).

## License

MIT.
