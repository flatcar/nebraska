import moment from 'moment';

export function cleanSemverVersion(version) {
  let shortVersion = version
  if (version.includes("+")) {
    shortVersion = version.split("+")[0]
  }
  return shortVersion
}

export function makeLocaleTime(timestamp) {
  let dateFormat = moment.localeData().longDateFormat('L');
  let timeFormat = moment.localeData().longDateFormat('LT');
  return moment.utc(timestamp).local().format(`${dateFormat} ${timeFormat}`);
}
