// Runs the callback script the auth template renders against two populated in-memory stores and
// prints what they hold afterwards, so a Go test can assert on the result. Reads the script path
// and a JSON object of preset entries per store from argv, and writes one JSON object to stdout.
const fs = require("node:fs");
const vm = require("node:vm");

const script = fs.readFileSync(process.argv[2], "utf8");
const preset = JSON.parse(fs.readFileSync(process.argv[3], "utf8"));

// Storage implements the subset of the Web Storage API the callback script uses.
class Storage {
  constructor(entries) {
    this.entries = new Map(Object.entries(entries));
  }

  getItem(key) {
    return this.entries.has(key) ? this.entries.get(key) : null;
  }

  setItem(key, value) {
    this.entries.set(key, String(value));
  }

  removeItem(key) {
    this.entries.delete(key);
  }
}

const localStorage = new Storage(preset.localStorage || {});
const sessionStorage = new Storage(preset.sessionStorage || {});
const window = { localStorage, sessionStorage, location: { href: "" } };

vm.runInNewContext(script, { window, localStorage, sessionStorage, console });

process.stdout.write(
  JSON.stringify({
    localStorage: Object.fromEntries(localStorage.entries),
    sessionStorage: Object.fromEntries(sessionStorage.entries),
    location: window.location.href,
  })
);
