import { describe, it, expect } from "vitest";
import $util from "common/util";
import PPeopleMergeDialog from "component/people/merge/dialog.vue";

// interpolate replaces %{name} placeholders like $gettext.
const interpolate = (s, params) => s.replace(/%{(\w+)}/g, (_, k) => params[k]);

describe("component/people/merge/dialog", () => {
  it("isolates both names in the prompt", () => {
    const prompt = PPeopleMergeDialog.computed.prompt.call({
      subj1: { originalValue: () => "سارة\u2069\u200F" },
      subj2: { Name: "Max" },
      $gettext: interpolate,
      $util,
    });
    expect(prompt).toBe("Merge \u2068سارة\u200F\u2069 with \u2068Max\u2069?");
  });
  it("returns an empty prompt without both people", () => {
    expect(PPeopleMergeDialog.computed.prompt.call({ subj1: null, subj2: { Name: "Max" }, $gettext: interpolate, $util })).toBe("");
  });
});
