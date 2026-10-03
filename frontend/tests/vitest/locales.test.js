import { describe, it, expect } from "vitest";
import { en } from "vuetify/locale";
import { Messages } from "../../src/locales";

// Vuetify messages left in English because the UI does not render the components that use them.
const untranslatedPaths = [
  "datePicker.ariaLabel.selectDate",
  "monthPicker",
  "timePicker.hour",
  "timePicker.minute",
  "timePicker.second",
  "timePicker.notAllowed",
  "heatmap",
  "rules",
  "command",
  "hotkey",
  "video",
  "colorPicker",
];

// keyPaths returns the dot-separated paths of all leaf values in a message tree.
function keyPaths(tree, prefix = "") {
  return Object.entries(tree).flatMap(([key, value]) =>
    value && typeof value === "object" ? keyPaths(value, `${prefix}${key}.`) : [`${prefix}${key}`]
  );
}

// valueAt returns the value at a dot-separated path.
function valueAt(tree, path) {
  return path.split(".").reduce((node, key) => node?.[key], tree);
}

describe("locales Messages", () => {
  it("defines every message key of the Vuetify English locale", () => {
    expect(keyPaths(Messages((msgid) => msgid)).sort()).toEqual(keyPaths(en).sort());
  });
  it("uses the Vuetify English wording as msgid, including its placeholders", () => {
    expect(Messages((msgid) => msgid)).toEqual(en);
  });
  it("translates every message except those of components the UI does not use", () => {
    const messages = Messages((msgid) => `T:${msgid}`);
    const untranslated = keyPaths(en).filter((path) => !valueAt(messages, path).startsWith("T:"));
    const unexpected = untranslated.filter((path) => !untranslatedPaths.some((p) => path === p || path.startsWith(`${p}.`)));
    expect(unexpected).toEqual([]);
    expect(valueAt(messages, "datePicker.ariaLabel.nextMonth")).toBe("T:Next month");
    expect(valueAt(messages, "hotkey.ctrl")).toBe(en.hotkey.ctrl);
  });
});
