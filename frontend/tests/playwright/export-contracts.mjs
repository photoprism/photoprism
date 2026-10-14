import assert from "node:assert/strict";

export const stackCases = {
  TwoImage: ["jpg", "jpg"],
  TwoRAW: ["dng", "dng"],
  TwoSVG: ["svg", "svg"],
  TwoPDF: ["pdf", "pdf"],
  TwoGIF: ["gif", "gif"],
  TwoAPNG: ["png", "png"],
  TwoWebP: ["webp", "webp"],
  TwoVideo: ["mov", "mov"],
  TwoMotion: ["jpg", "jpg"],
  TwoApple: ["jpg", "mov", "jpg", "mov"],
  RAWPlusJPEG: ["dng", "jpg"],
};

export const backendLabels = ["HEVC", "AVC", "AppleJPEG", "AppleHEIC", "Motion", "Animationgif", "Animationpng", "Animationwebp", "Mixed", ...Object.keys(stackCases)];
export const editionRoles = {
  ce: { admin: true, guest: false, "share-link": false },
  plus: { admin: true, user: true, viewer: false, guest: false, "share-link": false },
  pro: { admin: true, manager: true, user: true, viewer: false, contributor: false, guest: false, "share-link": false },
};

// validateBackendCases requires every media scenario and its independent original-file manifest.
export function validateBackendCases(cases) {
  assert.ok(Array.isArray(cases), "backend cases must be an array");
  assert.equal(new Set(cases.map((item) => item.label)).size, cases.length, "duplicate backend labels");
  for (const label of backendLabels) {
    const fixture = cases.find((item) => item.label === label);
    assert.ok(fixture, `missing mandatory fixture ${label}`);
    assert.match(fixture.uid, /^p[a-z0-9]{15}$/, `${label}: invalid photo UID`);
    assert.ok(Array.isArray(fixture.files) && fixture.files.length, `${label}: missing file manifest`);
    for (const file of fixture.files) {
      assert.match(file.hash, /^[a-f0-9]{40}$/, `${label}: invalid source hash`);
      assert.ok(file.name && ["/", "sidecar"].includes(file.root) && file.format, `${label}: invalid source record`);
    }
    if (stackCases[label]) {
      assert.deepEqual(fixture.files.filter((file) => file.root === "/").map((file) => file.format).sort(), [...stackCases[label]].sort(), `${label}: incomplete original stack`);
    }
  }
  return cases;
}

// validateRoleCases refuses incomplete edition matrices or substitute roles.
export function validateRoleCases(manifest) {
  assert.ok(editionRoles[manifest.edition], "edition must be ce, plus, or pro");
  assert.ok(Array.isArray(manifest.sessions), "missing role sessions");
  assert.equal(new Set(manifest.sessions.map((item) => item.role)).size, manifest.sessions.length, "duplicate role sessions");
  for (const [role, allowed] of Object.entries(editionRoles[manifest.edition])) {
    const session = manifest.sessions.find((item) => item.role === role);
    assert.ok(session, `missing mandatory ${manifest.edition} role ${role}`);
    assert.equal(session.export, allowed, `${role}: unexpected export permission`);
    assert.equal(typeof session.download, "boolean", `${role}: missing legacy Download expectation`);
    assert.ok(session.url && session.photo, `${role}: missing route or photo`);
    if (role !== "share-link") {
      assert.ok(session.username && session.passwordEnv, `${role}: missing login inputs`);
    }
  }
  assert.equal(manifest.sessions.length, Object.keys(editionRoles[manifest.edition]).length, "unexpected edition roles");
  return manifest;
}

// validateEdition rejects public mode and inactive paid-edition fallbacks before fixture writes.
export function validateEdition(edition, state) {
  assert.equal(state.public, false, "role acceptance requires private password authentication");
  const names = { ce: ["ce"], plus: ["plus"], pro: ["evaluation", "starter", "nonprofit", "business", "enterprise", "ultimate"] };
  assert.ok(names[edition]?.includes(state.edition), `${edition}: required registered edition is unavailable (${state.edition})`);
  if (edition !== "ce") {
    assert.equal(state.sponsor, true, `${edition}: required registered feature set is unavailable`);
  }
}
