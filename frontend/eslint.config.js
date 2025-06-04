// For more info, see https://github.com/storybookjs/eslint-plugin-storybook#configuration-flat-config-format
import storybook from "eslint-plugin-storybook";

import js from '@eslint/js';
import globals from 'globals';
import reactHooks from 'eslint-plugin-react-hooks';
import reactRefresh from 'eslint-plugin-react-refresh';
import tseslint from 'typescript-eslint';
import simpleImportSort from 'eslint-plugin-simple-import-sort';
import unusedImports from 'eslint-plugin-unused-imports';
import prettierConfig from "eslint-config-prettier";

export default tseslint.config({ ignores: ['dist', 'node_modules', 'build', 'storybook-static'] }, {
  extends: [js.configs.recommended, ...tseslint.configs.recommended, prettierConfig],
  files: ['**/*.{ts,tsx}'],
  languageOptions: {
    ecmaVersion: 2020,
    globals: globals.browser,
  },
  plugins: {
    'react-hooks': reactHooks,
    'react-refresh': reactRefresh,
    'simple-import-sort': simpleImportSort,
    'unused-imports': unusedImports,
  },
  rules: {
    ...reactHooks.configs.recommended.rules,

    '@typescript-eslint/no-explicit-any': 'off',

    // Import sorting
    'simple-import-sort/imports': [
      'warn',
      {
        groups: [
          ['^\\u0000'], // Side effect imports
          ['^@?\\w'], // Packages
          ['^[^.]'], // Absolute imports
          ['^\\.'], // Relative imports
        ],
      },
    ],

    // Unused imports
    'unused-imports/no-unused-imports': 'error',

    // General rules
    'max-len': [
      'warn',
      {
        code: 100,
        ignoreStrings: true,
        ignoreTemplateLiterals: true,
        ignoreUrls: true,
      },
    ],
    semi: ['warn', 'always'],
    quotes: ['warn', 'single', { avoidEscape: true }],
    eqeqeq: ['warn', 'always'],
    'prefer-const': 'warn',
  },
}, storybook.configs["flat/recommended"]);
