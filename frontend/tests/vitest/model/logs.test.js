import { describe, expect, it } from "vitest";
import LogEntry, { AuditSeverityNames } from "model/logs";

describe("model/logs", () => {
  it("provides default values", () => {
    const entry = new LogEntry();

    expect(entry.ID).toBe(0);
    expect(entry.Time).toBe("");
    expect(entry.Message).toBe("");
    expect(entry.hasRepeats()).toBe(false);
    expect(LogEntry.limit()).toBe(1000);
  });

  it("maps severity names and tags", () => {
    const entry = new LogEntry({ Severity: 3 });

    expect(entry.severityName()).toBe("warning");
    expect(entry.severityTag()).toBe("WARNING");

    entry.Severity = 99;
    expect(entry.severityName()).toBe("info");

    entry.Severity = "warning";
    expect(entry.severityName()).toBe("warning");
    expect(entry.severityTag()).toBe("WARNING");

    entry.Severity = "4";
    expect(entry.severityName()).toBe("info");

    entry.Severity = "ERROR";
    expect(entry.severityName()).toBe("error");
  });

  it("splits audit messages into parts", () => {
    const entry = new LogEntry({
      Message: "172.18.0.1 › manage sessions › denied",
      IP: "172.18.0.1",
    });

    expect(entry.messageParts()).toEqual(["manage sessions", "denied"]);
    expect(entry.summary()).toBe("manage sessions");
    expect(entry.messageChain(" > ")).toBe("manage sessions > denied");
    expect(entry.ipAddress()).toBe("172.18.0.1");

    entry.Message = "single message";
    entry.IP = "";
    expect(entry.messageParts()).toEqual(["single message"]);
    expect(entry.summary()).toBe("single message");
    expect(entry.messageChain(" > ")).toBe("single message");
    expect(entry.ipAddress()).toBe("");
  });

  it("derives ip addresses from messages when missing", () => {
    const entry = new LogEntry({
      Message: "::1 › session sess123 › granted",
      IP: "",
    });

    expect(entry.ipAddress()).toBe("::1");
    expect(entry.messageParts()).toEqual(["session sess123", "granted"]);
  });

  it("exposes the severity catalogue", () => {
    expect(Array.isArray(AuditSeverityNames)).toBe(true);
    expect(AuditSeverityNames).toContain("info");
  });
});
