export const num = "numeric";
export const short = "short";
export const long = "long";

export const DATE_FULL = {
  year: num,
  month: long,
  day: num,
  weekday: long,
  hour: num,
  minute: num,
};

export const DATE_FULL_TZ = {
  year: num,
  month: short,
  day: num,
  weekday: short,
  hour: num,
  minute: num,
  timeZoneName: short,
};

export const DATE_MED_TZ = {
  year: num,
  month: short,
  day: num,
  hour: num,
  minute: num,
  timeZoneName: short,
};

export const DATE_MED = {
  year: num,
  month: short,
  day: num,
};

export const DATETIME_MED = {
  year: num,
  month: short,
  day: num,
  hour: num,
  minute: num,
};
