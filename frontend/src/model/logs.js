import RestModel from "model/rest";
import { $gettext } from "common/gettext";

const SEGMENT_SPLIT = /\s*[›>]\s*/;

function splitSegments(message) {
  if (!message) {
    return [];
  }

  return String(message)
    .split(SEGMENT_SPLIT)
    .map((part) => part.trim())
    .filter((part) => part.length > 0);
}

// All log levels ordered by severity.
export const AuditSeverityNames = Object.freeze(["panic", "fatal", "error", "warning", "info", "debug", "trace"]);

// Audit logs currently only exist with the following levels:
// error, warning, info, and debug.
export const AuditSeverityOptions = [
  { title: "—", value: "" }, // unselected option
  { title: "Debug", value: "debug" },
  { title: "Info", value: "info" },
  { title: "Warning", value: "warning" },
  { title: "Error", value: "error" },
];

// Colors to use for each level.
export const SeverityPalette = {
  panic: "error",
  fatal: "error",
  error: "error",
  warning: "warning",
  info: "primary",
  debug: "surface-variant",
  trace: "surface-variant",
};

// LogEntry represents an audit log row returned by GET /api/v1/logs/audit.
export class LogEntry extends RestModel {
  getDefaults() {
    return {
      ID: 0,
      Time: "",
      Severity: 0,
      IP: "",
      Message: "",
      Repeated: 0,
    };
  }

  // severityName returns the severity label (e.g., "info") or "info" when unknown.
  severityName() {
    const value = this.Severity;

    if (typeof value === "string") {
      const normalized = value.trim().toLowerCase();

      if (AuditSeverityNames.includes(normalized)) {
        return normalized;
      }

      const numeric = Number(normalized);

      if (Number.isFinite(numeric) && numeric >= 0 && numeric < AuditSeverityNames.length) {
        return AuditSeverityNames[numeric];
      }
    } else if (Number.isFinite(value)) {
      const index = Number(value);

      if (index >= 0 && index < AuditSeverityNames.length) {
        return AuditSeverityNames[index];
      }
    }

    return "info";
  }

  // severityTag returns the uppercase severity tag for badges and chips.
  severityTag() {
    return this.severityName().toUpperCase();
  }

  // hasRepeats indicates whether the log entry was repeated.
  hasRepeats() {
    return Number(this.Repeated) > 0;
  }

  // messageParts splits the log message into segments, less the one holding the client address.
  messageParts() {
    const segments = splitSegments(this.Message);
    const explicitIp = this.ipAddress();

    return segments.filter((segment) => segment && segment !== explicitIp);
  }

  // summary returns the first message part or the entire message if no separator is present.
  summary() {
    const parts = this.messageParts();

    if (parts.length > 0) {
      return parts[0];
    }

    const segments = splitSegments(this.Message);

    if (segments.length > 0) {
      return segments[0];
    }

    return this.Message || "";
  }

  // messageChain joins message parts with the provided separator, falling back to the summary.
  messageChain(separator = " \u203A ") {
    const parts = this.messageParts();

    if (parts.length > 0) {
      return parts.join(separator);
    }

    return this.summary();
  }

  // ipAddress returns the client address the server reported for the entry.
  ipAddress() {
    return (this.IP || "").trim();
  }

  static getCollectionResource() {
    return "logs/audit";
  }

  static getModelName() {
    return $gettext("Audit Log");
  }

  static limit() {
    return 1000;
  }
}

export default LogEntry;
