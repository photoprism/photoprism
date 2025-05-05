import { vi } from "vitest";
import { Settings } from "luxon";

Settings.defaultLocale = "en";
Settings.defaultZoneName = "UTC";

// Mock Config
export const mockConfig = {
  contentUri: "/api/v1",
  previewToken: "public",
  apiUri: "/api/v1",
  baseUri: "",
  staticUri: "/static",
  downloadToken: "2lbh9x09",
  mode: "user",
  debug: false,
};

// Mock RestModel
export class MockRestModel {
  constructor(values) {
    this.__originalValues = {};
    this.setValues(values || {});
  }

  setValues(values) {
    if (!values) return this;

    for (let key in values) {
      if (values.hasOwnProperty(key)) {
        this[key] = values[key];
        this.__originalValues[key] = values[key];
      }
    }

    return this;
  }

  getId() {
    return this.UID || this.ID;
  }

  getValues() {
    return { ...this.__originalValues };
  }

  getEntityResource() {
    return `${this.constructor.getCollectionResource()}/${this.getId()}`;
  }

  update() {
    return Promise.resolve({ success: "ok" });
  }

  static getCollectionResource() {
    return "items";
  }
}

// API Response Helpers
export const mockApiResponse = (data) => {
  return { data };
};

export const mockPutResponse = (data = { success: "ok" }) => {
  return vi.fn().mockResolvedValue(mockApiResponse(data));
};

export const mockDeleteResponse = (data = { success: "ok" }) => {
  return vi.fn().mockResolvedValue(mockApiResponse(data));
};

export const mockGetResponse = (data) => {
  return vi.fn().mockResolvedValue(mockApiResponse(data));
};

export const mockPostResponse = (data = { success: "ok" }) => {
  return vi.fn().mockResolvedValue(mockApiResponse(data));
};

// Global mock variables
export const apiMock = {
  put: mockPutResponse(),
  delete: mockDeleteResponse(),
  get: mockGetResponse(),
  post: mockPostResponse(),
};

// Setup common mocks
export const setupCommonMocks = () => {
  // Mock Model
  vi.mock("model/rest", () => ({
    default: MockRestModel,
  }));

  // Mock API
  vi.mock("common/api", () => ({
    default: apiMock,
  }));

  // Mock session
  vi.mock("app/session", () => ({
    $config: mockConfig,
  }));

  // Mock gettext
  vi.mock("common/gettext", () => ({
    $gettext: vi.fn((text) => text),
  }));
};

// Setup common headers
export const mockHeaders = {
  "Content-Type": "application/json; charset=utf-8",
};

export const setupMarkerMocks = () => {
  apiMock.put.mockImplementation((url, data) => {
    if (url.includes("markers/mBC123ghytr")) {
      return Promise.resolve({ data: { success: "ok" } });
    } else if (url.includes("markers/mCC123ghytr")) {
      return Promise.resolve({ data: { success: "ok" } });
    } else if (url.includes("markers/mDC123ghytr")) {
      return Promise.resolve({ data: { success: "ok", Name: "testname" } });
    }

    return Promise.resolve({ data: { success: "ok" } });
  });

  apiMock.delete.mockImplementation((url) => {
    if (url.includes("markers/mEC123ghytr/subject")) {
      return Promise.resolve({ data: { success: "ok" } });
    }
    return Promise.resolve({ data: { success: "ok" } });
  });
};

export default { setupCommonMocks, setupMarkerMocks };
