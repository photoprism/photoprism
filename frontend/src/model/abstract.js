import Api from "common/api";
import Form from "common/form";

class Abstract {
    constructor(values) {
        this.__originalValues = {};

        if (values) {
            this.setValues(values);
        }
    }

    setValues(values) {
        if(!values) return;

        for(let key in values) {
            if(values.hasOwnProperty(key) && key !== "__originalValues") {
                this[key] = values[key];
                this.__originalValues[key] = values[key];
            }
        }

        return this;
    }

    getValues() {
        const result = {};

        for(let key in this.__originalValues) {
            if(this.__originalValues.hasOwnProperty(key) && key !== "__originalValues") {
                result[key] = this[key];
            }
        }

        return result;
    }

    getId() {
        return this.ID;
    }

    hasId() {
        return !!this.getId();
    }

    find(id, params) {
        return Api.get(this.getEntityResource(id), params).then((response) => Promise.resolve(new this(response.data)));
    }

    save() {
        if (this.hasId()) {
            return this.update();
        }

        return Api.post(this.constructor.getCollectionResource(), this.getValues()).then((response) => Promise.resolve(this.setValues(response.data)));
    }

    update() {
        return Api.put(this.getEntityResource(), this.getValues()).then((response) => Promise.resolve(this.setValues(response.data)));
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
            response.models = [];

            for(let i = 0; i < response.data.length; i++) {
                response.models.push(new this(response.data[i]));
            }

            return Promise.resolve(response);
        });
    }
}

export default Abstract;
