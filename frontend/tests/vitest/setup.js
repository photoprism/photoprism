import { afterEach, vi } from "vitest";
import { Settings } from "luxon";
import "@testing-library/jest-dom";

// Set config for Luxon
Settings.defaultLocale = "en";
Settings.defaultZoneName = "UTC";

// Import and set up global config
import clientConfig from "./config";
import { $config } from "app/session";

$config.setValues(clientConfig);

// Make config available in browser environment
window.__CONFIG__ = clientConfig;

console.log("Running tests in real browser environment");

// Clean up after each test
afterEach(() => {
  vi.resetAllMocks();
});

// Export shared configuration
export { clientConfig };
