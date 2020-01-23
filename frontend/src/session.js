import Config from "common/config";
import Session from "common/session";

export const config = new Config(window.localStorage, window.clientConfig);
export const session = new Session(window.localStorage, config);

export default session;
