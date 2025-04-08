/**
 * Build JSON icon specifications for iconify.
 *
 * This file was taken and adapted from:
 *   https://github.com/iconify/tools/blob/master/sample/parse.js
 */
import fs from 'fs';
import path from 'path';
import { SVG, runSVGO } from '@iconify/tools';

const args = process.argv.slice(2);
const sourceDir = args[0];
const destDir = args[1];

if (!fs.existsSync(sourceDir)) {
  console.error(`Input folder ${sourceDir} does not exist; quitting...`);
  process.exit(1);
}

fs.mkdirSync(destDir, { recursive: true });

// Read directory
const files = await fs.promises.readdir(sourceDir);
console.log(files);

// Filter SVG files
const collection = files
  .filter(file => file.endsWith('.svg'))
  .map(file => {
    const name = path.basename(file, '.svg');
    const content = fs.readFileSync(`${sourceDir}/${file}`, 'utf8');
    return { svg: new SVG(content), name };
  });

console.log(`Found ${collection.length} icons`);

collection.forEach(async ({ svg, name }) => {
  try {
    runSVGO(svg);

    const body = svg.body;
    const width = svg.width;
    const height = svg.height;

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
