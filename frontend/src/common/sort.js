import { $gettext } from "common/gettext";

// photoOptions returns the sort menu entries for a photo route.
export function photoOptions(route) {
  const items = [
    { value: "newest", text: $gettext("Date Taken"), icon: "mdi-calendar", direction: "date" },
    { value: "title", text: $gettext("Title"), icon: "mdi-format-letter-case", direction: "text" },
    { value: "name", text: $gettext("File Name"), icon: "mdi-file-outline", direction: "text" },
    { value: "size", text: $gettext("File Size"), icon: "mdi-file-chart-outline", direction: "size" },
    { value: "duration", text: $gettext("Video Duration"), icon: "mdi-timer-outline", direction: "duration" },
    { value: "added", text: $gettext("Recently Added"), icon: "mdi-clock-plus-outline" },
    { value: "edited", text: $gettext("Recently Edited"), icon: "mdi-clock-edit-outline" },
  ];
  if (route === "archive") {
    items.push({ value: "archived", text: $gettext("Recently Archived"), icon: "mdi-archive-outline" });
  } else if (route === "browse" || route === "all") {
    items.push(
      { value: "similar", text: $gettext("Visual Similarity"), icon: "mdi-image-search-outline" },
      { value: "relevance", text: $gettext("Most Relevant"), icon: "mdi-bullseye-arrow" }
    );
  }
  return items;
}

// albumOptions returns the sort menu entries for an album type.
export function albumOptions(type) {
  const items = [
    { value: "name", text: $gettext("Name"), icon: "mdi-format-letter-case", direction: "text" },
    { value: "added", text: $gettext("Date Created"), icon: "mdi-calendar-plus", direction: "date" },
  ];
  if (type === "moment" || type === "month") {
    items.push({ value: "newest", text: type === "moment" ? $gettext("Year") : $gettext("Date"), icon: "mdi-calendar", direction: "date" });
  }
  items.push({ value: "place", text: $gettext("Location"), icon: "mdi-map-marker-outline" });
  if (type === "album") {
    items.push({ value: "edited", text: $gettext("Recently Edited"), icon: "mdi-clock-edit-outline" });
  }
  items.push({ value: "favorites", text: $gettext("Favorites First"), icon: "mdi-star-outline" });
  return items;
}

// normalizeSort maps stored and URL sort values to an available menu option.
export function normalizeSort(order, reverse, items, defaultOrder) {
  reverse = reverse === true || reverse === "true";
  if ((order === "newest" || order === "oldest") && !items.some((item) => item.value === "newest")) {
    reverse = (order === "oldest") !== reverse;
    order = "added";
  } else if (order === "oldest" || order === "newest") {
    order = (order === "oldest") !== reverse ? "oldest" : "newest";
    reverse = false;
  }
  const item = items.find((entry) => entry.value === (order === "oldest" ? "newest" : order));
  if (!item) {
    return { order: defaultOrder, reverse: false };
  }
  return { order, reverse: !!item.direction && reverse };
}

// selectSort applies a row toggle or an explicit ascending/descending direction.
export function selectSort(current, item, ascending) {
  const active = current.order === item.value || (item.value === "newest" && current.order === "oldest");
  let reverse = false;
  if (item.direction) {
    const defaultAscending = item.direction === "text";
    reverse = ascending === undefined ? active && !(current.order === "oldest" || current.reverse) : ascending !== defaultAscending;
  }
  return item.value === "newest" ? { order: reverse ? "oldest" : "newest", reverse: false } : { order: item.value, reverse };
}

// sortAscending reports the visual direction of a selected sort option.
export function sortAscending(state, item) {
  const reversed = state.order === "oldest" || state.reverse;
  return item.direction === "text" ? !reversed : !!reversed;
}

// sortLabel returns a complete translated description of the active sort.
export function sortLabel(state, item) {
  if (!item) {
    return $gettext("Sort By");
  }
  if (!item.direction) {
    return $gettext("Sorted by %{name}", { name: item.text });
  }
  return sortAscending(state, item)
    ? $gettext("Sorted by %{name}, ascending", { name: item.text })
    : $gettext("Sorted by %{name}, descending", { name: item.text });
}

// photoSortKeys preserves existing preference keys and isolates other photo routes.
export function photoSortKeys(route) {
  if (["archive", "favorites", "hidden", "review"].includes(route)) {
    return { order: `${route}.order`, reverse: `${route}.reverse` };
  }
  if (route === "browse") {
    return { order: "photos.order", reverse: "photos.reverse" };
  }
  return { order: `photos.order.${route}`, reverse: `photos.reverse.${route}`, legacy: "photos.order" };
}

// resolveSort restores and remembers a normalized per-page sort preference.
export function resolveSort(query, storage, keys, items, defaultOrder) {
  const stored = storage.getItem(keys.order);
  const order = query.order || stored || (keys.legacy && storage.getItem(keys.legacy)) || defaultOrder;
  const reverse = query.order ? query.reverse : query.reverse ?? (stored ? storage.getItem(keys.reverse) : false);
  const state = normalizeSort(order, reverse, items, defaultOrder);
  storage.setItem(keys.order, state.order);
  storage.setItem(keys.reverse, state.reverse);
  return state;
}

// sortParams omits reverse unless the selected option needs it.
export function sortParams(params) {
  const result = { ...params };
  if (!result.reverse || result.order === "newest" || result.order === "oldest") {
    delete result.reverse;
  }
  return result;
}
