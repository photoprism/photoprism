// todayMemoriesFilter returns search filters for photos captured on today's calendar date.
export function todayMemoriesFilter(date = new Date()) {
  return {
    photo: "true",
    month: `${date.getMonth() + 1}`,
    day: `${date.getDate()}`,
    before: `${date.getFullYear() - 1}-12-31`,
    order: "newest",
  };
}
