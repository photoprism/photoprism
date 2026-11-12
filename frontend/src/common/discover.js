import { DateTime } from "luxon";

export const RandomCount = 36;

// viewerNow returns the current instant in the user's Settings time zone.
export function viewerNow(timeZone, now = new Date()) {
  const zone = !timeZone || timeZone === "Local" ? undefined : timeZone;
  const dt = zone ? DateTime.fromJSDate(now, { zone }) : DateTime.fromJSDate(now);

  if (!dt.isValid) {
    return DateTime.fromJSDate(now);
  }

  return dt;
}

// viewerToday returns the calendar year, month, and day in the viewer's zone.
export function viewerToday(timeZone, now = new Date()) {
  const dt = viewerNow(timeZone, now);

  return { year: dt.year, month: dt.month, day: dt.day };
}

// pastYears returns indexed years strictly before currentYear, newest first.
export function pastYears(indexedYears, currentYear) {
  if (!Array.isArray(indexedYears) || indexedYears.length === 0) {
    return [];
  }

  return indexedYears
    .map((year) => parseInt(year, 10))
    .filter((year) => Number.isInteger(year) && year > 0 && year < currentYear)
    .sort((a, b) => b - a);
}

// isPastYearPhoto reports whether a search result belongs to a previous year.
export function isPastYearPhoto(photo, currentYear) {
  const year = photo?.Year ?? photo?.PhotoYear;
  return Number.isInteger(year) && year > 0 && year < currentYear;
}

// startOfCurrentYear returns the RFC 3339 date passed to the before filter.
// The comparison is midnight, so the previous year's last day stays included.
export function startOfCurrentYear(currentYear) {
  return `${currentYear}-01-01`;
}

// pageAdvance reports whether a photos page is the last and where the next page starts.
export function pageAdvance(response, requestedOffset = 0, requestedCount = 0) {
  const limit = Number.isFinite(response?.limit) ? response.limit : 0;
  const offset = Number.isFinite(response?.offset) ? response.offset : requestedOffset;
  const count = Number.isFinite(response?.count) ? response.count : 0;
  const complete = limit <= 0 || count < limit;

  return {
    complete,
    offset: complete ? offset : offset + (limit > 0 ? limit : requestedCount),
  };
}

// discoverSearchParams builds GET /api/v1/photos query fields for a Discover day or month tab.
export function discoverSearchParams({
  mode = "day",
  timeZone = "Local",
  years = [],
  now = new Date(),
  count = 156,
  offset = 0,
  quality,
  public: isPublic,
} = {}) {
  const today = viewerToday(timeZone, now);
  const past = pastYears(years, today.year);
  const params = {
    count,
    offset,
    merged: true,
    order: "newest",
    month: String(today.month),
  };

  if (mode === "day") {
    params.day = String(today.day);
  }

  if (past.length) {
    params.year = past.join("|");
  } else {
    params.before = startOfCurrentYear(today.year);
  }

  if (quality) {
    params.quality = quality;
  }

  if (isPublic) {
    params.public = "true";
  }

  return params;
}

// randomSearchParams builds GET /api/v1/photos query fields for a finite random set.
export function randomSearchParams({ count = RandomCount, quality, public: isPublic } = {}) {
  const params = {
    count,
    offset: 0,
    merged: true,
    order: "random",
  };

  if (quality) {
    params.quality = quality;
  }

  if (isPublic) {
    params.public = "true";
  }

  return params;
}

// groupPhotosByYear groups past-year photos newest year first and skips empty years.
export function groupPhotosByYear(photos, currentYear) {
  const byYear = new Map();

  for (const photo of photos || []) {
    if (!isPastYearPhoto(photo, currentYear)) {
      continue;
    }

    const year = photo.Year ?? photo.PhotoYear;

    if (!byYear.has(year)) {
      byYear.set(year, []);
    }

    byYear.get(year).push(photo);
  }

  return [...byYear.keys()]
    .sort((a, b) => b - a)
    .map((year) => {
      const items = byYear.get(year);
      return {
        year,
        yearsAgo: currentYear - year,
        count: items.length,
        photos: items,
      };
    });
}
