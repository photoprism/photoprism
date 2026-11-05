import { describe, it, expect } from "vitest";
import { photoTypeOptions } from "common/photo-types";
import { PhotoTypes } from "options/options";

describe("page media types", () => {
  it("reuses all shared types on unrestricted pages", () => {
    expect(photoTypeOptions()).toEqual(PhotoTypes());
    expect(photoTypeOptions({}, "cats").map((item) => item.value)).toContain("document");
    expect(photoTypeOptions({}, 'title:"vectors"')).toEqual(PhotoTypes());
    expect(photoTypeOptions({}, '"vectors"')).toEqual(PhotoTypes());
  });
  it.each(["", "stacks", "scans"])("limits Photos with query %s", (q) => {
    expect(photoTypeOptions({ photo: "true" }, q)).toEqual(PhotoTypes().filter((item) => ["image", "raw", "live"].includes(item.value)));
  });
  it.each(["mono:true quality:3 photo:true"])("limits Monochrome query %s", (q) => {
    expect(photoTypeOptions({}, q).map((item) => item.value)).toEqual(["image", "raw", "live"]);
  });
  it("limits Media and recognizes affirmative filters", () => {
    expect(photoTypeOptions({ media: true }).map((item) => item.value)).toEqual(["live", "video", "audio", "animated"]);
    expect(photoTypeOptions({}, "media:true")).toEqual(PhotoTypes());
    expect(photoTypeOptions({ video: "false", photo: "false" })).toEqual(PhotoTypes());
  });
  it.each(["video", "live", "animated", "audio", "vector", "document", "raw"])("hides fixed %s pages", (key) => {
    expect(photoTypeOptions({ [key]: "true" })).toEqual([]);
  });
  it("hides fixed type pages", () => {
    expect(photoTypeOptions({ type: "vector" })).toEqual([]);
  });
  it.each(["vectors", "documents"])("hides fixed query %s", (q) => {
    expect(photoTypeOptions({}, q)).toEqual([]);
  });
  it.each(["live", "raw", "animated", "video", "audio", "monochrome", "mono", "photo:true", "media:true", "vector:true", "document:true", "type:vector", "cats documents"])("preserves options for free search %s", (q) => {
    expect(photoTypeOptions({}, q)).toEqual(PhotoTypes());
    expect(photoTypeOptions({ photo: "true" }, q).map((item) => item.value)).toEqual(["image", "raw", "live"]);
    expect(photoTypeOptions({ media: "true" }, q).map((item) => item.value)).toEqual(["live", "video", "audio", "animated"]);
  });
});
