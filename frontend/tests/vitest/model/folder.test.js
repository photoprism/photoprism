import { describe, it, expect } from "vitest";
import "../fixtures";
import Folder from "model/folder";

describe("model/folder", () => {
  it("should return classes", () => {
    const values = {
      Folder: true,
      Path: "2011/10-Halloween",
      UID: "dqbevau2zlhxrxww",
      Title: "Halloween Party",
      Favorite: true,
      Private: true,
      Ignore: false,
      Watch: false,
      FileCount: 0,
    };
    const folder = new Folder(values);
    const result = folder.classes(true);
    expect(result).toContain("is-folder");
    expect(result).toContain("uid-dqbevau2zlhxrxww");
    expect(result).toContain("is-favorite");
    expect(result).toContain("is-private");
    expect(result).toContain("is-selected");
  });

  it("should get folder defaults", () => {
    const values = {
      Folder: true,
      Path: "2011/10-Halloween",
      Root: "",
      UID: "dqbevau2zlhxrxww",
      Type: "",
      Title: "Halloween Party",
      Category: "",
      Description: "",
      Order: "",
      Country: "",
      Year: "",
      Month: "",
      Favorite: false,
      Private: false,
      Ignore: false,
      Watch: false,
      FileCount: 0,
      CreatedAt: "",
      UpdatedAt: "",
    };
    const model = new Folder(values);
    const result = model.getDefaults();
    expect(result.Folder).toBe(true);
    expect(result.Path).toBe("");
    expect(result.Favorite).toBe(false);
  });

  it("should get folder base name", () => {
    const values = {
      Folder: true,
      Path: "2011/10-Halloween",
      Root: "",
      UID: "dqbevau2zlhxrxww",
      Type: "",
      Title: "Halloween Party",
      Category: "",
      Description: "",
      Order: "",
      Country: "",
      Year: "",
      Month: "",
      Favorite: false,
      Private: false,
      Ignore: false,
      Watch: false,
      FileCount: 0,
      CreatedAt: "",
      UpdatedAt: "",
    };
    const folder = new Folder(values);
    const result = folder.baseName();
    expect(result).toBe("10-Halloween");
    const result2 = folder.baseName(8);
    expect(result2).toBe("10-Hall…");
  });

  it("should return false", () => {
    const values = {
      Folder: true,
      Path: "2011/10-Halloween",
      UID: "dqbevau2zlhxrxww",
      Title: "Halloween Party",
    };
    const folder = new Folder(values);
    expect(folder.isFile()).toBe(false);
  });

  it("should return entity name", () => {
    const values = {
      Folder: true,
      Path: "2011/10-Halloween",
      Root: "",
      UID: "dqbevau2zlhxrxww",
      Title: "Halloween Party",
    };
    const folder = new Folder(values);
    expect(folder.getEntityName()).toBe("/2011/10-Halloween");
  });

  it("should return thumbnail url", () => {
    const values = {
      Folder: true,
      Path: "2011/10-Halloween",
      Root: "",
      UID: "dqbevau2zlhxrxww",
      Title: "Halloween Party",
    };
    const folder = new Folder(values);
    expect(folder.thumbnailUrl("tile_224")).toBe("/api/v1/folders/t/dqbevau2zlhxrxww/public/tile_224");
  });

  it("should get date string", () => {
    const values = {
      Folder: true,
      Path: "2011/10-Halloween",
      Root: "",
      UID: "dqbevau2zlhxrxww",
      Title: "Halloween Party",
      CreatedAt: "2012-07-08T14:45:39Z",
      UpdatedAt: "2012-07-08T14:45:39Z",
    };
    const folder = new Folder(values);
    expect(folder.getDateString().replaceAll("\u202f", " ")).toBe("Jul 8, 2012, 2:45 PM");
  });

  it("should toggle like", () => {
    const values = {
      Folder: true,
      Path: "2011/10-Halloween",
      Root: "",
      UID: "dqbevau2zlhxrxww",
      Title: "Halloween Party",
      Favorite: true,
      Private: true,
    };
    const folder = new Folder(values);
    expect(folder.Favorite).toBe(true);
    folder.toggleLike();
    expect(folder.Favorite).toBe(false);
    folder.toggleLike();
    expect(folder.Favorite).toBe(true);
  });

  it("should like folder", () => {
    const values = {
      Folder: true,
      Path: "2011/10-Halloween",
      Root: "",
      UID: "dqbevau2zlhxrxww",
      Title: "Halloween Party",
      Favorite: false,
      Private: true,
    };
    const folder = new Folder(values);
    expect(folder.Favorite).toBe(false);
    folder.like();
    expect(folder.Favorite).toBe(true);
  });

  it("should unlike folder", () => {
    const values = {
      Folder: true,
      Path: "2011/10-Halloween",
      Root: "",
      UID: "dqbevau2zlhxrxww",
      Title: "Halloween Party",
      Favorite: true,
      Private: true,
    };
    const folder = new Folder(values);
    expect(folder.Favorite).toBe(true);
    folder.unlike();
    expect(folder.Favorite).toBe(false);
  });

  it("should get collection resource", () => {
    const result = Folder.getCollectionResource();
    expect(result).toBe("folders");
  });

  it("should get model name", () => {
    const result = Folder.getModelName();
    expect(result).toBe("Folder");
  });

  it("should test find all", async () => {
    try {
      const response = await Folder.findAll("2011/10-Halloween");
      expect(response.status).toBe(200);
      expect(response.count).toBe(4);
      expect(response.folders).toBe(3);
    } catch (error) {
      throw error;
    }
  });

  it("should test find all uncached", async () => {
    try {
      const response = await Folder.findAllUncached("2011/10-Halloween");
      expect(response.status).toBe(200);
      expect(response.count).toBe(3);
      expect(response.folders).toBe(2);
    } catch (error) {
      throw error;
    }
  });

  it("should test find in originals", async () => {
    try {
      const response = await Folder.originals("2011/10-Halloween", { recursive: true });
      expect(response.status).toBe(200);
      expect(response.count).toBe(4);
      expect(response.folders).toBe(3);
    } catch (error) {
      throw error;
    }
  });

  it("should test search", async () => {
    try {
      const response = await Folder.search("2011/10-Halloween", { recursive: true, uncached: true });
      expect(response.status).toBe(200);
      expect(response.count).toBe(3);
      expect(response.folders).toBe(2);
    } catch (error) {
      throw error;
    }
  });
});
