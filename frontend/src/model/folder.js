import RestModel from "model/rest";
import { DateTime } from "luxon";
import File from "model/file";
import { $config } from "app/session";
import $api from "common/api";
import $util from "common/util";
import { $gettext } from "common/gettext";

export const RootImport = "import";
export const RootOriginals = "originals";

// Folder wraps server folder metadata and provides helpers for imports/originals trees.
export class Folder extends RestModel {
  getDefaults() {
    return {
      Folder: true,
      Path: "",
      Root: "",
      UID: "",
      Type: "",
      Title: "",
      Category: "",
      Description: "",
      Order: "",
      Country: "",
      Year: "",
      Month: "",
      Favorite: false,
      Private: false,
      Ignore: false,
      Watch: false,
      FileCount: 0,
      CreatedAt: "",
      UpdatedAt: "",
    };
  }

  classes(selected) {
    let classes = ["is-folder", "uid-" + this.UID];

    if (this.Favorite) classes.push("is-favorite");
    if (this.Private) classes.push("is-private");
    if (selected) classes.push("is-selected");

    return classes;
  }

  baseName(truncate) {
    let result = this.Path;
    const slash = result.lastIndexOf("/");

    if (slash >= 0) {
      result = this.Path.substring(slash + 1);
    }

    if (truncate) {
      result = $util.truncate(result, truncate, "…");
    }

    return result;
  }

  isFile() {
    return false;
  }

  getEntityName() {
    return this.Root + "/" + this.Path;
  }

  thumbnailUrl(size) {
    return `${$config.contentUri}/folders/t/${this.UID}/${$config.previewToken}/${size}`;
  }

  getDateString() {
    return DateTime.fromISO(this.CreatedAt).toLocaleString(DateTime.DATETIME_MED);
  }

  toggleLike() {
    this.Favorite = !this.Favorite;

    if (this.Favorite) {
      return $api.post(this.getEntityResource() + "/like");
    } else {
      return $api.delete(this.getEntityResource() + "/like");
    }
  }

  like() {
    this.Favorite = true;
    return $api.post(this.getEntityResource() + "/like");
  }

  unlike() {
    this.Favorite = false;
    return $api.delete(this.getEntityResource() + "/like");
  }

  static findAll(path) {
    return this.search(path, { recursive: true });
  }

  static findAllUncached(path) {
    return this.search(path, { recursive: true, uncached: true });
  }

  static originals(path, params) {
    if (!path || path[0] !== "/") {
      path = "/" + path;
    }

    return this.search(RootOriginals + path, params);
  }

  static search(path, params) {
    const options = {
      params: params,
    };

    if (!path || path[0] !== "/") {
      path = "/" + path;
    }

    // "#" chars in path names must be explicitly escaped,
    // see https://github.com/photoprism/photoprism/issues/3695
    path = path.replaceAll(":", "%3A").replaceAll("#", "%23");

    return $api.get(this.getCollectionResource() + path, options).then((response) => {
      let folders = response.data.folders ? response.data.folders : [];
      let files = response.data.files ? response.data.files : [];

      let count = folders.length + files.length;

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
      response.files = files.length;
      response.folders = folders.length;
      response.count = count;
      response.limit = limit;
      response.offset = offset;

      for (let i = 0; i < folders.length; i++) {
        response.models.push(new this(folders[i]));
      }

      for (let i = 0; i < files.length; i++) {
        response.models.push(new File(files[i]));
      }

      return Promise.resolve(response);
    });
  }

  static getCollectionResource() {
    return "folders";
  }

  static getModelName() {
    return $gettext("Folder");
  }
}

export default Folder;
