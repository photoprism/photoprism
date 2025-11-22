export const mockGettext = (msg, params = {}) => {
  if (typeof msg !== "string") {
    msg = String(msg ?? "");
  }

  return msg.replace(/%\{(\w+)\}/g, (_, key) => {
    if (params && Object.prototype.hasOwnProperty.call(params, key)) {
      const value = params[key];
      return value === undefined || value === null ? "" : String(value);
    }
    return "";
  });
};

export const mockPgettext = (_ctx, msg, params = {}) => mockGettext(msg, params);
