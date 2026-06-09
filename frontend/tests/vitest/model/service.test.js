import { describe, it, expect } from "vitest";
import "../fixtures";

import Service from "model/service";
import Photo from "model/photo";

describe("model/service", () => {
  it("should get service defaults", () => {
    const values = { ID: 5 };
    const service = new Service(values);
    const result = service.getDefaults();
    expect(result.ID).toBe(0);
    expect(result.AccShare).toBe(true);
    expect(result.AccName).toBe("");
  });

  it("should get service entity name", () => {
    const values = { ID: 5, AccName: "Test Name" };
    const service = new Service(values);
    const result = service.getEntityName();
    expect(result).toBe("Test Name");
  });

  it("should get service id", () => {
    const values = { ID: 5, AccName: "Test Name" };
    const service = new Service(values);
    const result = service.getId();
    expect(result).toBe(5);
  });

  it("should get folders", async () => {
    const values = { ID: 123, AccName: "Test Name" };
    const service = new Service(values);
    const response = await service.Folders();
    expect(response.foo).toBe("folders");
  });

  it("should get share photos", async () => {
    const values = { ID: 123, AccName: "Test Name" };
    const service = new Service(values);
    const values1 = { ID: 5, Title: "Crazy Cat", UID: 789 };
    const photo = new Photo(values1);
    const values2 = { ID: 6, Title: "Crazy Cat 2", UID: 783 };
    const photo2 = new Photo(values2);
    const Photos = [photo, photo2];
    const response = await service.Upload(Photos, "destination");
    expect(response.foo).toBe("upload");
  });

  it("should get collection resource", () => {
    const result = Service.getCollectionResource();
    expect(result).toBe("services");
  });

  it("should get model name", () => {
    const result = Service.getModelName();
    expect(result).toBe("Account");
  });

  describe("write-only credentials", () => {
    it("tracks AccPass/AccKey even when the API response omits them", () => {
      // The backend tags AccPass/AccKey with `json:"-"`, so responses never
      // carry them. The Service constructor must still seed __originalValues
      // so a user edit in the dialog reaches getValues(true).
      const service = new Service({ ID: 1, AccName: "Nextcloud", AccURL: "https://nc.example/" });
      expect(service.AccPass).toBe("");
      expect(service.AccKey).toBe("");
      expect(service.__originalValues.AccPass).toBe("");
      expect(service.__originalValues.AccKey).toBe("");
    });

    it("includes AccPass/AccKey in the change diff after a dialog edit", () => {
      const service = new Service({ ID: 1, AccName: "Nextcloud", AccURL: "https://nc.example/" });
      service.AccPass = "new-secret";
      service.AccKey = "new-key";
      const diff = service.getValues(true);
      expect(diff.AccPass).toBe("new-secret");
      expect(diff.AccKey).toBe("new-key");
    });

    it("does not include AccPass/AccKey when the user leaves them blank", () => {
      const service = new Service({ ID: 1, AccName: "Nextcloud", AccURL: "https://nc.example/" });
      service.AccName = "Renamed";
      const diff = service.getValues(true);
      expect(diff.AccName).toBe("Renamed");
      expect(diff).not.toHaveProperty("AccPass");
      expect(diff).not.toHaveProperty("AccKey");
    });

    it("preserves credentials when the constructor receives them via clone()", () => {
      // clone() routes through getValues() → new Service(); AccPass that was
      // typed into the original must survive the round-trip on __originalValues
      // so subsequent edits keep flowing through the diff.
      const original = new Service({ ID: 1, AccName: "Nextcloud" });
      original.AccPass = "typed";
      const copy = new Service(original.getValues());
      expect(copy.AccPass).toBe("typed");
      expect(copy.__originalValues.AccPass).toBe("typed");
    });
  });
});
