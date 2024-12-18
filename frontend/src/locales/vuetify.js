import { config } from "app/session";
import { T } from "common/gettext";

export const Locale = () => {
  const locale = config.getLanguage();
  const isRTL = config.rtl();

  return {
    locale: locale,
    fallback: locale,
    rtl: { [locale]: isRTL },
    messages: { [locale]: Messages(T) },
  };
};

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
