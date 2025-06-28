// https://storybook.js.org/docs/react/configure/theming#create-a-theme-quickstart
//  To workaround a bug at time of writing, where theme is not refreshed,
//  you may need to `npm run storybook --no-manager-cache`
import { create } from '@storybook/theming';
import logoUrl from '../../docs/nebraska-logo.svg';

export default create({
  base: 'light',
  brandTitle: 'Nebraska is an update manager for Flatcar Container Linux and Kubernetes.',
  brandUrl: 'https://flatcar.org/docs/latest/nebraska/development/',
  brandImage: logoUrl,
});
