import $api from "common/api";
import Model from "./model";
import { Photo } from "model/photo";

export class Batch extends Model {
  constructor(values) {
    super(values);
    this.selectionById = new Map();
  }

  getDefaults() {
    return {
      models: [],
      values: {},
      selection: [],
    };
  }

  getDefaultFormData() {
    return {
      Title: {},
      DetailsSubject: {},
      Caption: {},
      Day: {},
      Month: {},
      Year: {},
      TimeZone: {},
      Country: {},
      Altitude: {},
      Lat: {},
      Lng: {},
      DetailsArtist: {},
      DetailsCopyright: {},
      DetailsLicense: {},
      DetailsKeywords: {},
      Type: {},
      Scan: {},
      Private: {},
      Favorite: {},
      Panorama: {},
      Iso: {},
      FocalLength: {},
      FNumber: {},
      Exposure: {},
      CameraID: {},
      LensID: {},
      Albums: {
        action: "none",
        mixed: false,
        items: [],
      },
      Labels: {
        action: "none",
        mixed: false,
        items: [],
      },
    };
  }

  save(selection, values) {
    return $api.post("batch/photos/edit", { photos: selection, values }).then((response) => {
      if (response?.data?.models?.length) {
        const updatedMap = new Map(
          response.data.models.map((raw) => {
            const photo = new Photo();
            photo.setValues(raw);
            return [photo.UID, photo];
          })
        );

        this.models = this.models.map((existing) => {
          const updated = updatedMap.get(existing.UID);
          if (updated) {
            existing.setValues(updated);
            updatedMap.delete(existing.UID);
          }
          return existing;
        });

        updatedMap.forEach((photo) => {
          this.models.push(photo);
        });
      }

      if (response?.data?.values) {
        this.values = response.data.values;
      }

      return this;
    });
  }

  // load fetches the current selection (+ aggregated form values) and hydrates Photo instances.
  load(selection) {
    return $api.post("batch/photos/edit", { photos: selection }).then((response) => {
      const models = response.data.models || [];
      this.models = models.map((m) => new Photo(m));
      this.values = response.data.values;
      this.setSelections(selection);
      return this;
    });
  }

  setSelections(selection) {
    // The backend only returns models that are still editable (not archived/deleted).
    // Filter the original selection so the dialog counts only photos that can be edited now.
    const models = new Set((this.models || []).map((m) => m.UID));

    this.selection = selection.flatMap((id) => {
      if (!models.has(id)) {
        return [];
      }

      return [{ id: id, selected: true }];
    });

    this.selectionById = new Map(this.selection.map((entry) => [entry.id, entry]));
  }

  isSelected(id) {
    const entry = this.selectionById && this.selectionById.get(id);
    return entry ? entry.selected : null;
  }

  getLengthOfAllSelected() {
    return this.selection.filter((photo) => photo.selected).length;
  }

  toggle(id) {
    const entry = this.selectionById && this.selectionById.get(id);
    if (entry) {
      entry.selected = !entry.selected;
    }
  }

  toggleAll(isToggledAll) {
    this.selection.forEach((element) => {
      element.selected = isToggledAll;
    });
  }

  wasChanged() {
    return super.wasChanged();
  }
}
