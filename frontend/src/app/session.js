import Config from "common/config";
import Session from "common/session";

export const config = new Config(window.localStorage, window.__CONFIG__);
export const session = new Session(window.localStorage, config, window["__SHARED__"]);

export default session;
