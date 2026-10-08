#!/usr/bin/env node
/*

Copyright (c) 2018 - 2026 PhotoPrism UG. All rights reserved.

Reports HTML bindings and DOM HTML sinks in the frontend code that were
not marked as reviewed, and exits with status 1 when it finds any.

*/

import fs from "node:fs";
import path from "node:path";
import process from "node:process";
import { fileURLToPath } from "node:url";

// CODE_EXTENSIONS lists the file types that can contain a binding or sink.
const CODE_EXTENSIONS = new Set([".js", ".mjs", ".cjs", ".vue", ".html", ".htm"]);

// SRCDOC matches the srcdoc property and attribute name, which HTML compares case-insensitively.
const SRCDOC = "[sS][rR][cC][dD][oO][cC]";

// HTML_PROPS matches the names of properties whose value is parsed as HTML.
const HTML_PROPS = `(?:(?:inner|outer)HTML|${SRCDOC})`;

// HTML_BINDINGS matches template bindings that render a value as HTML: v-html and innerHTML/outerHTML/srcdoc props.
const HTML_BINDINGS = [/\bv-html\s*=/, new RegExp(`(?:^|[\\s<])(?:v-bind)?:${HTML_PROPS}(?:\\.prop)?\\s*=`)];

// BINDING_REVIEWED matches a comment that marks the binding on the next line as reviewed, with a reason.
const BINDING_REVIEWED = /(?:<!--|\/\/|\/\*)\s*eslint-disable-next-line\s+vue\/no-v-html\s+--\s+(?!\*\/)[^\s-]/;

// DOM_SINKS matches assignments (not comparisons), object properties, and calls that parse a string as HTML.
const DOM_SINKS = [
  new RegExp(`\\.${HTML_PROPS}\\s*(?:\\+|\\|\\||\\?\\?|&&)?=(?!=)`),
  new RegExp(`\\[\\s*["'\`]${HTML_PROPS}["'\`]\\s*\\]\\s*(?:\\+|\\|\\||\\?\\?|&&)?=(?!=)`),
  new RegExp(`(?:^|[\\s{,(])["'\`]?${HTML_PROPS}["'\`]?\\s*:`),
  new RegExp(`(?:^|[{,])\\s*${HTML_PROPS}\\s*(?:[,}]|$)`),
  new RegExp(`setAttribute\\s*\\(\\s*["'\`]${SRCDOC}["'\`]`),
  /insertAdjacentHTML\s*\(/,
  /document\.write(?:ln)?\s*\(/,
  /\b(?:createContextualFragment|setHTMLUnsafe|parseHTMLUnsafe|parseFromString)\s*\(/,
];

// EDITION_OVERLAYS lists the edition frontend directories, which are compiled into the same bundle
// when present.
const EDITION_OVERLAYS = ["plus", "pro", "portal"];

// SKIPPED_DIRS lists directories that hold tests or dependencies rather than shipped code.
const SKIPPED_DIRS = new Set(["node_modules", "tests"]);

// SINK_REVIEWED matches sink lines that are reviewed or only clear an element.
const SINK_REVIEWED = [/security-reviewed/, /^[^=]*\.innerHTML\s*=\s*(?:""|'')\s*;?\s*$/];

// findUnreviewedVHtml returns the line numbers of HTML bindings whose directly preceding
// line is not a reviewed eslint-disable comment for vue/no-v-html.
export function findUnreviewedVHtml(text) {
  const lines = text.split("\n");
  const found = [];

  lines.forEach((line, i) => {
    if (HTML_BINDINGS.some((re) => re.test(line)) && (i === 0 || !BINDING_REVIEWED.test(lines[i - 1]))) {
      found.push(i + 1);
    }
  });

  return found;
}

// findUnreviewedSinks returns the line numbers of DOM HTML sinks not marked as reviewed.
export function findUnreviewedSinks(text) {
  const found = [];

  text.split("\n").forEach((line, i) => {
    if (DOM_SINKS.some((re) => re.test(line)) && !SINK_REVIEWED.some((re) => re.test(line))) {
      found.push(i + 1);
    }
  });

  return found;
}

// listFiles returns the code files below dir, sorted by path.
export function listFiles(dir) {
  const files = [];

  for (const entry of fs.readdirSync(dir, { withFileTypes: true })) {
    const p = path.join(dir, entry.name);
    if (entry.isDirectory()) {
      if (!SKIPPED_DIRS.has(entry.name)) {
        files.push(...listFiles(p));
      }
    } else if (entry.isFile() && CODE_EXTENSIONS.has(path.extname(entry.name))) {
      files.push(p);
    }
  }

  return files.sort();
}

// defaultDirs returns the frontend sources and the edition overlays that exist next to it.
export function defaultDirs(frontendDir) {
  const overlays = EDITION_OVERLAYS.map((edition) => path.resolve(frontendDir, "..", edition, "frontend")).filter((dir) => fs.existsSync(dir));

  return [path.join(frontendDir, "src"), ...overlays];
}

// scan returns one "file:line: kind" entry per unreviewed finding in the given directories.
export function scan(dirs) {
  const findings = [];

  for (const dir of dirs) {
    for (const file of listFiles(dir)) {
      const text = fs.readFileSync(file, "utf8");
      for (const line of findUnreviewedVHtml(text)) {
        findings.push(`${file}:${line}: HTML binding without a reviewed eslint-disable comment`);
      }
      for (const line of findUnreviewedSinks(text)) {
        findings.push(`${file}:${line}: DOM HTML sink without a security-reviewed note`);
      }
    }
  }

  return findings;
}

if (process.argv[1] === fileURLToPath(import.meta.url)) {
  const frontendDir = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..");
  const dirs = process.argv.length > 2 ? process.argv.slice(2) : defaultDirs(frontendDir).map((dir) => path.relative(process.cwd(), dir) || ".");
  const findings = scan(dirs);

  if (findings.length > 0) {
    console.log("ERROR: unreviewed HTML sinks detected; prefer text bindings or $util.sanitizeHtml()");
    console.log(findings.join("\n"));
    process.exit(1);
  }

  console.log("OK: No unreviewed HTML bindings or DOM HTML sinks detected.");
}
