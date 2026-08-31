#!/usr/bin/env node
/**
 * Resolve a Google /goto URL by reading the Location header.
 *
 * Google Search results often wrap destinations as:
 *   https://www.google.com/goto?url=CAES...
 * The CAES blob is opaque — you cannot decode it. Google only reveals the
 * real URL in the HTTP Location header.
 *
 * Use fetch with redirect: "manual". Do not follow the hop. Try HEAD
 * first, then GET. Google sometimes returns 402 (or another non-3xx)
 * with Location: still read the header. In Node, Location is visible.
 *
 * Usage:
 *   node examples/javascript/resolve_goto.js 'https://www.google.com/goto?url=CAES...'
 */

function isGoogleGoto(raw) {
  try {
    const url = new URL(raw);
    return (
      url.protocol === "https:" &&
      (url.hostname === "google.com" || url.hostname === "www.google.com") &&
      (url.pathname === "/goto" || url.pathname.startsWith("/goto/"))
    );
  } catch {
    return false;
  }
}

async function resolveGoto(url) {
  if (!isGoogleGoto(url)) {
    throw new Error("expected https://www.google.com/goto?url=...");
  }

  // Referer matches a real search click; some hops expect it.
  const headers = { Referer: "https://www.google.com/search" };

  for (const method of ["HEAD", "GET"]) {
    const res = await fetch(url, {
      method,
      redirect: "manual",
      headers,
    });
    const location = res.headers.get("location");
    if (location) {
      // Location may be relative; resolve it against the request URL.
      return new URL(location, url).href;
    }
  }
  return null;
}

const input = process.argv[2];
if (!input) {
  console.error(
    "Usage: node resolve_goto.js 'https://www.google.com/goto?url=CAES...'",
  );
  process.exit(2);
}

resolveGoto(input)
  .then((dest) => {
    if (!dest) {
      console.error("no Location header");
      process.exit(1);
    }
    console.log(dest);
  })
  .catch((err) => {
    console.error(err.message);
    process.exit(1);
  });
