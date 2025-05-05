import "@testing-library/jest-dom";
import { cleanup } from "@testing-library/react";
import { afterEach, vi, beforeAll } from "vitest";
import { setupCommonMocks } from "./fixtures";

global.window = global.window || {};
global.window.__CONFIG__ = {
  debug: false,
  trace: false,
};

global.window.location = {
  protocol: "https:",
};

global.navigator = {
  userAgent: "node.js",
  maxTouchPoints: 0,
};

afterEach(() => {
  cleanup();
});

beforeAll(() => {
  setupCommonMocks();
});

vi.mock("luxon", () => ({
  DateTime: {
    fromISO: vi.fn().mockReturnValue({
      toLocaleString: vi.fn().mockReturnValue("2023-10-01 10:00:00"),
    }),
    DATETIME_MED: {},
    DATETIME_MED_WITH_WEEKDAY: {},
    DATE_MED: {},
    TIME_24_SIMPLE: {},
  },
  Settings: {
    defaultLocale: "en",
    defaultZoneName: "UTC",
  },
}));

vi.mock("common/gettext", () => ({
  $gettext: vi.fn((text) => text),
}));

vi.mock("app/session", () => ({
  $config: {},
}));

vi.mock("common/notify", () => ({
  default: {
    success: vi.fn(),
    warn: vi.fn(),
    error: vi.fn(),
  },
}));
