import { createGettext as vue3Gettext } from "vue3-gettext";

function interpolate(message, params = {}) {
  if (message === null || message === undefined) {
    return "";
  }

  const text = String(message);

  if (!params || typeof params !== "object") {
    return text;
  }

  return text.replace(/%\{(\w+)\}/g, (_, key) => {
    if (!Object.prototype.hasOwnProperty.call(params, key)) {
      return "";
    }

    const value = params[key];
    return value === undefined || value === null ? "" : String(value);
  });
}

export let gettext = {
  $gettext: (msgid, params) => interpolate(msgid, params),
  $ngettext: (msgid, plural, n, params) => interpolate(n > 1 ? plural : msgid, params),
  $pgettext: (context, msgid, params) => interpolate(msgid, params),
  $npgettext: (domain, context, msgid, plural, n, params) => interpolate(n > 1 ? plural : msgid, params),
};

export function T(msgid, params) {
  return gettext.$gettext(msgid, params);
}

export function $gettext(msgid, params) {
  return gettext.$gettext(msgid, params);
}

export function $ngettext(msgid, plural, n) {
  return gettext.$ngettext(msgid, plural, n);
}

export function $pgettext(context, msgid, params) {
  return gettext.$pgettext(context, msgid, params);
}

export function $npgettext(domain, context, msgid, plural, n) {
  return gettext.$npgettext(domain, context, msgid, plural, n);
}

export function createGettext(config) {
  gettext = vue3Gettext({
    translations: config.translations,
    silent: true, // !config.values.debug,
    defaultLanguage: config.getLanguageLocale(),
    // autoAddKeyAttributes: true,
  });

  return gettext;
}
