export const UnknownTimelineKey = "unknown";

function positiveInt(value, min, max) {
  const n = Number.parseInt(value, 10);

  if (!Number.isInteger(n) || n < min || n > max) {
    return 0;
  }

  return n;
}

function localPart(value, start, length, min, max) {
  if (typeof value !== "string" || value.length < start + length) {
    return 0;
  }

  return positiveInt(value.substring(start, start + length), min, max);
}

export function photoLocalDateParts(photo) {
  const takenAtLocal = photo?.TakenAtLocal || "";
  const year = localPart(takenAtLocal, 0, 4, 1000, 9999) || positiveInt(photo?.Year, 1000, 9999);
  const month = localPart(takenAtLocal, 5, 2, 1, 12) || positiveInt(photo?.Month, 1, 12);
  const day = localPart(takenAtLocal, 8, 2, 1, 31) || positiveInt(photo?.Day, 1, 31);

  if (!year || !month) {
    return {
      known: false,
      year: 0,
      month: 0,
      day: 0,
    };
  }

  return {
    known: true,
    year,
    month,
    day,
  };
}

function monthLabel(month, monthOptions) {
  const option = monthOptions.find((item) => item.value === month);
  return option ? option.text : month.toString().padStart(2, "0");
}

function countLabel(count, gettext) {
  return count === 1 ? gettext("1 picture") : `${count} ${gettext("pictures")}`;
}

export function buildTimelineSections(photos, monthOptions, gettext = (s) => s) {
  const sections = [];
  const sectionMap = new Map();
  const unknownSection = {
    key: UnknownTimelineKey,
    title: gettext("Unknown date"),
    count: 0,
    countLabel: "",
    days: [
      {
        key: UnknownTimelineKey,
        title: gettext("Unknown"),
        entries: [],
      },
    ],
  };

  for (let index = 0; index < photos.length; index++) {
    const photo = photos[index];
    const parts = photoLocalDateParts(photo);
    const entry = { photo, index };

    if (!parts.known) {
      unknownSection.days[0].entries.push(entry);
      unknownSection.count++;
      continue;
    }

    const sectionKey = `${parts.year}-${parts.month.toString().padStart(2, "0")}`;
    let section = sectionMap.get(sectionKey);

    if (!section) {
      section = {
        key: sectionKey,
        title: `${monthLabel(parts.month, monthOptions)} ${parts.year}`,
        count: 0,
        countLabel: "",
        days: [],
        dayMap: new Map(),
      };
      sectionMap.set(sectionKey, section);
      sections.push(section);
    }

    const dayKey = parts.day ? `${sectionKey}-${parts.day.toString().padStart(2, "0")}` : `${sectionKey}-unknown`;
    let day = section.dayMap.get(dayKey);

    if (!day) {
      day = {
        key: dayKey,
        title: parts.day ? parts.day.toString() : gettext("Unknown"),
        entries: [],
      };
      section.dayMap.set(dayKey, day);
      section.days.push(day);
    }

    day.entries.push(entry);
    section.count++;
  }

  if (unknownSection.count > 0) {
    sections.push(unknownSection);
  }

  for (const section of sections) {
    section.countLabel = countLabel(section.count, gettext);
    delete section.dayMap;
  }

  return sections;
}
