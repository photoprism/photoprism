import Api from "common/api";

class Session {
    /**
     * @param {Storage} storage
     */
    constructor (storage) {
        this.storage = storage;
        this.storageKey = "session_key";
        this.storageNoAuth = "session_no_auth";
        const lastKey = this.getKey()
        if (!!lastKey) {
            this.setKey(lastKey);
        }
    }

    setKey (token) {
        this.storage.setItem(this.storageKey, token);
        this.storage.removeItem(this.storageNoAuth);
        Api.defaults.headers.common["X-Auth-Key"] = "default";
        Api.defaults.headers.common["X-Auth-Secret"] = token;
    }

    getKey () {
        return this.storage.getItem(this.storageKey);
    }

    deleteKey () {
        this.storage.removeItem(this.storageKey);
        delete Api.defaults.headers.common["X-Auth-Key"];
        delete Api.defaults.headers.common["X-Auth-Secret"];
    }

    async isAuthed () {
        try {
            const resp = await Api.get("/ping")
            return resp.status === 200
        } catch (e) {
            return false
        }
    }

    logout () {
        this.deleteKey();
        window.location = "/";
    }

    setNoAuthMode (value) {
        this.storage.setItem(this.storageNoAuth, value);
    }

    getNoAuthMode () {
        return this.storage.getItem(this.storageNoAuth);
    }
}

export default Session;
