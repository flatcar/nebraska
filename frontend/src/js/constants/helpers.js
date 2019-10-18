import moment from 'moment';

export function cleanSemverVersion(version) {
  let shortVersion = version
  if (version.includes("+")) {
    shortVersion = version.split("+")[0]
  }
  return shortVersion
}

export function makeLocaleTime(timestamp, formats={}) {
  const {dateFormat='L', timeFormat='LT'} = formats;
  let localeDateFormat = dateFormat ? dateFormat : '';
  let localeTimeFormat = timeFormat ? timeFormat : '';
  let format = localeDateFormat;

  if (format != '' && localeTimeFormat)
    format += ' ';
  format += localeTimeFormat;

  return moment.utc(timestamp).local().format(format, moment.locale());
}
