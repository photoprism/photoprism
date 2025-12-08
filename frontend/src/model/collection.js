import $api from "common/api";
import RestModel from "model/rest";

// Collection is the shared base class for Album, Label, and other gallery collections.
// It currently provides shared helpers so lightbox and other components can treat
// collections uniformly.
export class Collection extends RestModel {
  setCover(hash) {
    if (!hash || typeof hash !== "string" || hash.length < 40) {
      console.warn("collection: could not change cover because an invalid hash was specified");
      return;
    }

    if (!this.hasId()) {
      console.warn("collection: could not change cover because the UID is missing");
      return;
    }

    const values = {
      Thumb: hash,
      ThumbSrc: "manual",
    };

    return $api.put(this.getEntityResource(), values).then((resp) => Promise.resolve(this.setValues(resp.data)));
  }
}

export default Collection;
