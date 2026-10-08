import { describe, it, expect, beforeAll, afterAll } from "vitest";
import { spawnSync } from "node:child_process";
import fs from "node:fs";
import os from "node:os";
import path from "node:path";
import { defaultDirs, findUnreviewedSinks, findUnreviewedVHtml, listFiles, scan } from "../../../scripts/scan-xss.mjs";

const script = path.resolve(import.meta.dirname, "../../../scripts/scan-xss.mjs");
const note = "<!-- eslint-disable-next-line vue/no-v-html -- value is sanitized -->";

describe("findUnreviewedVHtml", () => {
  it("accepts a binding directly below a reviewed comment", () => {
    expect(findUnreviewedVHtml(`${note}\n<div v-html="html"></div>`)).toEqual([]);
  });
  it("accepts a reviewed comment in a script", () => {
    expect(findUnreviewedVHtml(`// eslint-disable-next-line vue/no-v-html -- sanitized\ntemplate: '<div v-html="html"></div>'`)).toEqual([]);
  });
  it("accepts CRLF line endings", () => {
    expect(findUnreviewedVHtml(`${note}\r\n<div v-html="html"></div>\r\n`)).toEqual([]);
  });
  it("reports a binding without a comment", () => {
    expect(findUnreviewedVHtml(`<p>text</p>\n<div v-html="html"></div>`)).toEqual([2]);
  });
  it("reports a binding on the first line", () => {
    expect(findUnreviewedVHtml(`<div v-html="html"></div>`)).toEqual([1]);
  });
  it("reports a binding separated from the comment by a blank line", () => {
    expect(findUnreviewedVHtml(`${note}\n\n<div v-html="html"></div>`)).toEqual([3]);
  });
  it("reports a comment without a reason", () => {
    expect(findUnreviewedVHtml(`<!-- eslint-disable-next-line vue/no-v-html -->\n<div v-html="html"></div>`)).toEqual([2]);
    expect(findUnreviewedVHtml(`<!-- eslint-disable-next-line vue/no-v-html -- -->\n<div v-html="html"></div>`)).toEqual([2]);
    expect(findUnreviewedVHtml(`/* eslint-disable-next-line vue/no-v-html -- */\n<div v-html="html"></div>`)).toEqual([2]);
  });
  it("reports a note outside a comment", () => {
    expect(findUnreviewedVHtml(`<p title="eslint-disable-next-line vue/no-v-html -- x"></p>\n<div v-html="html"></div>`)).toEqual([2]);
  });
  it("reports a comment for another rule", () => {
    expect(findUnreviewedVHtml(`<!-- eslint-disable-next-line vue/no-unused-vars -- x -->\n<div v-html="html"></div>`)).toEqual([2]);
  });
  it("covers only the next line", () => {
    expect(findUnreviewedVHtml(`${note}\n<div v-html="a"></div>\n<div v-html="b"></div>`)).toEqual([3]);
  });
  it("reports binding variants", () => {
    const lines = [`<div v-html='html'></div>`, `<div v-html = "html"></div>`, `<div :innerHTML="html"></div>`, `<div v-bind:innerHTML="html"></div>`, `<div :innerHTML.prop="html"></div>`, `<div :outerHTML="html"></div>`];
    expect(findUnreviewedVHtml(lines.join("\n"))).toEqual([1, 2, 3, 4, 5, 6]);
  });
  it("reports srcdoc bindings", () => {
    const lines = [`<iframe :srcdoc="html"></iframe>`, `<iframe v-bind:srcdoc="html"></iframe>`, `<iframe :srcdoc.prop="html"></iframe>`, `<iframe v-bind:srcdoc.prop="html"></iframe>`, `<iframe :srcDoc="html"></iframe>`];
    expect(findUnreviewedVHtml(lines.join("\n"))).toEqual([1, 2, 3, 4, 5]);
  });
  it("ignores mentions without a binding", () => {
    expect(findUnreviewedVHtml(`// v-html fields like caption and notes`)).toEqual([]);
  });
});

describe("findUnreviewedSinks", () => {
  it("reports HTML assignments and writes", () => {
    const lines = ["el.innerHTML = html;", "el.outerHTML = html;", "el.insertAdjacentHTML('beforeend', html);", "document.write(html);", "document.writeln(html);"];
    expect(findUnreviewedSinks(lines.join("\n"))).toEqual([1, 2, 3, 4, 5]);
  });
  it("reports assignment variants", () => {
    const lines = ["el.innerHTML=html;", "el.innerHTML += html;", 'el["innerHTML"] = html;', "el.insertAdjacentHTML ('beforeend', html);", 'el.innerHTML = "" + html;', "el.innerHTML = \"'\" + html;", 'a.innerHTML = html; b.innerHTML = "";', "document.write (html);", "el.innerHTML ||= html;", "el.innerHTML ??= html;", "el.outerHTML &&= html;"];
    expect(findUnreviewedSinks(lines.join("\n"))).toEqual([1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11]);
  });
  it("accepts reviewed and clearing assignments", () => {
    const lines = ["el.innerHTML = html; // security-reviewed: sanitized", 'el.innerHTML = "";', "el.innerHTML = ''", 'this.$refs.map.innerHTML = "";'];
    expect(findUnreviewedSinks(lines.join("\n"))).toEqual([]);
  });
  it("ignores comparisons", () => {
    expect(findUnreviewedSinks("if (el.innerHTML === html) {}\nif (el.outerHTML == html) {}")).toEqual([]);
  });
  it("ignores text assignments", () => {
    expect(findUnreviewedSinks("el.textContent = text;")).toEqual([]);
  });
  it("reports HTML properties in objects", () => {
    const lines = ['<div v-bind="{ innerHTML: html }"></div>', 'h("div", { innerHTML: html })', "Object.assign(el, {outerHTML: html})", 'const props = { "innerHTML": html };', "  innerHTML: html,"];
    expect(findUnreviewedSinks(lines.join("\n"))).toEqual([1, 2, 3, 4, 5]);
  });
  it("reports srcdoc assignments and parse calls", () => {
    const lines = ["frame.srcdoc = html;", "range.createContextualFragment(html);", "el.setHTMLUnsafe(html);", "Document.parseHTMLUnsafe(html);", 'new DOMParser().parseFromString(html, "text/html");', "frame.srcdoc += html;", "frame.srcdoc ||= html;", 'frame["srcdoc"] = html;', 'frame.setAttribute("srcdoc", html);', "frame.setAttribute('srcDoc', html);", "frame.srcDoc = html;"];
    expect(findUnreviewedSinks(lines.join("\n"))).toEqual([1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11]);
  });
  it("reports srcdoc properties and shorthand properties in objects", () => {
    const lines = ['h("iframe", { srcdoc: html })', '<iframe v-bind="{ srcdoc: html }"></iframe>', "h('div', { innerHTML })", "const props = { id, innerHTML };"];
    expect(findUnreviewedSinks(lines.join("\n"))).toEqual([1, 2, 3, 4]);
  });
  it("ignores similar names and text properties", () => {
    const lines = ["const innerHTMLLength = 3;", "{ textContent: text }", "if (frame.srcdoc === html) {}", "parseFromStringValue = 1;", "const srcdocs = [];", 'frame.setAttribute("src", url);'];
    expect(findUnreviewedSinks(lines.join("\n"))).toEqual([]);
  });
});

describe("listFiles, scan, and the command", () => {
  let dir;

  beforeAll(() => {
    dir = fs.mkdtempSync(path.join(os.tmpdir(), "scan-xss-"));
    fs.mkdirSync(path.join(dir, "component"));
    fs.mkdirSync(path.join(dir, "locales"));
    fs.writeFileSync(path.join(dir, "component", "ok.vue"), `${note}\n<div v-html="html"></div>\n`);
    fs.writeFileSync(path.join(dir, "component", "bad.vue"), `<div v-html="html"></div>\n`);
    fs.writeFileSync(path.join(dir, ".hidden.js"), "el.outerHTML = html;\n");
    fs.writeFileSync(path.join(dir, "util.js"), "el.innerHTML = html;\n");
    fs.writeFileSync(path.join(dir, "locales", "de.json"), '{"text": "document.write(x)"}\n');
  });

  afterAll(() => {
    fs.rmSync(dir, { recursive: true, force: true });
  });

  it("lists code files recursively in path order", () => {
    expect(listFiles(dir).map((f) => path.relative(dir, f))).toEqual([".hidden.js", path.join("component", "bad.vue"), path.join("component", "ok.vue"), "util.js"]);
  });
  it("reports each unreviewed finding with its file and line", () => {
    expect(scan([dir])).toEqual([
      `${path.join(dir, ".hidden.js")}:1: DOM HTML sink without a security-reviewed note`,
      `${path.join(dir, "component", "bad.vue")}:1: HTML binding without a reviewed eslint-disable comment`,
      `${path.join(dir, "util.js")}:1: DOM HTML sink without a security-reviewed note`,
    ]);
  });
  it("exits with status 1 and lists findings", () => {
    const result = spawnSync(process.execPath, [script, dir], { encoding: "utf8" });
    expect(result.status).toBe(1);
    expect(result.stdout).toContain(`${path.join(dir, "util.js")}:1:`);
  });
  it("exits with status 0 when nothing is found", () => {
    const result = spawnSync(process.execPath, [script, path.join(dir, "locales")], { encoding: "utf8" });
    expect(result.status).toBe(0);
    expect(result.stdout).toContain("OK:");
  });
  it("skips test and dependency directories", () => {
    const root = fs.mkdtempSync(path.join(os.tmpdir(), "scan-xss-skip-"));

    try {
      for (const sub of ["tests", "node_modules", "page"]) {
        fs.mkdirSync(path.join(root, sub));
        fs.writeFileSync(path.join(root, sub, "x.js"), "el.innerHTML = html;\n");
      }

      expect(listFiles(root).map((f) => path.relative(root, f))).toEqual([path.join("page", "x.js")]);
    } finally {
      fs.rmSync(root, { recursive: true, force: true });
    }
  });
  it("includes the edition overlays that exist", () => {
    const root = fs.mkdtempSync(path.join(os.tmpdir(), "scan-xss-overlays-"));

    try {
      fs.mkdirSync(path.join(root, "frontend", "src"), { recursive: true });
      fs.mkdirSync(path.join(root, "pro", "frontend"), { recursive: true });

      expect(defaultDirs(path.join(root, "frontend"))).toEqual([path.join(root, "frontend", "src"), path.join(root, "pro", "frontend")]);
    } finally {
      fs.rmSync(root, { recursive: true, force: true });
    }
  });
  it("finds nothing in the application sources and edition overlays", () => {
    expect(scan(defaultDirs(path.resolve(import.meta.dirname, "../../..")))).toEqual([]);
  });
});
