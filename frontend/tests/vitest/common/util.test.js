import { describe, it, expect } from "vitest";
import $util from "common/util";

describe("$util", () => {
  describe("formatBytes", () => {
    it("should format bytes as KB", () => {
      expect($util.formatBytes(1000)).toBe("1 KB");
      expect($util.formatBytes(2000)).toBe("2 KB");
      expect($util.formatBytes("3000")).toBe("3 KB");
    });

    it("should format bytes as MB", () => {
      expect($util.formatBytes(1048576)).toBe("1.0 MB");
      expect($util.formatBytes(2097152)).toBe("2.0 MB");
      expect($util.formatBytes(3145728)).toBe("3.0 MB");
    });

    it("should format bytes as GB", () => {
      expect($util.formatBytes(1073741824)).toBe("1.0 GB");
      expect($util.formatBytes(2147483648)).toBe("2.0 GB");
      expect($util.formatBytes(3221225472)).toBe("3.0 GB");
    });

    it("should handle zero and falsy values", () => {
      expect($util.formatBytes(0)).toBe("0 KB");
      expect($util.formatBytes(null)).toBe("0 KB");
      expect($util.formatBytes(undefined)).toBe("0 KB");
      expect($util.formatBytes("")).toBe("0 KB");
    });
  });

  describe("truncate", () => {
    it("should truncate text longer than specified length", () => {
      expect($util.truncate("This is a test", 7)).toBe("This i…");
      expect($util.truncate("Hello world!", 5)).toBe("Hell…");
    });

    it("should not truncate text shorter than specified length", () => {
      expect($util.truncate("Test", 10)).toBe("Test");
      expect($util.truncate("Short", 10)).toBe("Short");
    });

    it("should use custom ending if specified", () => {
      expect($util.truncate("This is a test", 7, "...")).toBe("This...");
      expect($util.truncate("Hello world!", 5, " [more]")).toBe(" [more]");
    });

    it("should use default values if not specified", () => {
      expect($util.truncate("This is a very long text that should be truncated")).toBe(
        "This is a very long text that should be truncated"
      );
      // Default length is 100 characters
    });
  });

  describe("capitalize", () => {
    it("should capitalize first letter of each word", () => {
      expect($util.capitalize("hello world")).toBe("Hello World");
      expect($util.capitalize("test string")).toBe("Test String");
    });

    it("should handle empty strings", () => {
      expect($util.capitalize("")).toBe("");
      expect($util.capitalize(null)).toBe("");
      expect($util.capitalize(undefined)).toBe("");
    });

    it("should handle already capitalized text", () => {
      expect($util.capitalize("Hello World")).toBe("Hello World");
      expect($util.capitalize("HELLO WORLD")).toBe("HELLO WORLD");
    });
  });

  describe("ucFirst", () => {
    it("should capitalize only first letter of string", () => {
      expect($util.ucFirst("hello world")).toBe("Hello world");
      expect($util.ucFirst("test string")).toBe("Test string");
    });

    it("should handle empty strings", () => {
      expect($util.ucFirst("")).toBe("");
      expect($util.ucFirst(null)).toBe("");
      expect($util.ucFirst(undefined)).toBe("");
    });

    it("should handle already capitalized text", () => {
      expect($util.ucFirst("Hello world")).toBe("Hello world");
      expect($util.ucFirst("HELLO world")).toBe("HELLO world");
    });
  });

  describe("formatSeconds", () => {
    it("should format seconds as mm:ss", () => {
      expect($util.formatSeconds(0)).toBe("0:00");
      expect($util.formatSeconds(1)).toBe("0:01");
      expect($util.formatSeconds(10)).toBe("0:10");
      expect($util.formatSeconds(60)).toBe("1:00");
      expect($util.formatSeconds(65)).toBe("1:05");
      expect($util.formatSeconds(125)).toBe("2:05");
    });

    it("should handle negative or falsy values", () => {
      expect($util.formatSeconds(-1)).toBe("0:00");
      expect($util.formatSeconds(null)).toBe("0:00");
      expect($util.formatSeconds(undefined)).toBe("0:00");
    });
  });
});
