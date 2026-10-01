import { describe, it, expect } from "vitest";
import { en } from "vuetify/locale";
import { Messages } from "../../src/locales";

// keyPaths returns the dot-separated paths of all leaf values in a message tree.
function keyPaths(tree, prefix = "") {
  return Object.entries(tree).flatMap(([key, value]) =>
    value && typeof value === "object" ? keyPaths(value, `${prefix}${key}.`) : [`${prefix}${key}`]
  );
}

describe("locales Messages", () => {
  it("defines every message key of the Vuetify English locale", () => {
    expect(keyPaths(Messages((msgid) => msgid)).sort()).toEqual(keyPaths(en).sort());
  });
  it("uses the Vuetify English wording as msgid, including its placeholders", () => {
    expect(Messages((msgid) => msgid)).toEqual(en);
  });
  it("passes every message except the bare date placeholder through gettext", () => {
    const seen = [];
    const messages = Messages((msgid) => {
      seen.push(msgid);
      return `T:${msgid}`;
    });
    expect(messages.datePicker.ariaLabel.selectDate).toBe("{0}");
    expect(seen.length).toBe(keyPaths(en).length - 1);
  });
});
