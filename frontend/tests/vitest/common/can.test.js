import "../fixtures";
import * as can from "common/can";
import { expect, describe, it } from "vitest";

// These tests are not suitable for running on jsdom

describe.skip("common/can", () => {
  it("useVideo", () => {
    expect(can.useVideo).toBe(true);
  });

  it("useMp4Avc", () => {
    expect(can.useMp4Avc).toBe(true);
  });

  it("useMp4Hvc", () => {
    expect(can.useMp4Hvc).toBe(false);
  });

  it("useMp4Hev", () => {
    expect(can.useMp4Hev).toBe(false);
  });

  it("useMp4Vvc", () => {
    expect(can.useMp4Vvc).toBe(false);
  });

  it("useMp4Evc", () => {
    expect(can.useMp4Evc).toBe(false);
  });

  it("useMp4Av1", () => {
    expect(can.useMp4Av1).toBe(true);
  });

  it("useVP8", () => {
    expect(can.useVP8).toBe(true);
  });

  it("useVP9", () => {
    expect(can.useVP9).toBe(true);
  });

  it("useWebmAv1", () => {
    expect(can.useWebmAv1).toBe(true);
  });

  it("useMkvAv1", () => {
    expect(can.useMkvAv1).toBe(false);
  });

  it("useWebM", () => {
    expect(can.useWebM).toBe(true);
  });

  it("useTheora", () => {
    expect(can.useTheora).toBe(true);
  });
});
