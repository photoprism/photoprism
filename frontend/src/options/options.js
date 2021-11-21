import { timeZonesNames } from "@vvo/tzdb";
import { $gettext } from "common/vm";
import { Info } from "luxon";
import { config } from "../session";
import { TypeVideo, TypeImage, TypeLive, TypeRaw } from "../model/photo";

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

export const Languages = () => [
  {
    text: "English",
    translated: $gettext("English"),
    value: "en",
  },
  {
    text: "Čeština",
    translated: $gettext("Czech"),
    value: "cs",
  },
  {
    text: "Dansk",
    translated: $gettext("Danish"),
    value: "da",
  },
  {
    text: "Deutsch",
    translated: $gettext("German"),
    value: "de",
  },
  {
    text: "Español",
    translated: $gettext("Spanish"),
    value: "es",
  },
  {
    text: "Français",
    translated: $gettext("French"),
    value: "fr",
  },
  {
    text: "עברית",
    translated: $gettext("Hebrew"),
    value: "he",
    rtl: true,
  },
  {
    text: "हिन्दी",
    translated: $gettext("Hindi"),
    value: "hi",
  },
  {
    text: "Magyar",
    translated: $gettext("Hungarian"),
    value: "hu",
  },
  {
    text: "Bahasa Indonesia",
    translated: $gettext("Bahasa Indonesia"),
    value: "id",
  },
  {
    text: "Italian",
    translated: $gettext("Italian"),
    value: "it",
  },
  {
    text: "한국어",
    translated: $gettext("Korean"),
    value: "ko",
  },
  {
    text: "Norsk (Bokmål)",
    translated: $gettext("Norwegian"),
    value: "nb",
  },
  {
    text: "Nederlands",
    translated: $gettext("Dutch"),
    value: "nl",
  },
  {
    text: "Polski",
    translated: $gettext("Polish"),
    value: "pl",
  },
  {
    text: "Português de Portugal",
    translated: $gettext("Português de Portugal"),
    value: "pt",
  },
  {
    text: "Português do Brasil",
    translated: $gettext("Brazilian Portuguese"),
    value: "pt_BR",
  },
  {
    text: "Русский",
    translated: $gettext("Russian"),
    value: "ru",
  },
  {
    text: "Slovenčina",
    translated: $gettext("Slovak"),
    value: "sk",
  },
  {
    text: "简体中文",
    translated: $gettext("Chinese Simplified"),
    value: "zh",
  },
  {
    text: "繁体中文",
    translated: $gettext("Chinese Traditional"),
    value: "zh_TW",
  },
  {
    text: "日本語",
    translated: $gettext("Japanese"),
    value: "ja_JP",
  },
  {
    text: "کوردی",
    translated: $gettext("Kurdish"),
    value: "ku",
    rtl: true,
  },
];

export const Themes = () => [
  {
    text: $gettext("Default"),
    value: "default",
    disabled: false,
  },
  {
    text: $gettext("Grayscale"),
    value: "grayscale",
    disabled: false,
  },
  {
    text: $gettext("Vanta"),
    value: "vanta",
    disabled: false,
  },
  {
    text: $gettext("Moonlight"),
    value: "moonlight",
    disabled: false,
  },
  {
    text: $gettext("Onyx"),
    value: "onyx",
    disabled: false,
  },
  {
    text: $gettext("Cyano"),
    value: "cyano",
    disabled: false,
  },
  {
    text: $gettext("Lavender"),
    value: "lavender",
    disabled: false,
  },
  {
    text: $gettext("Raspberry"),
    value: "raspberry",
    disabled: false,
  },
  {
    text: $gettext("Seaweed"),
    value: "seaweed",
    disabled: false,
  },
  {
    text: $gettext("Shadow"),
    value: "shadow",
    disabled: false,
  },
  {
    text: $gettext("Yellowstone"),
    value: "yellowstone",
    disabled: false,
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
    text: $gettext("Offline"),
    value: "offline",
  },
  {
    text: $gettext("Streets"),
    value: "streets",
  },
  {
    text: $gettext("Hybrid"),
    value: "hybrid",
  },
  {
    text: $gettext("Topographic"),
    value: "topographique",
  },
  {
    text: $gettext("Outdoor"),
    value: "outdoor",
  },
];

export const PhotoTypes = () => [
  {
    text: $gettext("Image"),
    value: TypeImage,
  },
  {
    text: $gettext("Raw"),
    value: TypeRaw,
  },
  {
    text: $gettext("Live"),
    value: TypeLive,
  },
  {
    text: $gettext("Video"),
    value: TypeVideo,
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
