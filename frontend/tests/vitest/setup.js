import { afterEach, vi } from "vitest";
import "@testing-library/jest-dom";
import "./vue-setup";

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
