import RestModel from "model/rest";

class Clipboard {
    /**
     * @param {Storage} storage
     * @param {string} key
     */
    constructor(storage, key) {
        this.storageKey = key ? key : "clipboard";

        this.storage = storage;
        this.selectionMap = {};
        this.selection = [];
        this.lastId = "";

        this.loadFromStorage();
    }

    loadFromStorage() {
        const photosJson = this.storage.getItem(this.storageKey);

        if (photosJson !== null && typeof photosJson !== "undefined") {
            this.setIds(JSON.parse(photosJson));
        }
    }

    saveToStorage() {
        this.storage.setItem(this.storageKey, JSON.stringify(this.selection));
    }

    toggle(model) {
        if(!model || !(model instanceof RestModel)) {
            console.log("Clipboard::toggle() - not a model:", model);
            return;
        }

        const id = model.getId();
        this.toggleId(id);
    }

    toggleId(id) {
        const index = this.selection.indexOf(id);

        if (index === -1) {
            this.selection.push(id);
            this.selectionMap["id:" + id] = true;
            this.lastId = id;
        } else {
            this.selection.splice(index, 1);
            delete this.selectionMap["id:" + id];
            this.lastId = "";
        }

        this.saveToStorage();
    }

    add(model) {
        if(!model || !(model instanceof RestModel)) {
            console.log("Clipboard::add() - not a model:", model);
            return;
        }

        const id = model.getId();

        this.addId(id);
    }

    addId(id) {
        if (this.hasId(id)) return;

        this.selection.push(id);
        this.selectionMap["id:" + id] = true;
        this.lastId = id;

        this.saveToStorage();
    }

    addRange(rangeEnd, models) {
        if(!models || !models[rangeEnd] || !(models[rangeEnd] instanceof RestModel)) {
            console.warn("Clipboard::addRange() - invalid arguments:", rangeEnd, models);
            return;
        }

        let rangeStart = models.findIndex((photo) => photo.UID === this.lastId);

        if(rangeStart === -1) {
            this.toggle(models[rangeEnd]);
            return 1;
        }

        if (rangeStart > rangeEnd) {
            const newEnd = rangeStart;
            rangeStart = rangeEnd;
            rangeEnd = newEnd;
        }

        for (let i = rangeStart; i <= rangeEnd; i++) {
            this.add(models[i]);
        }

        return (rangeEnd - rangeStart) + 1;
    }

    has(model) {
        if(!model || !(model instanceof RestModel)) {
            console.log("Clipboard::has() - not a model:", model);
            return;
        }

        return this.hasId(model.getId());
    }

    hasId(id) {
        return typeof this.selectionMap["id:" + id] !== "undefined";
    }

    remove(model) {
        if(!model || !(model instanceof RestModel)) {
            console.log("Clipboard::remove() - not a model:", model);
            return;
        }

        this.removeId(model.getId());
    }

    removeId(id) {
        if (!this.hasId(id)) return;

        const index = this.selection.indexOf(id);

        this.selection.splice(index, 1);
        this.lastId = "";
        delete this.selectionMap["id:" + id];

        this.saveToStorage();
    }

    getIds() {
        return this.selection;
    }

    setIds(ids) {
        if (!Array.isArray(ids)) return;

        this.selection = ids;
        this.selectionMap = {};
        this.lastId = "";

        for (let i = 0; i < this.selection.length; i++) {
            this.selectionMap["id:" + this.selection[i]] = true;
        }
    }

    clear() {
        this.lastId = "";
        this.selectionMap = {};
        this.selection.splice(0, this.selection.length);
        this.storage.removeItem(this.storageKey);
    }
}

export default Clipboard;
