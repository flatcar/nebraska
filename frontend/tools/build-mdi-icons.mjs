/**
 * Generate local ESM-friendly JSON icon data for the Material Design Icons
 * used by the app, sourced from the maintained `@iconify-json/mdi` package.
 *
 * We vendor the handful of icons we use as plain JSON (same convention as the
 * other files in `src/icons/`). JSON imports are native ESM in the bundler, so
 * this avoids relying on the deprecated, CJS-only `@iconify/icons-mdi` package
 * and its `__esModule`/default-export interop.
 *
 * NOTE: `src/icons/mdi/` is exclusively owned by this generator. It is fully
 * regenerated from `ICON_NAMES` on every run: any `*.json` no longer listed is
 * pruned. Do not place hand-authored files there, as they would be removed.
 *
 * Regenerate with: `npm run build-mdi-icons`
 */
import fs from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

import { getIconData } from '@iconify/utils';
import { icons } from '@iconify-json/mdi';

// Icons referenced across the app (mdi prefix). Keep sorted.
const ICON_NAMES = [
  'alert-circle-outline',
  'alert-outline',
  'cancel',
  'check-circle-outline',
  'chevron-down',
  'chevron-up',
  'cube-outline',
  'download-circle-outline',
  'help-circle-outline',
  'information-circle-outline',
  'menu-down',
  'menu-swap',
  'menu-up',
  'package-variant-closed',
  'pause-circle',
  'play-circle',
  'plus',
  'progress-download',
  'search',
  'square',
];

const outDir = path.join(path.dirname(fileURLToPath(import.meta.url)), '..', 'src', 'icons', 'mdi');
fs.mkdirSync(outDir, { recursive: true });

// Prune stale icons no longer listed in ICON_NAMES so the directory always
// reflects the current set (see the note at the top of this file).
const keep = new Set(ICON_NAMES.map(name => `${name}.json`));
for (const file of fs.readdirSync(outDir)) {
  if (file.endsWith('.json') && !keep.has(file)) {
    fs.rmSync(path.join(outDir, file));
    console.log(`removed stale src/icons/mdi/${file}`);
  }
}

for (const name of ICON_NAMES) {
  const data = getIconData(icons, name);
  if (!data) {
    console.error(`Icon "mdi:${name}" not found in @iconify-json/mdi`);
    process.exit(1);
  }
  const { body, width, height } = data;
  const icon = { body, width, height };
  fs.writeFileSync(path.join(outDir, `${name}.json`), `${JSON.stringify(icon, null, 2)}\n`);
  console.log(`generated src/icons/mdi/${name}.json`);
}
