import Config from "common/config";
import Session from "common/session";
import { reactive } from "vue";

export const $config = new Config(window.localStorage, window.__CONFIG__);
export const $session = reactive(new Session(window.localStorage, $config, window["__SHARED__"]));

export default $session;
