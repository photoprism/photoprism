import { describe, it, expect, beforeEach, afterEach } from "vitest";
import "../fixtures";
import { Subject, BatchSize, MaxLength, BirthYearMin } from "model/subject";

describe("model/subject", () => {
  let originalBatchSize;

  beforeEach(() => {
    originalBatchSize = Subject.batchSize();
  });

  afterEach(() => {
    Subject.setBatchSize(originalBatchSize);
  });

  // Pins per-field caps to the backend VARCHAR columns on internal/entity/subject.go
  // so client-side validation moves in lockstep with the server.
  it("MaxLength mirrors the backend VARCHAR caps", () => {
    expect(MaxLength).toEqual({
      Name: 160,
    });
    expect(Object.isFrozen(MaxLength)).toBe(true);
  });

  it("trimInputs() strips whitespace from MaxLength string fields", () => {
    const subject = new Subject({ Name: "  Alice  " });
    subject.trimInputs();
    expect(subject.Name).toBe("Alice");
  });

  it("should get face defaults", () => {
    const values = {};
    const subject = new Subject(values);
    const result = subject.getDefaults();
    expect(result.UID).toBe("");
    expect(result.Favorite).toBe(false);
  });

  it("should get route view", () => {
    const values = { UID: "s123ghytrfggd", Type: "person", Src: "manual" };
    const subject = new Subject(values);
    const result = subject.route("test");
    expect(result.name).toBe("test");
    expect(result.query.q).toBe("subject:s123ghytrfggd");
    const values2 = {
      UID: "s123ghytrfggd",
      Type: "person",
      Src: "manual",
      Name: "Jane Doe",
      Slug: "jane-doe",
    };
    const subject2 = new Subject(values2);
    const result2 = subject2.route("test");
    expect(result2.name).toBe("test");
    expect(result2.query.q).toBe("person:jane-doe");
  });

  it("should return classes", () => {
    const values = {
      UID: "s123ghytrfggd",
      Type: "person",
      Src: "manual",
      Name: "Jane Doe",
      Slug: "jane-doe",
      Favorite: false,
      Excluded: true,
      Private: true,
      Hidden: true,
    };
    const subject = new Subject(values);
    const result = subject.classes(true);
    expect(result).toContain("is-subject");
    expect(result).toContain("uid-s123ghytrfggd");
    expect(result).toContain("is-selected");
    expect(result).not.toContain("is-favorite");
    expect(result).toContain("is-private");
    expect(result).toContain("is-excluded");
    expect(result).toContain("is-hidden");
    const values2 = {
      UID: "s123ghytrfggd",
      Type: "person",
      Src: "manual",
      Name: "Jane Doe",
      Slug: "jane-doe",
      Favorite: true,
      Excluded: false,
      Private: false,
    };
    const subject2 = new Subject(values2);
    const result2 = subject2.classes(false);
    expect(result2).toContain("is-subject");
    expect(result2).toContain("uid-s123ghytrfggd");
    expect(result2).not.toContain("is-selected");
    expect(result2).toContain("is-favorite");
    expect(result2).not.toContain("is-private");
    expect(result2).not.toContain("is-excluded");
  });

  it("should get subject entity name", () => {
    const values = {
      UID: "s123ghytrfggd",
      Type: "person",
      Src: "manual",
      Name: "Jane Doe",
      Slug: "jane-doe",
      Favorite: false,
      Excluded: true,
      Private: true,
    };
    const subject = new Subject(values);
    const result = subject.getEntityName();
    expect(result).toBe("jane-doe");
  });

  it("should get subject title", () => {
    const values = {
      UID: "s123ghytrfggd",
      Type: "person",
      Src: "manual",
      Name: "Jane Doe",
      Slug: "jane-doe",
      Favorite: false,
      Excluded: true,
      Private: true,
    };
    const subject = new Subject(values);
    const result = subject.getTitle();
    expect(result).toBe("Jane Doe");
  });

  it("should get thumbnail url", () => {
    const values = {
      UID: "s123ghytrfggd",
      Type: "person",
      Src: "manual",
      Name: "Jane Doe",
      Slug: "jane-doe",
      Favorite: false,
      Excluded: true,
      Private: true,
      Thumb: "nicethumb",
    };
    const subject = new Subject(values);
    const result = subject.thumbnailUrl("xyz");
    expect(result).toBe("/api/v1/t/nicethumb/public/xyz");
    const result2 = subject.thumbnailUrl();
    expect(result2).toBe("/api/v1/t/nicethumb/public/tile_160");
    const values2 = {
      UID: "s123ghytrfggd",
      Type: "person",
      Src: "manual",
      Name: "Jane Doe",
      Slug: "jane-doe",
      Favorite: false,
      Excluded: true,
      Private: true,
    };
    const subject2 = new Subject(values2);
    const result3 = subject2.thumbnailUrl("xyz");
    expect(result3).toBe("/api/v1/svg/portrait");
  });

  it("should get date string", () => {
    const values = {
      UID: "s123ghytrfggd",
      Type: "person",
      Src: "manual",
      Name: "Jane Doe",
      Slug: "jane-doe",
      Favorite: false,
      Excluded: true,
      Private: true,
      Thumb: "nicethumb",
      CreatedAt: "2012-07-08T14:45:39Z",
    };
    const subject = new Subject(values);
    const result = subject.getDateString();
    expect(result.replaceAll("\u202f", " ")).toBe("Jul 8, 2012, 2:45 PM");
  });

  it("should like subject", () => {
    const values = {
      UID: "s123ghytrfggd",
      Type: "person",
      Src: "manual",
      Name: "Jane Doe",
      Slug: "jane-doe",
      Favorite: false,
    };
    const subject = new Subject(values);
    expect(subject.Favorite).toBe(false);
    subject.like();
    expect(subject.Favorite).toBe(true);
  });

  it("should unlike subject", () => {
    const values = {
      UID: "s123ghytrfggd",
      Type: "person",
      Src: "manual",
      Name: "Jane Doe",
      Slug: "jane-doe",
      Favorite: true,
    };
    const subject = new Subject(values);
    expect(subject.Favorite).toBe(true);
    subject.unlike();
    expect(subject.Favorite).toBe(false);
  });

  it("should toggle like", () => {
    const values = {
      UID: "s123ghytrfggd",
      Type: "person",
      Src: "manual",
      Name: "Jane Doe",
      Slug: "jane-doe",
      Favorite: true,
    };
    const subject = new Subject(values);
    expect(subject.Favorite).toBe(true);
    subject.toggleLike();
    expect(subject.Favorite).toBe(false);
    subject.toggleLike();
    expect(subject.Favorite).toBe(true);
  });

  it("show and hide subject", () => {
    const values = {
      UID: "s123ghytrfggd",
      Type: "person",
      Src: "manual",
      Name: "Jane Doe",
      Slug: "jane-doe",
      Hidden: true,
    };
    const subject = new Subject(values);
    expect(subject.Hidden).toBe(true);
    subject.show();
    expect(subject.Hidden).toBe(false);
    subject.hide();
    expect(subject.Hidden).toBe(true);
  });

  it("should toggle hidden", () => {
    const values = {
      UID: "s123ghytrfggd",
      Type: "person",
      Src: "manual",
      Name: "Jane Doe",
      Slug: "jane-doe",
      Hidden: true,
    };
    const subject = new Subject(values);
    expect(subject.Hidden).toBe(true);
    subject.toggleHidden();
    expect(subject.Hidden).toBe(false);
    subject.toggleHidden();
    expect(subject.Hidden).toBe(true);
  });

  it("should return batch size", () => {
    expect(Subject.batchSize()).toBe(BatchSize);
    Subject.setBatchSize(30);
    expect(Subject.batchSize()).toBe(30);
  });

  it("should get collection resource", () => {
    const result = Subject.getCollectionResource();
    expect(result).toBe("subjects");
  });

  it("should get model name", () => {
    const result = Subject.getModelName();
    expect(result).toBe("Person");
  });

  describe("birthday", () => {
    const withTimezone = (tz, fn) => {
      const before = process.env.TZ;
      process.env.TZ = tz;
      try {
        fn();
      } finally {
        // Deleted rather than reassigned: an unset TZ reads back as undefined, and writing that
        // stringifies to "undefined" and leaks a bogus zone into every later test in the worker.
        if (before === undefined) {
          delete process.env.TZ;
        } else {
          process.env.TZ = before;
        }
      }
    };

    it("getBirthday() reads the stored day, not the stored instant", () => {
      // Run west of Greenwich, where parsing "1990-08-01T00:00:00Z" as an instant lands on July 31.
      withTimezone("America/New_York", () => {
        expect(new Date("1990-08-01T00:00:00Z").getDate()).toBe(31);

        const subject = new Subject({ UID: "sbj1", Birthday: "1990-08-01T00:00:00Z" });
        const born = subject.getBirthday();

        expect(born.getFullYear()).toBe(1990);
        expect(born.getMonth()).toBe(7);
        expect(born.getDate()).toBe(1);
      });
    });

    it("BirthYearMin mirrors the backend bound", () => {
      // The picker's lower bound and the API's rejection have to name the same year, or the dialog
      // offers a date that saving refuses.
      expect(BirthYearMin).toBe(1800);
    });

    it("getBirthday() reads the day the instant stores, whatever offset it arrives in", () => {
      // A driver configured with a non-UTC loc renders the same instant with an offset. The stored
      // day is what the check reading it compares, so the rendering must not be able to move it.
      withTimezone("America/New_York", () => {
        const utc = new Subject({ UID: "sbj1", Birthday: "1990-08-01T00:00:00Z" });
        const offset = new Subject({ UID: "sbj2", Birthday: "1990-07-31T20:00:00-04:00" });

        expect(offset.getBirthday()).toEqual(utc.getBirthday());
        expect(offset.getBirthday().getDate()).toBe(1);
      });
    });

    it("getBirthday() returns null when the field is unset or unusable", () => {
      expect(new Subject({ UID: "sbj1" }).getBirthday()).toBeNull();
      expect(new Subject({ UID: "sbj1", Birthday: null }).getBirthday()).toBeNull();
      expect(new Subject({ UID: "sbj1", Birthday: "" }).getBirthday()).toBeNull();
      expect(new Subject({ UID: "sbj1", Birthday: "not a date" }).getBirthday()).toBeNull();
    });

    it("setBirthday() stores the picked day at UTC midnight", () => {
      // East of Greenwich, where toISOString() of a local midnight would send the day before.
      withTimezone("Europe/Berlin", () => {
        const subject = new Subject({ UID: "sbj1" });
        const picked = new Date(1990, 7, 1);

        expect(picked.toISOString().slice(0, 10)).toBe("1990-07-31");

        subject.setBirthday(picked);
        expect(subject.Birthday).toBe("1990-08-01T00:00:00Z");
      });
    });

    it("setBirthday() clears the field for anything that is not a date", () => {
      const subject = new Subject({ UID: "sbj1", Birthday: "1990-08-01T00:00:00Z" });

      subject.setBirthday(null);
      expect(subject.Birthday).toBeNull();

      subject.setBirthday("1990-08-01");
      expect(subject.Birthday).toBeNull();

      subject.setBirthday(new Date("nope"));
      expect(subject.Birthday).toBeNull();
    });

    it("only ships the birthday when it changed", () => {
      const subject = new Subject({ UID: "sbj1", Name: "Alice", Birthday: "1990-08-01T00:00:00Z" });

      expect(subject.getValues(true)).toEqual({});

      subject.setBirthday(new Date(1991, 8, 2));
      expect(subject.getValues(true)).toEqual({ Birthday: "1991-09-02T00:00:00Z" });

      subject.setBirthday(null);
      expect(subject.getValues(true)).toEqual({ Birthday: null });
    });
  });
});
