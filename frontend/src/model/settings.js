import $api from "common/api";
import { defaultMetadataLayout, MetadataView } from "common/metadata";
import Model from "./model";

// Settings stores the nested user/admin settings tree used across the UI.
export class Settings extends Model {
  changed(area, key) {
    if (typeof this.__originalValues[area] === "undefined") {
      return false;
    }

    return this[area][key] !== this.__originalValues[area][key];
  }

  setValues(values, scalarOnly) {
    if (!values) return;

    if (values.maps?.style === "basic" || values.maps?.style === "offline") {
      values.maps.style = "";
    }

    // Ensure display settings exist with defaults.
    if (!values.display) {
      values.display = {
        originals: false,
        retinaLightbox: false,
        retinaThumbnails: false,
        metadata: {
          cards: defaultMetadataLayout(MetadataView.Cards),
          list: defaultMetadataLayout(MetadataView.List),
          lightbox: defaultMetadataLayout(MetadataView.Lightbox),
        },
      };
    } else {
      if (typeof values.display.originals === "undefined") {
        values.display.originals = false;
      }

      if (typeof values.display.retinaLightbox === "undefined") {
        values.display.retinaLightbox = false;
      }

      if (typeof values.display.retinaThumbnails === "undefined") {
        values.display.retinaThumbnails = false;
      }
    }

    if (!values.display.metadata) {
      values.display.metadata = {};
    }

    if (!Array.isArray(values.display.metadata.cards)) {
      values.display.metadata.cards = defaultMetadataLayout(MetadataView.Cards);
    }

    if (!Array.isArray(values.display.metadata.list)) {
      values.display.metadata.list = defaultMetadataLayout(MetadataView.List);
    }

    if (!Array.isArray(values.display.metadata.lightbox)) {
      values.display.metadata.lightbox = defaultMetadataLayout(MetadataView.Lightbox);
    }

    super.setValues(values, scalarOnly);

    return this;
  }

  load() {
    return $api.get("settings").then((response) => {
      return Promise.resolve(this.setValues(response.data));
    });
  }

  save() {
    return $api.post("settings", this.getValues(true)).then((response) => Promise.resolve(this.setValues(response.data)));
  }
}

export default Settings;
