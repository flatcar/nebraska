/**
 * Build JSON icon specifications for iconify.
 *
 * This file was taken and adapted from:
 *   https://github.com/iconify/tools/blob/master/sample/parse.js
 */
import fs from 'fs';
import path from 'path';
import {
  importDirectory,
  parseSVG,
  runSVGO,
} from '@iconify/tools';


const args = process.argv.slice(2);
const sourceDir = args[0];
const destDir = args[1];

if (!fs.existsSync(sourceDir)) {
  console.error(`Input folder ${sourceDir} does not exist; quitting...`);
  process.exit(1);
}

fs.mkdirSync(destDir, { recursive: true });

const collection = await importDirectory(sourceDir, {
  prefix: '', // Match original behavior: no prefix
  includeSubDirs: true,
});

console.log(`Found ${collection.size} icons`);

await collection.forEach(async (svg, name) => {
  try {
    // Manually parse the SVG if needed
    const parsedSVG = parseSVG(svg);
    await runSVGO(parsedSVG);

    // Proceed with other operations
    const body = parsedSVG.body.replace(/\s*\n\s*/g, '');
    const width = parsedSVG.width;
    const height = parsedSVG.height;

    const json = {
      body,
      width,
      height,
    };

    const targetPath = path.join(destDir, `${name}.json`);
    fs.writeFileSync(targetPath, JSON.stringify(json, null, 2), 'utf8');
    console.log(`Created ${targetPath}`);
  } catch (err) {
    console.error(`Error processing icon "${name}":`, err);
  }
});
