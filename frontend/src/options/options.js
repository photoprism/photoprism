import { timeZonesNames } from "@vvo/tzdb";
import { $gettext } from "common/gettext";
import { Info } from "luxon";
import { $config } from "app/session";
import * as media from "common/media";
import * as locales from "../locales";

export const GmtOffsets = [
  { ID: "GMT", Name: "Etc/GMT" },
  { ID: "UTC+1", Name: "Etc/GMT+01:00" },
  { ID: "UTC+2", Name: "Etc/GMT+02:00" },
  { ID: "UTC+3", Name: "Etc/GMT+03:00" },
  { ID: "UTC+4", Name: "Etc/GMT+04:00" },
  { ID: "UTC+5", Name: "Etc/GMT+05:00" },
  { ID: "UTC+6", Name: "Etc/GMT+06:00" },
  { ID: "UTC+7", Name: "Etc/GMT+07:00" },
  { ID: "UTC+8", Name: "Etc/GMT+08:00" },
  { ID: "UTC+9", Name: "Etc/GMT+09:00" },
  { ID: "UTC+10", Name: "Etc/GMT+10:00" },
  { ID: "UTC+11", Name: "Etc/GMT+11:00" },
  { ID: "UTC+12", Name: "Etc/GMT+12:00" },
  { ID: "UTC-1", Name: "Etc/GMT-01:00" },
  { ID: "UTC-2", Name: "Etc/GMT-02:00" },
  { ID: "UTC-3", Name: "Etc/GMT-03:00" },
  { ID: "UTC-4", Name: "Etc/GMT-04:00" },
  { ID: "UTC-5", Name: "Etc/GMT-05:00" },
  { ID: "UTC-6", Name: "Etc/GMT-06:00" },
  { ID: "UTC-7", Name: "Etc/GMT-07:00" },
  { ID: "UTC-8", Name: "Etc/GMT-08:00" },
  { ID: "UTC-9", Name: "Etc/GMT-09:00" },
  { ID: "UTC-10", Name: "Etc/GMT-10:00" },
  { ID: "UTC-11", Name: "Etc/GMT-11:00" },
  { ID: "UTC-12", Name: "Etc/GMT-12:00" },
];

export const TimeZones = (defaultName) =>
  [
    { ID: "", Name: defaultName ? defaultName : $gettext("Local Time") },
    { ID: "UTC", Name: "UTC" },
  ]
    .concat(timeZonesNames)
    .concat(GmtOffsets);

export const Days = () => {
  let result = [];

  for (let i = 1; i <= 31; i++) {
    result.push({ value: i, text: i.toString().padStart(2, "0") });
  }

  result.push({ value: -1, text: $gettext("Unknown") });

  return result;
};

export const Years = (start) => {
  if (!start) {
    start = 1000;
  }

  let result = [];

  const currentYear = new Date().getUTCFullYear();

  for (let i = currentYear; i >= start; i--) {
    result.push({ value: i, text: i.toString().padStart(4, "0") });
  }

  result.push({ value: -1, text: $gettext("Unknown") });

  return result;
};

export const IndexedYears = () => {
  let result = [];

  if ($config.values.years) {
    for (let i = 0; i < $config.values.years.length; i++) {
      result.push({
        value: parseInt($config.values.years[i]),
        text: $config.values.years[i].toString(),
      });
    }
  }

  result.push({ value: -1, text: $gettext("Unknown") });

  return result;
};

export const Months = () => {
  let result = [];

  const months = Info.months("long");

  for (let i = 0; i < months.length; i++) {
    result.push({ value: i + 1, text: months[i] });
  }

  result.push({ value: -1, text: $gettext("Unknown") });

  return result;
};

export const MonthsShort = () => {
  let result = [];

  for (let i = 1; i <= 12; i++) {
    result.push({ value: i, text: i.toString().padStart(2, "0") });
  }

  result.push({ value: -1, text: $gettext("Unknown") });

  return result;
};

// Available locales sorted by region and alphabet.
export const Languages = () => (window.__LOCALES__ ? window.__LOCALES__ : locales.Options);

export const ItemsPerPage = () => [
  { text: "10", title: "10", value: 10 },
  { text: "20", title: "20", value: 20 },
  { text: "50", title: "50", value: 50 },
  { text: "100", title: "100", value: 100 },
];

export const StartPages = (features) => [
  { value: "default", text: $gettext("Default"), visible: true },
  { value: "browse", text: $gettext("Search"), props: { disabled: !features?.library } },
  { value: "albums", text: $gettext("Albums"), props: { disabled: !features?.albums } },
  { value: "media", text: $gettext("Media"), props: { disabled: !features?.videos } },
  { value: "videos", text: $gettext("Videos"), props: { disabled: !features?.videos } },
  { value: "people", text: $gettext("People"), props: { disabled: !(features?.people && features?.edit) } },
  { value: "favorites", text: $gettext("Favorites"), props: { disabled: !features?.favorites } },
  { value: "places", text: $gettext("Places"), props: { disabled: !features?.places } },
  { value: "calendar", text: $gettext("Calendar"), props: { disabled: !features?.calendar } },
  { value: "moments", text: $gettext("Moments"), props: { disabled: !features?.moments } },
  { value: "labels", text: $gettext("Labels"), props: { disabled: !features?.labels } },
  { value: "folders", text: $gettext("Folders"), props: { disabled: !features?.folders } },
];

export const MapsAnimate = () => [
  {
    text: $gettext("None"),
    value: 0,
  },
  {
    text: $gettext("Fast"),
    value: 2500,
  },
  {
    text: $gettext("Medium"),
    value: 6250,
  },
  {
    text: $gettext("Slow"),
    value: 10000,
  },
];

export const MapsStyle = (experimental) => {
  const styles = [
    {
      title: $gettext("Default"),
      value: "",
      style: "default",
    },
    {
      title: $gettext("Streets"),
      value: "streets",
      style: "streets-v2",
      sponsor: true,
    },
    {
      title: $gettext("Hybrid"),
      value: "hybrid",
      style: "414c531c-926d-4164-a057-455a215c0eee",
      sponsor: true,
    },
    {
      title: $gettext("Satellite"),
      value: "satellite",
      style: "0195eda5-6f09-7acd-8520-ab103fc75810",
      sponsor: true,
    },
    {
      title: $gettext("Outdoor"),
      value: "outdoor",
      style: "outdoor-v2",
      sponsor: true,
    },
    {
      title: $gettext("Terrain"),
      value: "topographique",
      style: "topo-v2",
      sponsor: true,
    },
  ];

  if (experimental) {
    styles.push({
      title: $gettext("Offline"),
      value: "low-resolution",
      style: "low-resolution",
    });
  }

  return styles;
};

export const PhotoTypes = () => [
  {
    text: $gettext("Image"),
    value: media.Image,
  },
  {
    text: $gettext("Raw"),
    value: media.Raw,
  },
  {
    text: $gettext("Live"),
    value: media.Live,
  },
  {
    text: $gettext("Video"),
    value: media.Video,
  },
  {
    text: $gettext("Audio"),
    value: media.Audio,
  },
  {
    text: $gettext("Animated"),
    value: media.Animated,
  },
  {
    text: $gettext("Vector"),
    value: media.Vector,
  },
  {
    text: $gettext("Document"),
    value: media.Document,
  },
];

export const Timeouts = () => [
  {
    text: $gettext("Default"),
    value: "",
  },
  {
    text: $gettext("High"),
    value: "high",
  },
  {
    text: $gettext("Low"),
    value: "low",
  },
  {
    text: $gettext("None"),
    value: "none",
  },
];

export const RetryLimits = () => [
  {
    text: "None",
    value: -1,
  },
  {
    text: "1",
    value: 1,
  },
  {
    text: "2",
    value: 2,
  },
  {
    text: "3",
    value: 3,
  },
  {
    text: "4",
    value: 4,
  },
  {
    text: "5",
    value: 5,
  },
];

export const Intervals = () => [
  { value: 0, text: $gettext("Never") },
  { value: 3600, text: $gettext("1 hour") },
  { value: 3600 * 4, text: $gettext("4 hours") },
  { value: 3600 * 12, text: $gettext("12 hours") },
  { value: 86400, text: $gettext("Daily") },
  { value: 86400 * 2, text: $gettext("Every two days") },
  { value: 86400 * 7, text: $gettext("Once a week") },
];

export const Expires = () => [
  { value: 0, text: $gettext("Never") },
  { value: 86400, text: $gettext("After 1 day") },
  { value: 86400 * 3, text: $gettext("After 3 days") },
  { value: 86400 * 7, text: $gettext("After 7 days") },
  { value: 86400 * 14, text: $gettext("After two weeks") },
  { value: 86400 * 31, text: $gettext("After one month") },
  { value: 86400 * 60, text: $gettext("After two months") },
  { value: 86400 * 365, text: $gettext("After one year") },
];

export const Colors = () => [
  { Example: "#AB47BC", Name: $gettext("Purple"), Slug: "purple" },
  { Example: "#FF00FF", Name: $gettext("Magenta"), Slug: "magenta" },
  { Example: "#EC407A", Name: $gettext("Pink"), Slug: "pink" },
  { Example: "#EF5350", Name: $gettext("Red"), Slug: "red" },
  { Example: "#FFA726", Name: $gettext("Orange"), Slug: "orange" },
  { Example: "#D4AF37", Name: $gettext("Gold"), Slug: "gold" },
  { Example: "#FDD835", Name: $gettext("Yellow"), Slug: "yellow" },
  { Example: "#CDDC39", Name: $gettext("Lime"), Slug: "lime" },
  { Example: "#66BB6A", Name: $gettext("Green"), Slug: "green" },
  { Example: "#009688", Name: $gettext("Teal"), Slug: "teal" },
  { Example: "#00BCD4", Name: $gettext("Cyan"), Slug: "cyan" },
  { Example: "#2196F3", Name: $gettext("Blue"), Slug: "blue" },
  { Example: "#A1887F", Name: $gettext("Brown"), Slug: "brown" },
  { Example: "#F5F5F5", Name: $gettext("White"), Slug: "white" },
  { Example: "#9E9E9E", Name: $gettext("Grey"), Slug: "grey" },
  { Example: "#212121", Name: $gettext("Black"), Slug: "black" },
];

export const FeedbackCategories = () => [
  { value: "feedback", text: $gettext("Product Feedback") },
  { value: "feature", text: $gettext("Feature Request") },
  { value: "bug", text: $gettext("Bug Report") },
  { value: "other", text: $gettext("Other") },
];

export const Thumbs = () => {
  return $config.values.thumbs;
};

export const ThumbSizes = () => {
  const thumbs = Thumbs();
  const result = [{ text: $gettext("Originals"), value: "" }];

  for (let i = 0; i < thumbs.length; i++) {
    let t = thumbs[i];

    result.push({ text: t.w + " × " + t.h, value: t.size });
  }

  return result;
};

export const ThumbFilters = () => [
  { value: "blackman", text: $gettext("Blackman: Lanczos Modification, Less Ringing Artifacts") },
  { value: "lanczos", text: $gettext("Lanczos: Detail Preservation, Minimal Artifacts") },
  { value: "cubic", text: $gettext("Cubic: Moderate Quality, Good Performance") },
  { value: "linear", text: $gettext("Linear: Very Smooth, Best Performance") },
];

export const Gender = () => [
  { value: "male", text: $gettext("Male") },
  { value: "female", text: $gettext("Female") },
  { value: "other", text: $gettext("Other") },
];

export const Orientations = () => [
  { value: 1, text: "0°" },
  { value: 6, text: "90°" },
  { value: 3, text: "180°" },
  { value: 8, text: "270°" },
];

export const AccountTypes = () => [{ value: "webdav", text: $gettext("WebDAV") }];
