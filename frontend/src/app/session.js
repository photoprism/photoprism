import Config from "common/config";
import Session from "common/session";
import { getAppStorage } from "common/storage";
import { reactive } from "vue";

const storage = getAppStorage();

export const $config = new Config(storage, window.__CONFIG__);
export const $session = reactive(new Session(storage, $config, window["__SHARED__"]));

export default $session;
