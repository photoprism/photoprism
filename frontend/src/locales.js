import { $config } from "app/session";
import { $gettext, T } from "common/gettext";

// Returns the id and messages of the current locale.
export const Locale = () => {
  const locale = $config.getLanguageLocale();
  const isRTL = $config.isRtl();

  return {
    locale: locale,
    fallback: locale,
    rtl: { [locale]: isRTL },
    messages: { [locale]: Messages(T) },
  };
};

// Contains the supported locales, their names and properties.
export let Options = [
  {
    text: "English", // English
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
    text: "Euskara", // Basque
    value: "eu",
  },
  {
    text: "Français", // French
    value: "fr",
  },
  {
    text: "Gaeilge", // Irish
    value: "ga",
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
    text: "Türkçe", // Turkish
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
    value: "ja",
  },
  {
    text: "한국어", // Korean
    value: "ko",
  },
  {
    text: "Tiếng Việt", // Vietnamese
    value: "vi",
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

// Returns the Vuetify UI messages translated with Gettext.
export const Messages = ($gettext) => {
  return {
    badge: $gettext("Badge"),
    open: $gettext("Open"),
    close: $gettext("Close"),
    dismiss: $gettext("Dismiss"),
    confirmEdit: {
      ok: $gettext("OK"),
      cancel: $gettext("Cancel"),
    },
    dataIterator: {
      noResultsText: $gettext("No matching records found"),
      loadingText: $gettext("Loading items..."),
    },
    dataTable: {
      itemsPerPageText: $gettext("Rows per page:"),
      itemsPerPageAll: $gettext("All"),
      ariaLabel: {
        sortDescending: $gettext("Sorted descending."),
        sortAscending: $gettext("Sorted ascending."),
        sortNone: $gettext("Not sorted."),
        activateNone: $gettext("Activate to remove sorting."),
        activateDescending: $gettext("Activate to sort descending."),
        activateAscending: $gettext("Activate to sort ascending."),
      },
      sortBy: $gettext("Sort by"),
    },
    dataFooter: {
      itemsPerPageText: $gettext("Items per page:"),
      itemsPerPageAll: $gettext("All"),
      nextPage: $gettext("Next page"),
      prevPage: $gettext("Previous page"),
      firstPage: $gettext("First page"),
      lastPage: $gettext("Last page"),
      pageText: $gettext("{0}-{1} of {2}"),
    },
    dateRangeInput: {
      divider: $gettext("to"),
    },
    datePicker: {
      itemsSelected: $gettext("{0} selected"),
      range: {
        title: $gettext("Select dates"),
        header: $gettext("Enter dates"),
      },
      title: $gettext("Select date"),
      header: $gettext("Enter date"),
      input: {
        placeholder: $gettext("Enter date"),
      },
    },
    noDataText: $gettext("No data available"),
    carousel: {
      prev: $gettext("Previous visual"),
      next: $gettext("Next visual"),
      ariaLabel: {
        delimiter: $gettext("Carousel slide {0} of {1}"),
      },
    },
    calendar: {
      moreEvents: $gettext("{0} more"),
      today: $gettext("Today"),
    },
    input: {
      clear: $gettext("Clear {0}"),
      prependAction: $gettext("{0} prepended action"),
      appendAction: $gettext("{0} appended action"),
      otp: $gettext("Please enter OTP character {0}"),
    },
    fileInput: {
      counter: $gettext("{0} files"),
      counterSize: $gettext("{0} files ({1} in total)"),
    },
    fileUpload: {
      title: $gettext("Drag and drop files here"),
      divider: $gettext("or"),
      browse: $gettext("Browse Files"),
    },
    timePicker: {
      am: $gettext("AM"),
      pm: $gettext("PM"),
      title: $gettext("Select Time"),
    },
    pagination: {
      ariaLabel: {
        root: $gettext("Pagination Navigation"),
        next: $gettext("Next page"),
        previous: $gettext("Previous page"),
        page: $gettext("Go to page {0}"),
        currentPage: $gettext("Page {0}, Current page"),
        first: $gettext("First page"),
        last: $gettext("Last page"),
      },
    },
    stepper: {
      next: $gettext("Next"),
      prev: $gettext("Previous"),
    },
    rating: {
      ariaLabel: {
        item: $gettext("Rating {0} of {1}"),
      },
    },
    loading: $gettext("Loading..."),
    infiniteScroll: {
      loadMore: $gettext("Load more"),
      empty: $gettext("No more"),
    },
  };
};

// Extra UI translation messages.
export const ExtraMessages = () => {
  $gettext("Search");
  $gettext("Refresh");
  $gettext("Delete");
  $gettext("Open");
  $gettext("Name");
  $gettext("Username");
  $gettext("Display Name");
  $gettext("Version");
  $gettext("Portal");
  $gettext("Theme");
  $gettext("Labels");
  $gettext("Removed");
  $gettext("Database");
  $gettext("Databases");
  $gettext("User");
  $gettext("Users");
  $gettext("Account");
  $gettext("Accounts");
  $gettext("Authentication");
  $gettext("Web Login");
  $gettext("Last Login");
  $gettext("Role");
  $gettext("Roles");
  $gettext("Attributes");
  $gettext("Scope");
  $gettext("Scopes");
  $gettext("Local");
  $gettext("Session");
  $gettext("Sessions");
  $gettext("Driver");
  $gettext("Engine");
  $gettext("Rotated");
  $gettext("Severity");
  $gettext("Activity");
  $gettext("Time");
  $gettext("IP Address");
  $gettext("Site URL");
  $gettext("Message");
  $gettext("Repeated");
  $gettext("Application");
  $gettext("Applications");
  $gettext("Node");
  $gettext("Nodes");
  $gettext("Service");
  $gettext("Services");
  $gettext("Instance");
  $gettext("Instances");
  $gettext("Remove the selected instance from the cluster registry?");
};
