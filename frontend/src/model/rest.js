import $api from "common/api";
import { Form } from "common/form";
import Model from "model.js";
import Link from "link.js";
import { $gettext } from "common/gettext";

// Rest is the abstract base for models backed by RESTful API resources.
export class Rest extends Model {
  getId() {
    if (this.UID) {
      return this.UID;
    }

    return this.ID ? this.ID : false;
  }

  hasId() {
    return !!this.getId();
  }

  getSlug() {
    return this.Slug ? this.Slug : "";
  }

  clone() {
    return new this.constructor(this.getValues());
  }

  find(id, params) {
    return $api
      .get(this.getEntityResource(id), params)
      .then((resp) => Promise.resolve(new this.constructor(resp.data)));
  }

  load() {
    if (!this.hasId()) {
      return;
    }

    return $api.get(this.getEntityResource(this.getId())).then((resp) => Promise.resolve(this.setValues(resp.data)));
  }

  save() {
    if (this.hasId()) {
      return this.update();
    }

    return $api
      .post(this.constructor.getCollectionResource(), this.getValues())
      .then((resp) => Promise.resolve(this.setValues(resp.data)));
  }

  update() {
    // Get updated values.
    const values = this.getValues(true);

    // Return if no values were changed.
    if (Object.keys(values).length === 0) {
      return Promise.resolve(this);
    }

    // Send PUT request.
    return $api.put(this.getEntityResource(), values).then((resp) => Promise.resolve(this.setValues(resp.data)));
  }

  remove() {
    return $api.delete(this.getEntityResource()).then(() => Promise.resolve(this));
  }

  restore() {
    return $api.put(this.getEntityResource(), { DeletedAt: null }).then(() => Promise.resolve(this));
  }

  getEditForm() {
    return $api.options(this.getEntityResource()).then((resp) => Promise.resolve(new Form(resp.data)));
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

  createLink(password, expires) {
    return $api
      .post(this.getEntityResource() + "/links", {
        Password: password ? password : "",
        Expires: expires ? expires : 0,
        Slug: this.getSlug(),
        Comment: "",
        Perm: 0,
      })
      .then((resp) => Promise.resolve(new Link(resp.data)));
  }

  updateLink(link) {
    let values = link.getValues(false);

    if (link.Token) {
      values["Token"] = link.getToken();
    }

    if (link.Password) {
      values["Password"] = link.Password;
    }

    return $api
      .put(this.getEntityResource() + "/links/" + link.getId(), values)
      .then((resp) => Promise.resolve(link.setValues(resp.data)));
  }

  removeLink(link) {
    return $api
      .delete(this.getEntityResource() + "/links/" + link.getId())
      .then((resp) => Promise.resolve(link.setValues(resp.data)));
  }

  links() {
    return $api.get(this.getEntityResource() + "/links").then((resp) => {
      resp.models = [];
      resp.count = resp.data.length;

      for (let i = 0; i < resp.data.length; i++) {
        resp.models.push(new Link(resp.data[i]));
      }

      return Promise.resolve(resp);
    });
  }

  modelName() {
    return this.constructor.getModelName();
  }

  static getCollectionResource() {
    // Needs to be implemented!
    return "";
  }

  static getCreateResource() {
    return this.getCollectionResource();
  }

  static getCreateForm() {
    return $api.options(this.getCreateResource()).then((resp) => Promise.resolve(new Form(resp.data)));
  }

  static getModelName() {
    return $gettext("Item");
  }

  static getSearchForm() {
    return $api.options(this.getCollectionResource()).then((resp) => Promise.resolve(new Form(resp.data)));
  }

  static limit() {
    return 100000;
  }

  static search(params) {
    const options = {
      params: params,
    };

    return $api.get(this.getCollectionResource(), options).then((resp) => {
      let count = resp.data ? resp.data.length : 0;
      let limit = 0;
      let offset = 0;

      if (resp.headers) {
        if (resp.headers["x-count"]) {
          count = parseInt(resp.headers["x-count"]);
        }

        if (resp.headers["x-limit"]) {
          limit = parseInt(resp.headers["x-limit"]);
        }

        if (resp.headers["x-offset"]) {
          offset = parseInt(resp.headers["x-offset"]);
        }
      }

      resp.models = [];
      resp.count = count;
      resp.limit = limit;
      resp.offset = offset;

      if (count > 0) {
        for (let i = 0; i < resp.data.length; i++) {
          resp.models.push(new this(resp.data[i]));
        }
      }

      return Promise.resolve(resp);
    });
  }
}

export default Rest;
