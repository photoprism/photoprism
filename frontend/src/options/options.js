import { timeZonesNames } from "@vvo/tzdb";
import { $gettext } from "common/vm";
import { Info } from "luxon";
import { config } from "app/session";
import {
  MediaImage,
  MediaLive,
  MediaVideo,
  MediaAnimated,
  MediaVector,
  MediaRaw,
} from "model/photo";

export const TimeZones = () =>
  [
    { ID: "UTC", Name: "UTC" },
    { ID: "", Name: $gettext("Local Time") },
  ].concat(timeZonesNames);

export const Days = () => {
  let result = [];

  for (let i = 1; i <= 31; i++) {
    result.push({ value: i, text: i.toString().padStart(2, "0") });
  }

  result.push({ value: -1, text: $gettext("Unknown") });

  return result;
};

export const Years = () => {
  let result = [];

  const currentYear = new Date().getUTCFullYear();

  for (let i = currentYear; i >= 1750; i--) {
    result.push({ value: i, text: i.toString().padStart(4, "0") });
  }

  result.push({ value: -1, text: $gettext("Unknown") });

  return result;
};

export const IndexedYears = () => {
  let result = [];

  if (config.values.years) {
    for (let i = 0; i < config.values.years.length; i++) {
      result.push({
        value: parseInt(config.values.years[i]),
        text: config.values.years[i].toString(),
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
export const Languages = () => [
  {
    text: "English", // English
    translated: "English",
    value: "en",
  },
  {
    text: "Afrikaans", // Afrikaans (South Africa)
    value: "af",
  },
  {
    text: "Bahasa Indonesia", // Bahasa Indonesia
    value: "id",
  },
  {
    text: "Català", // Catalan
    value: "ca",
  },
  {
    text: "Čeština", // Czech
    value: "cs",
  },
  {
    text: "Dansk", // Danish
    value: "da",
  },
  {
    text: "Deutsch", // German
    value: "de",
  },
  {
    text: "Eesti", // Estonian
    value: "et",
  },
  {
    text: "Español", // Spanish
    value: "es",
  },
  {
    text: "Français", // French
    value: "fr",
  },
  {
    text: "Ελληνικά", // Greek
    value: "el",
  },
  {
    text: "עברית", // Hebrew
    value: "he",
    rtl: true,
  },
  {
    text: "Hrvatski", // Croatian
    value: "hr",
  },
  {
    text: "Lietuvis", // Lithuanian
    value: "lt",
  },
  {
    text: "Magyar", // Hungarian
    value: "hu",
  },
  {
    text: "Melayu", // Malay
    value: "ms",
  },
  {
    text: "Norsk (Bokmål)", // Norwegian
    value: "nb",
  },
  {
    text: "Italiano", // Italian
    value: "it",
  },
  {
    text: "Nederlands", // Dutch
    value: "nl",
  },
  {
    text: "Polski", // Polish
    value: "pl",
  },
  {
    text: "Português", // Portuguese (Portugal)
    value: "pt",
  },
  {
    text: "Português do Brasil", // Portuguese (Brazil)
    value: "pt_BR",
  },
  {
    text: "Slovenčina", // Slovak
    value: "sk",
  },
  {
    text: "Slovenščina", // Slovene
    value: "sl",
  },
  {
    text: "Suomi", // Finnish
    value: "fi",
  },
  {
    text: "Svenska", // Swedish
    value: "sv",
  },
  {
    text: "Română", // Romanian
    value: "ro",
  },
  {
    text: "Türk", // Turkish
    value: "tr",
  },
  {
    text: "عربى", // Arabic
    value: "ar",
    rtl: true,
  },
  {
    text: "کوردی", // Kurdish
    value: "ku",
    rtl: true,
  },
  {
    text: "Беларуская", // Belarusian
    value: "be",
  },
  {
    text: "Български", // Bulgarian
    value: "bg",
  },
  {
    text: "Українська", // Ukrainian
    value: "uk",
  },
  {
    text: "Русский", // Russian
    value: "ru",
  },
  {
    text: "简体中文", // Chinese (Simplified)
    value: "zh",
  },
  {
    text: "繁體中文", // Chinese (Traditional)
    value: "zh_TW",
  },
  {
    text: "日本語", // Japanese
    value: "ja_JP",
  },
  {
    text: "한국어", // Korean
    value: "ko",
  },
  {
    text: "हिन्दी", // Hindi
    value: "hi",
  },
  {
    text: "ภาษาไทย", // Thai
    value: "th",
  },
  {
    text: "فارسی", // Persian
    value: "fa",
    rtl: true,
  },
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

export const MapsStyle = () => [
  {
    text: $gettext("Default"),
    value: "",
  },
  {
    text: $gettext("Low Resolution"),
    value: "offline",
  },
  {
    text: $gettext("Streets"),
    value: "streets",
    sponsor: true,
  },
  {
    text: $gettext("Satellite"),
    value: "hybrid",
    sponsor: true,
  },
  {
    text: $gettext("Outdoor"),
    value: "outdoor",
    sponsor: true,
  },
  {
    text: $gettext("Topographic"),
    value: "topographique",
    sponsor: true,
  },
];

export const PhotoTypes = () => [
  {
    text: $gettext("Image"),
    value: MediaImage,
  },
  {
    text: $gettext("Raw"),
    value: MediaRaw,
  },
  {
    text: $gettext("Animated"),
    value: MediaAnimated,
  },
  {
    text: $gettext("Live"),
    value: MediaLive,
  },
  {
    text: $gettext("Video"),
    value: MediaVideo,
  },
  {
    text: $gettext("Vector"),
    value: MediaVector,
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
  { value: "help", text: $gettext("Customer Support") },
  { value: "feedback", text: $gettext("Product Feedback") },
  { value: "feature", text: $gettext("Feature Request") },
  { value: "bug", text: $gettext("Bug Report") },
  { value: "donations", text: $gettext("Donations") },
  { value: "other", text: $gettext("Other") },
];

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
  { value: 1, text: "" },
  { value: 6, text: "90°" },
  { value: 3, text: "180°" },
  { value: 8, text: "270°" },
];
