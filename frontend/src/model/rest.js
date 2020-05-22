import Api from "common/api";
import Form from "common/form";
import Model from "./model";

export class Rest extends Model {
    getId() {
        return this.ID;
    }

    hasId() {
        return !!this.getId();
    }

    clone() {
        return new this.constructor(this.getValues());
    }

    find(id, params) {
        return Api.get(this.getEntityResource(id), params).then((response) => Promise.resolve(new this.constructor(response.data)));
    }

    save() {
        if (this.hasId()) {
            return this.update();
        }

        return Api.post(this.constructor.getCollectionResource(), this.getValues()).then((response) => Promise.resolve(this.setValues(response.data)));
    }

    update() {
        return Api.put(this.getEntityResource(), this.getValues(true)).then((response) => Promise.resolve(this.setValues(response.data)));
    }

    remove() {
        return Api.delete(this.getEntityResource()).then(() => Promise.resolve(this));
    }

    getEditForm() {
        return Api.options(this.getEntityResource()).then(response => Promise.resolve(new Form(response.data)));
    }

    getEntityResource(id) {
        if (!id) {
            id = this.getId();
        }

        return this.constructor.getCollectionResource() + "/" + id;
    }

    getEntityName() {
        return this.constructor.getModelName() + " " + this.getId();
    }

    addLink(password, expires, comment, edit) {
        expires = expires ? parseInt(expires) : 0;
        comment = !!comment;
        edit = !!edit;
        const values = {password, expires, comment, edit};
        return Api.post(this.getEntityResource() + "/link", values).then((response) => Promise.resolve(this.setValues(response.data)));
    }

    static getCollectionResource() {
        throw new Error("getCollectionResource() needs to be implemented");
    }

    static getCreateResource() {
        return this.getCollectionResource();
    }

    static getCreateForm() {
        return Api.options(this.getCreateResource()).then(response => Promise.resolve(new Form(response.data)));
    }

    static getModelName() {
        return "Item";
    }

    static getSearchForm() {
        return Api.options(this.getCollectionResource()).then(response => Promise.resolve(new Form(response.data)));
    }

    static search(params) {
        const options = {
            params: params,
        };

        return Api.get(this.getCollectionResource(), options).then((response) => {
            let count = response.data.length;
            let limit = 0;
            let offset = 0;

            if (response.headers) {
                if (response.headers["x-count"]) {
                    count = parseInt(response.headers["x-count"]);
                }

                if (response.headers["x-limit"]) {
                    limit = parseInt(response.headers["x-limit"]);
                }

                if (response.headers["x-offset"]) {
                    offset = parseInt(response.headers["x-offset"]);
                }
            }

            response.models = [];
            response.count = count;
            response.limit = limit;
            response.offset = offset;

            for (let i = 0; i < response.data.length; i++) {
                response.models.push(new this(response.data[i]));
            }

            return Promise.resolve(response);
        });
    }
}

export default Rest;
