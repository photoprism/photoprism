import { describe, it, expect, beforeEach } from "vitest";

import * as themes from "../../../src/options/themes";

describe("options/themes", () => {
  beforeEach(() => {
    themes.SetOptions([
      {
        text: "Default",
        value: "default",
        disabled: false,
      },
    ]);

    themes.Set("default", {
      name: "default",
      title: "Default",
      colors: {},
      variables: {},
    });
  });

  it("assigns multiple themes and enforces forced theme", () => {
    const forcedTheme = {
      name: "forced",
      title: "Forced",
      force: true,
      colors: { background: "#000000" },
      variables: {},
    };

    const optionalTheme = {
      name: "optional",
      title: "Optional",
      colors: { background: "#ffffff" },
      variables: {},
    };

    themes.Assign([optionalTheme, forcedTheme]);

    const available = themes.Options();
    expect(available).toHaveLength(1);
    expect(available[0].value).toBe("forced");

    const forced = themes.Get("default", true);
    expect(forced.name).toBe("forced");
    expect(forced.colors.background).toBe("#000000");
  });

  it("does not add duplicate entries when Assign is called multiple times", () => {
    const theme = {
      name: "example",
      title: "Example",
      colors: { background: "#123456" },
      variables: {},
    };

    themes.Assign([theme]);
    themes.Assign([theme]);

    const available = themes.Options();
    expect(available.findIndex((option) => option.value === "example")).toBeGreaterThan(-1);
    const filtered = available.filter((option) => option?.value === "example");
    expect(filtered).toHaveLength(1);
  });
});
