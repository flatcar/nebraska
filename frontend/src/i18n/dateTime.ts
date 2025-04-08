type DateParam = string | number | Date;

export function toLocaleString(
  date: DateParam,
  locales?: string | string[],
  options?: Intl.DateTimeFormatOptions
) {
  // Locales and timezones can be different, so we set a default when under test.
  if (import.meta.env.TEST_TZ) {
    const newOptions = options ? options : {};
    return new Date(date).toLocaleString('en-US', { ...newOptions, timeZone: 'UTC' });
  } else {
    return new Date(date).toLocaleString(locales, options);
  }
}

export function toLocaleDateString(
  date: DateParam,
  locales?: string | string[],
  options?: Intl.DateTimeFormatOptions
) {
  if (import.meta.env.TEST_TZ) {
    // Locales and timezones can be different, so we set a default when under test.
    const newOptions = options ? options : {};
    return new Date(date).toLocaleDateString('en-US', { ...newOptions, timeZone: 'UTC' });
  } else {
    return new Date(date).toLocaleDateString(locales, options);
  }
}

export function getMinuteDifference(date1: number, date2: number) {
  return (date1 - date2) / 1000 / 60;
}

export function makeLocaleTime(
  timestamp: string | Date | number,
  opts: { useDate?: boolean; showTime?: boolean; dateFormat?: Intl.DateTimeFormatOptions } = {}
) {
  const { useDate = true, showTime = true, dateFormat = {} } = opts;
  const date = new Date(timestamp);
  const formattedDate = toLocaleDateString(date, undefined, dateFormat);
  const timeFormat = toLocaleString(date, undefined, { hour: '2-digit', minute: '2-digit' });

  if (useDate && showTime) {
    return `${formattedDate} ${timeFormat}`;
  }
  if (useDate) {
    return formattedDate;
  }
  return timeFormat;
}
