import { createGettext as vue3Gettext } from "vue3-gettext";

export let gettext = {
  $gettext: (msgid) => msgid,
  $ngettext: (msgid, plural, n) => {
    return n > 1 ? plural : msgid;
  },
  $pgettext: (context, msgid) => msgid,
  $npgettext: (context, msgid) => msgid,
};

export function T(msgid) {
  return gettext.$gettext(msgid);
}

export function $gettext(msgid) {
  return gettext.$gettext(msgid);
}

export function $ngettext(msgid, plural, n) {
  return gettext.$ngettext(msgid, plural, n);
}

export function createGettext($config) {
  gettext = vue3Gettext({
    translations: $config.translations,
    silent: true, // !config.values.debug,
    defaultLanguage: $config.getLanguage(),
    // autoAddKeyAttributes: true,
  });

  return gettext;
}
