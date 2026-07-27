#!/usr/bin/env node
//
// gettext-lint.mjs — validate PhotoPrism's gettext catalogs for objective,
// non-linguistic defects that break rendering at runtime:
//
//   1. Placeholder integrity — the set of substitution placeholders in a
//      translation (msgstr) must match the source string (msgid). The frontend
//      (vue3-gettext) uses named `%{name}` placeholders; the backend (gotext)
//      uses printf verbs (`%s`, `%d`, `%.1f`, …). A translation that renames,
//      drops, or corrupts a placeholder silently fails to interpolate.
//   2. Whitespace / newline edges — leading/trailing spaces and a trailing
//      newline are structural; msgstr should mirror msgid at its edges.
//   3. Catalog coverage — every literal msgid passed to `$gettext` in the Vue/JS
//      sources must exist in translations.pot. The extractor drops some call and
//      markup forms silently, and a string it never sees renders English in every
//      locale with no other symptom.
//
// It also shells out to `msgfmt -c --check-format` to surface gettext's own
// c-format fatals. Human summary to stderr, JSONL findings to stdout, and a
// non-zero exit when any finding exists (so it can gate CI if desired).
//
// Dependency-free (Node stdlib only). Run from the repository root:
//
//   node scripts/gettext-lint.mjs            # lint everything
//   node scripts/gettext-lint.mjs --json     # JSONL findings only (no summary)
//
// Frontend catalogs: frontend/src/locales/*.po
// Backend catalogs:  assets/locales/<locale>/default.po

import { readFileSync, readdirSync, existsSync } from "node:fs";
import { spawnSync } from "node:child_process";
import { basename, join } from "node:path";

const FRONTEND_DIR = "frontend/src/locales";
const BACKEND_DIR = "assets/locales";
// Extraction sources, mirroring scripts/gettext-extract.sh: the CE frontend plus
// whichever private edition overlays are present in this clone.
const SOURCE_DIRS = ["frontend/src", "plus/frontend", "pro/frontend", "portal/frontend"];
const POT_PATH = join(FRONTEND_DIR, "translations.pot");
const jsonOnly = process.argv.includes("--json");

// --- Minimal, multiline-aware .po parser -----------------------------------
// Returns [{ msgctxt, msgid, msgstr, plurals: [msgstr[0], msgstr[1], …] }, …],
// excluding the header entry (empty msgid). Escapes are kept verbatim, so an
// escaped newline reads as the two characters `\` + `n`.
function parsePo(path) {
  const entries = [];
  let cur = null;
  let field = null; // "msgctxt" | "msgid" | "msgid_plural" | "msgstr" | "plural:N"

  const push = () => {
    if (cur && (cur.msgid.length || cur.msgstr.length || cur.plurals.length)) {
      entries.push(cur);
    }
  };

  for (const raw of readFileSync(path, "utf8").split("\n")) {
    const line = raw.replace(/\r$/, "");
    if (line === "" || line.startsWith("#")) {
      continue; // blank lines and comments (incl. flags/refs) are not needed
    }
    const kw = line.match(/^(msgctxt|msgid|msgid_plural|msgstr(?:\[(\d+)\])?)\s+"(.*)"$/);
    if (kw) {
      const key = kw[1];
      const val = kw[3];
      if (key === "msgid") {
        push();
        cur = { msgctxt: "", msgid: "", msgstr: "", plurals: [] };
        field = "msgid";
        cur.msgid += val;
      } else if (key === "msgctxt") {
        field = "msgctxt";
        cur.msgctxt += val;
      } else if (key === "msgid_plural") {
        field = "msgid_plural"; // ignored for placeholder checks (uses msgid)
      } else {
        // msgstr or msgstr[N]
        const idx = kw[2];
        if (idx === undefined || idx === "0") {
          field = "msgstr";
          cur.msgstr += val;
        } else {
          field = `plural:${idx}`;
        }
        if (idx !== undefined) {
          cur.plurals[Number(idx)] = (cur.plurals[Number(idx)] || "") + val;
        }
      }
      continue;
    }
    const cont = line.match(/^"(.*)"$/);
    if (cont && cur && field) {
      if (field === "msgid") cur.msgid += cont[1];
      else if (field === "msgctxt") cur.msgctxt += cont[1];
      else if (field === "msgstr") cur.msgstr += cont[1];
      else if (field.startsWith("plural:")) {
        const n = Number(field.slice(7));
        cur.plurals[n] = (cur.plurals[n] || "") + cont[1];
      }
      // msgid_plural continuations are intentionally dropped
    }
  }
  push();
  return entries;
}

// --- Placeholder extraction -------------------------------------------------
// Named `%{name}` (frontend). Literal `%%` is not a placeholder.
function namedPlaceholders(s) {
  return (s.replace(/%%/g, "").match(/%\{[^}]*\}/g) || []).sort();
}
// printf verbs (backend). Matches Go fmt verbs incl. width/precision/index.
function printfVerbs(s) {
  return (s.replace(/%%/g, "").match(/%(?:\d+\$)?[-+ #0]*\d*(?:\.\d+)?[bcdeEfFgGoqstTvxXsdp]/g) || []).sort();
}

function multisetEqual(a, b) {
  if (a.length !== b.length) return false;
  for (let i = 0; i < a.length; i++) if (a[i] !== b[i]) return false;
  return true;
}

// Edge-whitespace / trailing-newline signature.
function edgeSig(s) {
  return {
    lead: /^ /.test(s),
    trail: / $/.test(s),
    nl: s.endsWith("\\n"),
  };
}
function edgeEqual(a, b) {
  const x = edgeSig(a);
  const y = edgeSig(b);
  return x.lead === y.lead && x.trail === y.trail && x.nl === y.nl;
}

// --- msgfmt -c --------------------------------------------------------------
function msgfmtCheck(path) {
  const r = spawnSync("msgfmt", ["-c", "--check-format", "-o", "/dev/null", path], {
    encoding: "utf8",
  });
  if (r.error) return []; // msgfmt not installed → skip silently
  const out = `${r.stdout || ""}${r.stderr || ""}`;
  // Keep only fatal/error lines, drop header-field warnings (Weblate-owned).
  return out
    .split("\n")
    .filter((l) => l.trim() && !/warning: header field/.test(l) && !/warning: /.test(l))
    .map((l) => l.trim());
}

// --- Lint one catalog -------------------------------------------------------
function lintFile(path, side) {
  const findings = [];
  const getPlaceholders = side === "backend" ? printfVerbs : namedPlaceholders;
  for (const e of parsePo(path)) {
    if (!e.msgid) continue; // header
    // Compare the primary msgstr and any plural forms that carry content.
    const targets = [e.msgstr, ...e.plurals.filter(Boolean)];
    for (const str of targets) {
      if (!str) continue; // untranslated → not a defect
      const idP = getPlaceholders(e.msgid);
      const strP = getPlaceholders(str);
      if (!multisetEqual(idP, strP)) {
        findings.push({
          file: path,
          side,
          check: "placeholder",
          msgctxt: e.msgctxt || undefined,
          msgid: e.msgid,
          msgstr: str,
          expected: idP,
          found: strP,
        });
      }
      if (!edgeEqual(e.msgid, str)) {
        const ie = edgeSig(e.msgid);
        // A msgid that itself carries edge whitespace is a *source-string* smell:
        // the stray space usually separates adjacent `{{ }}` interpolations and is
        // rendering-sensitive, so it must be fixed at the source (and verified in a
        // browser), not blindly trimmed from each translation. Reported separately
        // from a genuine translation-only whitespace defect (clean msgid).
        findings.push({
          file: path,
          side,
          check: ie.lead || ie.trail || ie.nl ? "source-whitespace" : "whitespace",
          msgctxt: e.msgctxt || undefined,
          msgid: e.msgid,
          msgstr: str,
          idEdges: ie,
          strEdges: edgeSig(str),
        });
      }
    }
  }
  for (const err of msgfmtCheck(path)) {
    findings.push({ file: path, side, check: "msgfmt", detail: err });
  }
  return findings;
}

// --- Catalog coverage (source → POT) ---------------------------------------
// Blanks comments so commented-out code never counts as a live string. Over-
// stripping only hides a finding, so the crude line/block match is preferred
// over a tokenizer that could mistake a comment for source.
function stripComments(src) {
  return src
    .replace(/<!--[\s\S]*?-->/g, "")
    .replace(/\/\*[\s\S]*?\*\//g, "")
    .split("\n")
    .map((line) => (/^\s*\/\//.test(line) ? "" : line))
    .join("\n");
}

// Converts a JS string literal's body to the escaped form used inside a .po msgid.
function poEscape(body, quote) {
  const unescaped = body
    .replace(/\\n/g, "\n")
    .replace(/\\t/g, "\t")
    .replace(new RegExp(`\\\\\\${quote}`, "g"), quote)
    .replace(/\\\\/g, "\\");
  return unescaped
    .replace(/\\/g, "\\\\")
    .replace(/"/g, '\\"')
    .replace(/\n/g, "\\n")
    .replace(/\t/g, "\\t");
}

const LITERAL = /^\s*(`([^`\\]*)`|"((?:[^"\\]|\\.)*)"|'((?:[^'\\]|\\.)*)')/;

// Collects every literal msgid passed to a gettext call in one source file.
// `$pgettext`/`$npgettext` take the context first, so the msgid is their second
// argument; plural forms are checked via the singular, which the POT also carries.
// Aliased receivers (`view.$gettext(…)`) are matched on purpose: the extractor
// ignores them, which is exactly the defect this check exists to surface.
function sourceMsgids(path) {
  const out = [];
  const src = stripComments(readFileSync(path, "utf8"));
  const calls = /(?:^|[^\w.$])(?:[\w$]+\.)?\$(n?p?)gettext\(([\s\S]{0,400}?)\)/g;
  for (const call of src.matchAll(calls)) {
    let args = call[2];
    if (call[1].includes("p")) {
      const ctx = args.match(LITERAL);
      if (!ctx) continue; // context is not a literal → nothing reliable to check
      args = args.slice(ctx[0].length).replace(/^\s*,/, "");
    }
    const hit = args.match(LITERAL);
    if (!hit) continue; // dynamic argument → intentionally absent from the catalog
    const body = hit[2] ?? hit[3] ?? hit[4];
    if (body === undefined || (hit[1].startsWith("`") && body.includes("${"))) continue;
    out.push({
      msgid: poEscape(body, hit[1][0]),
      line: src.slice(0, call.index).split("\n").length,
    });
  }
  return out;
}

function sourceFiles(dir) {
  const out = [];
  for (const e of readdirSync(dir, { withFileTypes: true })) {
    if (e.isDirectory()) {
      if (["node_modules", "dist", "locales", ".git"].includes(e.name)) continue;
      out.push(...sourceFiles(join(dir, e.name)));
    } else if (/\.(vue|js|ts)$/.test(e.name)) {
      out.push(join(dir, e.name));
    }
  }
  return out;
}

// Flags live `$gettext` literals that never reached the POT.
function lintCoverage() {
  if (!existsSync(POT_PATH)) return [];
  const known = new Set(parsePo(POT_PATH).map((e) => e.msgid));
  const findings = [];
  for (const dir of SOURCE_DIRS.filter((d) => existsSync(d))) {
    for (const file of sourceFiles(dir)) {
      for (const { msgid, line } of sourceMsgids(file)) {
        if (!msgid || known.has(msgid)) {
          continue;
        }
        findings.push({ file, side: "frontend", check: "coverage", msgid, line });
      }
    }
  }
  return findings;
}

// --- Collect catalogs -------------------------------------------------------
function frontendFiles() {
  if (!existsSync(FRONTEND_DIR)) return [];
  return readdirSync(FRONTEND_DIR)
    .filter((f) => f.endsWith(".po"))
    .map((f) => join(FRONTEND_DIR, f))
    .sort();
}
function backendFiles() {
  if (!existsSync(BACKEND_DIR)) return [];
  const out = [];
  for (const d of readdirSync(BACKEND_DIR, { withFileTypes: true })) {
    if (!d.isDirectory()) continue;
    const p = join(BACKEND_DIR, d.name, "default.po");
    if (existsSync(p)) out.push(p);
  }
  return out.sort();
}

// --- Run --------------------------------------------------------------------
const raw = [];
for (const f of frontendFiles()) raw.push(...lintFile(f, "frontend"));
for (const f of backendFiles()) raw.push(...lintFile(f, "backend"));
raw.push(...lintCoverage());

// Collapse source-whitespace findings to one per (side, msgid): the same stray
// space in the source string surfaces once per locale, but it's a single defect.
const all = [];
const seenSource = new Set();
for (const f of raw) {
  if (f.check === "source-whitespace") {
    const key = `${f.side} ${f.msgid}`;
    if (seenSource.has(key)) continue;
    seenSource.add(key);
  }
  all.push(f);
}

for (const finding of all) process.stdout.write(`${JSON.stringify(finding)}\n`);

if (!jsonOnly) {
  const by = (side, check) =>
    all.filter((f) => f.side === side && f.check === check).length;
  const files = (check) => new Set(all.filter((f) => f.check === check).map((f) => f.file)).size;
  const w = (s) => process.stderr.write(`${s}\n`);
  w("");
  w("gettext-lint summary");
  w("--------------------");
  w(`  placeholder       frontend=${by("frontend", "placeholder")}  backend=${by("backend", "placeholder")}  (files=${files("placeholder")})`);
  w(`  whitespace        frontend=${by("frontend", "whitespace")}  backend=${by("backend", "whitespace")}  (translation-side defects, files=${files("whitespace")})`);
  w(`  source-whitespace frontend=${by("frontend", "source-whitespace")}  backend=${by("backend", "source-whitespace")}  (distinct source strings — fix in source, verify rendering)`);
  w(`  msgfmt            frontend=${by("frontend", "msgfmt")}  backend=${by("backend", "msgfmt")}`);
  w(`  coverage          frontend=${by("frontend", "coverage")}  (live $gettext literals missing from the POT)`);
  w(`  TOTAL findings: ${all.length}`);
  w("");
}

process.exit(all.length ? 1 : 0);
