/**
 * Build JSON icon specifications for iconify.
 *
 * This file was taken and adapted from:
 *   https://github.com/iconify/tools/blob/master/sample/parse.js
 *
 * @todo: Make this a webpack loader.
 */
'use strict';

const fs = require('fs');
const path = require('path');
const tools = require('@iconify/tools');

let collection;
let args = process.argv.slice(2);
let sourceDir = args[0];
let destDir = args[1];

if (!fs.existsSync(sourceDir)) {
  console.log(`Input folder ${sourceDir} does not exist; quitting...`);
  process.exit(1);
}

// Create directories
try {
  fs.mkdirSync(destDir);
} catch (err) {
  if (err.code !== 'EEXIST') {
    console.log(err);
  }
}

// Do stuff
tools
  .ImportDir(sourceDir)
  .then(result => {
    collection = result;
    console.log('Found ' + collection.length() + ' icons.');

    // SVGO optimization
    return collection.promiseAll(svg => tools.SVGO(svg));
  })
  .then(() => {
    // Clean up tags
    return collection.promiseAll(svg => tools.Tags(svg));
  })
  .then(() => {
    // SVGO optimization again. Might make files smaller after color/tags changes
    return collection.promiseAll(svg => tools.SVGO(svg));
  })
  .then(() => {
    // Export each icon as JSON
    collection.forEach((item, key) => {
      let json = {
        body: item.getBody().replace(/\s*\n\s*/g, ''),
        width: item.width,
        height: item.height,
      };
      let content = JSON.stringify(json, null, '\t');
      let target = path.join(destDir, `${key}.json`);

      try {
        try {
          fs.mkdirSync(destDir);
        } catch (err) {
          if (err.code !== 'EEXIST') {
            console.log(err);
            process.exit(1);
          }
        }

        fs.writeFileSync(target, content, 'utf8');
        console.log(`Created ${target}`);
      } catch (err) {
        console.log(err);
        process.exit(1);
      }
    });
  })
  .catch(err => {
    console.log(err);
  });
