import { describe, expect, it } from "vitest";
import { backendLabels, stackCases, editionRoles, validateBackendCases, validateRoleCases, validateEdition } from "../../playwright/export-contracts.mjs";

// fixtures builds a complete acceptance manifest without relying on live browser results.
function fixtures() {
  return backendLabels.map((label) => ({ label, uid: "ptmrt69ib1crdixs", files: (stackCases[label] || ["jpg"]).map((format, index) => ({ name: `${index}.${format}`, format, root: "/", hash: "a".repeat(40) })) }));
}

// roles supplies every session required by one edition.
function roles(edition) {
  return { edition, sessions: Object.entries(editionRoles[edition]).map(([role, allowed]) => ({ role, export: allowed, download: true, url: "/library/browse", photo: "ptmrt69ib1crdixs", username: role, passwordEnv: "PHOTOPRISM_TEST_PASSWORD" })) };
}

describe("export acceptance requirements", () => {
  it("accepts the complete media matrix", () => expect(validateBackendCases(fixtures())).toHaveLength(backendLabels.length));
  it.each(backendLabels)("fails when mandatory fixture %s is absent", (label) => {
    expect(() => validateBackendCases(fixtures().filter((item) => item.label !== label))).toThrow(`missing mandatory fixture ${label}`);
  });
  it.each(Object.keys(stackCases))("rejects incomplete %s stacks", (label) => {
    const cases = fixtures();
    cases.find((item) => item.label === label).files.pop();
    expect(() => validateBackendCases(cases)).toThrow("incomplete original stack");
  });
  it.each(Object.keys(editionRoles))("requires every real %s session", (edition) => {
    expect(validateRoleCases(roles(edition)).edition).toBe(edition);
    for (const session of roles(edition).sessions) {
      const manifest = roles(edition);
      manifest.sessions = manifest.sessions.filter((item) => item.role !== session.role);
      expect(() => validateRoleCases(manifest)).toThrow(`missing mandatory ${edition} role ${session.role}`);
    }
  });
  it("rejects public mode and inactive paid editions before role setup", () => {
    expect(() => validateEdition("ce", { edition: "ce", public: false })).not.toThrow();
    expect(() => validateEdition("ce", { edition: "ce", public: true })).toThrow("private password authentication");
    expect(() => validateEdition("plus", { edition: "plus", sponsor: false, public: false })).toThrow("registered feature set");
    expect(() => validateEdition("pro", { edition: "unregistered", sponsor: false, public: false })).toThrow("registered edition");
    expect(() => validateEdition("pro", { edition: "business", sponsor: true, public: false })).not.toThrow();
  });
  it("rejects substituted roles, wrong permissions, and absent Download expectations", () => {
    const manifest = roles("pro");
    manifest.sessions.find((item) => item.role === "contributor").role = "uploader";
    expect(() => validateRoleCases(manifest)).toThrow("missing mandatory pro role contributor");
    const viewer = roles("plus");
    viewer.sessions.find((item) => item.role === "viewer").export = true;
    expect(() => validateRoleCases(viewer)).toThrow("unexpected export permission");
    const missing = roles("ce");
    delete missing.sessions[0].download;
    expect(() => validateRoleCases(missing)).toThrow("missing legacy Download expectation");
  });
});
