import { DateTime } from "luxon";

import * as formats from "options/formats";
import * as media from "common/media";
import $util from "common/util";
import { $gettext } from "common/gettext";

export const MetadataView = {
  Cards: "cards",
  List: "list",
  Lightbox: "lightbox",
};

export const defaultMetadataLayouts = {
  [MetadataView.Cards]: ["caption", "date", "keywords"],
  [MetadataView.List]: ["filename", "date", "camera", "lens", "exposure"],
  [MetadataView.Lightbox]: ["date", "caption", "keywords", "camera", "lens", "exposure", "filename", "fileInfo"],
};

const detailsBackedFields = new Set(["keywords"]);

const metadataFields = [
  { id: "title", icon: "mdi-format-title", label: () => $gettext("Title"), views: [MetadataView.Cards, MetadataView.List, MetadataView.Lightbox] },
  { id: "caption", icon: "mdi-text", label: () => $gettext("Caption"), views: [MetadataView.Cards, MetadataView.List, MetadataView.Lightbox] },
  { id: "keywords", icon: "mdi-tag-multiple", label: () => $gettext("Keywords"), views: [MetadataView.Cards, MetadataView.List, MetadataView.Lightbox] },
  { id: "date", icon: "mdi-calendar-range", label: () => $gettext("Taken"), views: [MetadataView.Cards, MetadataView.List, MetadataView.Lightbox] },
  { id: "filename", icon: "mdi-film", label: () => $gettext("File Name"), views: [MetadataView.Cards, MetadataView.List, MetadataView.Lightbox] },
  { id: "camera", icon: "mdi-camera", label: () => $gettext("Camera"), views: [MetadataView.Cards, MetadataView.List, MetadataView.Lightbox] },
  { id: "lens", icon: "mdi-camera-iris", label: () => $gettext("Lens"), views: [MetadataView.Cards, MetadataView.List, MetadataView.Lightbox] },
  { id: "exposure", icon: "mdi-tune-variant", label: () => $gettext("Exposure"), views: [MetadataView.Cards, MetadataView.List, MetadataView.Lightbox] },
  { id: "fileInfo", icon: "mdi-image", label: () => $gettext("Image"), views: [MetadataView.Cards, MetadataView.List, MetadataView.Lightbox] },
  { id: "location", icon: "mdi-map-marker", label: () => $gettext("Location"), views: [MetadataView.Cards, MetadataView.List, MetadataView.Lightbox] },
  { id: "artist", icon: "mdi-account", label: () => $gettext("Artist"), views: [MetadataView.Lightbox] },
  { id: "copyright", icon: "mdi-copyright", label: () => $gettext("Copyright"), views: [MetadataView.Lightbox] },
  { id: "license", icon: "mdi-scale-balance", label: () => $gettext("License"), views: [MetadataView.Lightbox] },
  { id: "software", icon: "mdi-application-cog", label: () => $gettext("Software"), views: [MetadataView.Lightbox] },
  { id: "notes", icon: "mdi-note-text", label: () => $gettext("Notes"), views: [MetadataView.Lightbox] },
  { id: "subject", icon: "mdi-shape", label: () => $gettext("Subject"), views: [MetadataView.Lightbox] },
  { id: "labels", icon: "mdi-label-multiple", label: () => $gettext("Labels"), views: [MetadataView.Lightbox] },
];

// metadataFieldOptions returns the selectable metadata fields for a view.
export function metadataFieldOptions(view) {
  return metadataFields
    .filter((field) => field.views.includes(view))
    .map((field) => ({
      id: field.id,
      icon: field.icon,
      label: field.label(),
    }));
}

// defaultMetadataLayout returns the default metadata layout for a view.
export function defaultMetadataLayout(view) {
  return (defaultMetadataLayouts[view] || []).slice();
}

// normalizeMetadataLayout filters unsupported fields while preserving order and duplicates.
export function normalizeMetadataLayout(layout, view) {
  const allowed = new Set(metadataFieldOptions(view).map((field) => field.id));
  const values = Array.isArray(layout) ? layout.filter((field) => typeof field === "string" && allowed.has(field)) : [];

  if (values.length > 0 || Array.isArray(layout)) {
    return values;
  }

  return defaultMetadataLayout(view);
}

// metadataLayout returns a normalized metadata layout from settings.
export function metadataLayout(settings, view) {
  return normalizeMetadataLayout(settings?.display?.metadata?.[view], view);
}

// metadataFieldRequiresDetails returns true if a field depends on photo details.
export function metadataFieldRequiresDetails(fieldId) {
  return detailsBackedFields.has(fieldId);
}

// metadataLayoutRequiresDetails returns true if any configured field needs photo details.
export function metadataLayoutRequiresDetails(layout) {
  return Array.isArray(layout) && layout.some((fieldId) => metadataFieldRequiresDetails(fieldId));
}

// metadataViewRequiresDetails returns true if a view's active layout needs photo details.
export function metadataViewRequiresDetails(settings, view) {
  return metadataLayoutRequiresDetails(metadataLayout(settings, view));
}

// metadataLabel returns the translated label for a metadata field.
export function metadataLabel(fieldId) {
  return metadataFields.find((field) => field.id === fieldId)?.label() || fieldId;
}

// metadataIcon returns the icon name for a metadata field.
export function metadataIcon(fieldId, model) {
  if (fieldId === "fileInfo") {
    switch (model?.Type) {
      case media.Video:
        return "mdi-movie";
      case media.Live:
        return "mdi-play-circle-outline";
      case media.Animated:
        return "mdi-file-gif-box";
      case media.Vector:
        return "mdi-vector-polyline";
      case media.Document:
        return "mdi-text-box";
      default:
        return "mdi-image";
    }
  }

  return metadataFields.find((field) => field.id === fieldId)?.icon || "mdi-information";
}

// metadataText returns the display value for a metadata field.
export function metadataText(model, fieldId) {
  if (!model) {
    return "";
  }

  switch (fieldId) {
    case "title":
      return stringValue(model.Title);
    case "caption":
      return stringValue(model.Caption);
    case "keywords":
      return keywordsValue(model);
    case "date":
      return dateValue(model);
    case "filename":
      return fileNameValue(model);
    case "camera":
      return cameraValue(model);
    case "lens":
      return lensValue(model);
    case "exposure":
      return exposureValue(model);
    case "fileInfo":
      return fileInfoValue(model);
    case "location":
      return locationValue(model);
    case "artist":
      return detailValue(model, "Artist", "DetailsArtist");
    case "copyright":
      return detailValue(model, "Copyright", "DetailsCopyright");
    case "license":
      return detailValue(model, "License", "DetailsLicense");
    case "software":
      return detailValue(model, "Software", "DetailsSoftware");
    case "notes":
      return detailValue(model, "Notes", "DetailsNotes");
    case "subject":
      return detailValue(model, "Subject", "DetailsSubject");
    case "labels":
      return labelsValue(model);
    default:
      return "";
  }
}

// hasMetadataText returns true if a metadata field has a display value.
export function hasMetadataText(model, fieldId) {
  return metadataText(model, fieldId) !== "";
}

function stringValue(value) {
  return typeof value === "string" ? value.trim() : "";
}

function detailValue(model, detailKey, flatKey) {
  const nested = stringValue(model?.Details?.[detailKey]);
  if (nested) {
    return nested;
  }

  return stringValue(model?.[flatKey]);
}

function keywordsValue(model) {
  const flat = stringValue(model?.DetailsKeywords);
  if (flat) {
    return flat;
  }

  const nested = stringValue(model?.Details?.Keywords);
  if (nested) {
    return nested;
  }

  if (Array.isArray(model?.Keywords)) {
    return model.Keywords.map((keyword) => stringValue(keyword?.Name || keyword?.Title || keyword))
      .filter(Boolean)
      .join(", ");
  }

  return "";
}

function labelsValue(model) {
  if (!Array.isArray(model?.Labels)) {
    return "";
  }

  return model.Labels.map((label) => stringValue(label?.Name || label?.Title || label))
    .filter(Boolean)
    .join(", ");
}

function dateValue(model) {
  const takenAtLocal = stringValue(model?.TakenAtLocal);

  if (!takenAtLocal) {
    return "";
  }

  const dateTime = DateTime.fromISO(takenAtLocal, { zone: "UTC" });

  if (!dateTime.isValid) {
    return takenAtLocal;
  }

  if (model?.TimeZone && model.TimeZone !== "Local" && model.TimeZone !== "UTC") {
    return dateTime.setZone(model.TimeZone, { keepLocalTime: true }).toLocaleString(formats.DATE_MED);
  }

  return dateTime.toLocaleString(formats.DATE_MED);
}

function fileNameValue(model) {
  if (typeof model?.getOriginalName === "function") {
    return model.getOriginalName();
  }

  return stringValue(model?.FileName) || stringValue(model?.OriginalName) || stringValue(model?.Name);
}

function cameraValue(model) {
  const result = $util.formatCamera(model?.Camera, model?.CameraID, model?.CameraMake, model?.CameraModel, true);

  if (result === $gettext("Unknown")) {
    return "";
  }

  return result;
}

function lensValue(model) {
  if (model?.Lens && typeof model.Lens === "object") {
    const lens = [stringValue(model.Lens.Make), stringValue(model.Lens.Model)].filter(Boolean);
    if (lens.length > 0) {
      return Array.from(new Set(lens)).join(" ");
    }
  }

  const lens = [stringValue(model?.LensMake), stringValue(model?.LensModel)].filter(Boolean);

  if (lens.length > 0) {
    return Array.from(new Set(lens)).join(" ");
  }

  return "";
}

function exposureValue(model) {
  const info = [];

  if (model?.Iso) {
    info.push(`ISO ${model.Iso}`);
  }

  if (stringValue(model?.Exposure)) {
    info.push(model.Exposure);
  }

  if (model?.FNumber) {
    info.push(`ƒ/${model.FNumber}`);
  }

  if (model?.FocalLength) {
    info.push(`${model.FocalLength}mm`);
  }

  return info.join(", ");
}

function fileInfoValue(model) {
  if (typeof model?.getTypeInfo === "function") {
    return model.getTypeInfo();
  }

  switch (model?.Type) {
    case media.Video:
    case media.Live:
    case media.Animated:
      if (typeof model?.getVideoInfo === "function") {
        return model.getVideoInfo();
      }
      break;
    case media.Vector:
    case media.Document:
      if (typeof model?.getVectorInfo === "function") {
        return model.getVectorInfo();
      }
      break;
    default:
      if (typeof model?.getImageInfo === "function") {
        return model.getImageInfo();
      }
  }

  return "";
}

function locationValue(model) {
  if (typeof model?.locationInfo === "function") {
    const value = model.locationInfo();
    return value === $gettext("Unknown") ? "" : value;
  }

  return stringValue(model?.PlaceLabel);
}
