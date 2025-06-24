import { describe, it, expect } from 'vitest';
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
});