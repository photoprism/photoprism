#!/usr/bin/env node
/*

Copyright (c) 2018 - 2026 PhotoPrism UG. All rights reserved.

Pre-compresses bundled frontend assets (JS, CSS, fonts, JSON, SVG, …) into
`.gz` and `.zst` siblings so that the Go static handler in
`internal/server/routes_static.go` can serve them verbatim and skip the
runtime compression middleware on the hot static-asset path.

Skipped extensions are formats that are already compressed (woff2, webp, …)
or binary blobs where compression adds CPU without meaningful savings.

Run automatically as the npm `postbuild` hook for `npm run build`, so that
`make build-js` produces precompressed siblings without a separate step.

*/

"use strict";

const fs = require("fs");
const path = require("path");
const zlib = require("zlib");
const process = require("process");

// COMPRESSIBLE_EXTENSIONS lists file suffixes whose contents are worth
// precompressing. Anything not on this list is left alone — including
// already-compressed binaries (woff2, png, jpg, …) and out-of-scope formats
// (mp3, mp4, zip, gz, zst).
const COMPRESSIBLE_EXTENSIONS = new Set([
  ".js",
  ".mjs",
  ".cjs",
  ".css",
  ".html",
  ".htm",
  ".xml",
  ".svg",
  ".json",
  ".map",
  ".txt",
  ".webmanifest",
  ".eot",
  ".ttf",
  ".woff",
]);

// MIN_BYTES skips tiny files where the encoder framing overhead would
// outweigh any saving and where the per-request CPU cost of runtime
// compression is already negligible.
const MIN_BYTES = 1024;

// MIN_RATIO ensures we only keep a precompressed sibling when it's at least
// ~5% smaller than the source. Files that don't compress meaningfully
// (already encoded, near-random) are not worth a separate disk read.
const MIN_RATIO = 0.95;

const DEFAULT_TARGET = path.join(__dirname, "..", "..", "assets", "static", "build");

// `--clean` removes any precompressed siblings under the target directory
// without producing new ones. Used by the watch script so stale bundles
// from a previous `make build-js` don't get served while webpack rebuilds
// identity assets in development.
const args = process.argv.slice(2);
let cleanOnly = false;
const positional = [];
for (const arg of args) {
  if (arg === "--clean") {
    cleanOnly = true;
  } else if (arg === "--") {
    // pass-through separator, ignore
  } else if (arg.startsWith("-")) {
    console.error(`[precompress] error: unknown flag ${arg}`);
    process.exit(2);
  } else {
    positional.push(arg);
  }
}
const targetDir = positional[0] || DEFAULT_TARGET;

if (cleanOnly) {
  const removed = cleanSiblings(targetDir);
  console.log(`[precompress] removed ${removed} stale sibling(s) under ${path.relative(process.cwd(), targetDir) || "."}`);
  process.exit(0);
}

// Hard-fail on Node runtimes without built-in zstd support so the build
// produces both encodings as required by the static-handler contract.
if (typeof zlib.zstdCompressSync !== "function") {
  console.error(
    `[precompress] error: this Node.js (${process.version}) lacks built-in zstd support. ` +
      "Upgrade to Node 22.15+ or 24.x — see frontend/package.json engines."
  );
  process.exit(1);
}

const ZSTD_LEVEL = 19;
const GZIP_LEVEL = zlib.constants.Z_BEST_COMPRESSION; // level 9

let scanned = 0;
let compressed = 0;
let skippedSmall = 0;
let skippedRatio = 0;
let skippedExt = 0;
let savedGzip = 0;
let savedZstd = 0;

walk(targetDir);

console.log(
  `[precompress] processed ${compressed}/${scanned} files in ${path.relative(process.cwd(), targetDir) || "."}: ` +
    `gzip saved ${formatBytes(savedGzip)}, zstd saved ${formatBytes(savedZstd)} ` +
    `(skipped: ext=${skippedExt}, small=${skippedSmall}, ratio=${skippedRatio})`
);

function walk(dir) {
  let entries;
  try {
    entries = fs.readdirSync(dir, { withFileTypes: true });
  } catch (err) {
    if (err.code === "ENOENT") {
      console.error(`[precompress] warning: target directory ${dir} does not exist; nothing to do.`);
      return;
    }
    throw err;
  }

  for (const entry of entries) {
    const full = path.join(dir, entry.name);
    if (entry.isDirectory()) {
      walk(full);
      continue;
    }
    if (!entry.isFile()) {
      continue;
    }
    if (entry.name.endsWith(".gz") || entry.name.endsWith(".zst") || entry.name.endsWith(".br")) {
      // Already a precompressed sibling from a previous build; skip it
      // so we don't recursively encode siblings.
      continue;
    }
    scanned += 1;

    const ext = path.extname(entry.name).toLowerCase();
    if (!COMPRESSIBLE_EXTENSIONS.has(ext)) {
      skippedExt += 1;
      continue;
    }

    let buf;
    try {
      buf = fs.readFileSync(full);
    } catch (err) {
      console.error(`[precompress] error: failed to read ${full}: ${err.message}`);
      process.exit(1);
    }

    if (buf.length < MIN_BYTES) {
      skippedSmall += 1;
      continue;
    }

    const gz = zlib.gzipSync(buf, { level: GZIP_LEVEL });
    const zs = zlib.zstdCompressSync(buf, {
      params: { [zlib.constants.ZSTD_c_compressionLevel]: ZSTD_LEVEL },
    });

    // Round-trip the encoded bytes so a corrupt sibling never reaches disk.
    if (!zlib.gunzipSync(gz).equals(buf)) {
      console.error(`[precompress] error: gzip round-trip mismatch for ${full}`);
      process.exit(1);
    }
    if (!zlib.zstdDecompressSync(zs).equals(buf)) {
      console.error(`[precompress] error: zstd round-trip mismatch for ${full}`);
      process.exit(1);
    }

    if (gz.length / buf.length > MIN_RATIO && zs.length / buf.length > MIN_RATIO) {
      // Compression doesn't actually shrink this file (e.g. tightly packed
      // assets); leave it as identity-only. Remove any stale siblings from
      // a previous build.
      removeStale(full);
      skippedRatio += 1;
      continue;
    }

    if (gz.length / buf.length <= MIN_RATIO) {
      writeSibling(full + ".gz", gz);
      savedGzip += buf.length - gz.length;
    } else {
      removeIfExists(full + ".gz");
    }
    if (zs.length / buf.length <= MIN_RATIO) {
      writeSibling(full + ".zst", zs);
      savedZstd += buf.length - zs.length;
    } else {
      removeIfExists(full + ".zst");
    }

    compressed += 1;
  }
}

function writeSibling(target, contents) {
  fs.writeFileSync(target, contents, { mode: 0o644 });
}

function removeStale(originalPath) {
  removeIfExists(originalPath + ".gz");
  removeIfExists(originalPath + ".zst");
}

function removeIfExists(p) {
  try {
    fs.unlinkSync(p);
  } catch (err) {
    if (err.code !== "ENOENT") {
      throw err;
    }
  }
}

function cleanSiblings(dir) {
  let count = 0;
  let entries;
  try {
    entries = fs.readdirSync(dir, { withFileTypes: true });
  } catch (err) {
    if (err.code === "ENOENT") {
      return 0;
    }
    throw err;
  }
  for (const entry of entries) {
    const full = path.join(dir, entry.name);
    if (entry.isDirectory()) {
      count += cleanSiblings(full);
      continue;
    }
    if (!entry.isFile()) {
      continue;
    }
    if (entry.name.endsWith(".gz") || entry.name.endsWith(".zst")) {
      try {
        fs.unlinkSync(full);
        count += 1;
      } catch (err) {
        if (err.code !== "ENOENT") {
          throw err;
        }
      }
    }
  }
  return count;
}

function formatBytes(n) {
  if (n < 1024) {
    return `${n} B`;
  }
  const units = ["KiB", "MiB", "GiB"];
  let value = n / 1024;
  let unit = units[0];
  for (let i = 1; i < units.length && value >= 1024; i += 1) {
    value /= 1024;
    unit = units[i];
  }
  return `${value.toFixed(1)} ${unit}`;
}
