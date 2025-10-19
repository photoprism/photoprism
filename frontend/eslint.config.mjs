import { defineConfig, globalIgnores } from "eslint/config";
import globals from "globals";
import path from "node:path";
import { fileURLToPath } from "node:url";
import js from "@eslint/js";
import { FlatCompat } from "@eslint/eslintrc";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const compat = new FlatCompat({
  baseDirectory: __dirname,
  recommendedConfig: js.configs.recommended,
  allConfig: js.configs.all,
});

export default defineConfig([
  globalIgnores([
    "**/coverage/",
    "**/node_modules/",
    "tests/screenshots/",
    "tests/acceptance/screenshots/",
    "tests/upload-files/",
    "**/*.html",
    "**/.idea",
    "**/.codex",
    "**/.env",
    "**/.venv",
    "**/.github",
    "**/.tmp",
    "**/.local",
    "**/.cache",
    "**/.gocache",
    "**/.var",
  ]),
  {
    extends: compat.extends(
      "eslint:recommended",
      "plugin:vue/recommended",
      "plugin:prettier/recommended",
      "plugin:vue/base",
      "plugin:vuetify/base"
    ),
    languageOptions: {
      globals: {
        ...globals.browser,
        ...globals.commonjs,
        ...globals.node,
        ...globals.mocha,
      },

      ecmaVersion: "latest",
      sourceType: "module",
    },
    settings: {
      "prettier/prettier": {
        // Settings for how to process Vue SFC Blocks
        SFCBlocks: {
          template: false,
          script: false,
          style: false,
        },

        // Use prettierrc for prettier options or not (default: `true`)
        usePrettierrc: true,

        // Set the options for `prettier.getFileInfo`.
        // @see https://prettier.io/docs/en/api.html#prettiergetfileinfofilepath-options
        fileInfoOptions: {
          // Path to ignore file (default: `'.prettierignore'`)
          // Notice that the ignore file is only used for this plugin
          ignorePath: ".testignore",

          // Process the files in `node_modules` or not (default: `false`)
          withNodeModules: false,
        },
      },
    },
    rules: {
      "indent": ["error", 2, { SwitchCase: 1 }],
      "linebreak-style": ["error", "unix"],
      "quotes": [
        "off",
        "double",
        {
          avoidEscape: true,
          allowTemplateLiterals: true,
        },
      ],
      "semi": ["error", "always"],
      "no-unused-vars": ["warn"],
      "no-console": 0,
      "no-case-declarations": 0,
      "no-prototype-builtins": 0,
      "vue/no-v-text-v-html-on-component": 0,
      "vue/no-v-model-argument": 0,
      "vue/valid-model-definition": 0,
      "vue/valid-attribute-name": 0,
      "vue/singleline-html-element-content-newline": [
        "off",
        {
          ignoreWhenNoAttributes: true,
          ignoreWhenEmpty: true,
          ignores: [
            "pre",
            "textarea",
            "span",
            "translate",
            "a",
            "v-icon",
            "v-text-field",
            "v-input",
            "v-select",
            "v-switch",
            "v-checkbox",
            "v-img",
          ],
          externalIgnores: [],
        },
      ],
      "vue/first-attribute-linebreak": [
        "error",
        {
          singleline: "ignore",
          multiline: "ignore",
        },
      ],
      "no-restricted-syntax": [
        "error",
        {
          selector: "CallExpression[callee.object.name='window'][callee.property.name='open'][arguments.length<3]",
          message: "window.open must include a features string with 'noopener,noreferrer'.",
        },
        {
          selector:
            "CallExpression[callee.object.name='window'][callee.property.name='open']:not([arguments.2.value=/noopener/])",
          message: "window.open features must include 'noopener'.",
        },
        {
          selector:
            "CallExpression[callee.object.name='window'][callee.property.name='open']:not([arguments.2.value=/noreferrer/])",
          message: "window.open features must include 'noreferrer'.",
        },
      ],
      "prettier/prettier": [
        "warn",
        {
          printWidth: 120,
          semi: true,
          singleQuote: false,
          bracketSpacing: true,
          trailingComma: "es5",
          htmlWhitespaceSensitivity: "css",
          quoteProps: "consistent",
          proseWrap: "never",
        },
      ],
    },
  },
]);
