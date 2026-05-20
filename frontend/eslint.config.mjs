import { defineConfig, globalIgnores } from "eslint/config";
import globals from "globals";
import path from "node:path";
import { fileURLToPath } from "node:url";
import js from "@eslint/js";
import pluginVue from "eslint-plugin-vue";
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
    // CSS/SCSS/SASS are owned by Prettier (see `frontend/.prettierrc.json` overrides).
    // Ignored here to avoid the "no matching configuration" warning and any
    // accidental formatter jitter from a future ESLint CSS plugin.
    "**/*.css",
    "**/*.scss",
    "**/*.sass",
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
  ...pluginVue.configs["flat/recommended"],
  {
    extends: compat.extends("eslint:recommended", "eslint-config-prettier", "plugin:vuetify/base"),
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
    rules: {
      // Match what Prettier was producing: 2-space indent (4 for CSS lives in .prettierrc), switch
      // cases nested one level, method-chain continuations indented one level.
      "indent": ["warn", 2, { SwitchCase: 1, MemberExpression: 1 }],
      "linebreak-style": ["error", "unix"],
      "quotes": [
        "warn",
        "double",
        {
          avoidEscape: true,
          allowTemplateLiterals: true,
        },
      ],
      "semi": ["warn", "always"],
      "curly": ["warn", "all"],
      // Forces braced bodies onto their own line so curly's autofix produces
      // multi-line `if (x) {\n  return;\n}` instead of `if (x) {return;}`.
      // Deprecated in favor of @stylistic/brace-style; still functional in ESLint 9.
      "brace-style": ["warn", "1tbs", { allowSingleLine: false }],
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
        "warn",
        {
          singleline: "ignore",
          multiline: "ignore",
        },
      ],
      // Note: Prettier is no longer invoked by ESLint. The `eslint-config-prettier` extends
      // above still disables ESLint stylistic rules that would conflict with hand-run Prettier
      // formatting on CSS/SCSS/SASS. Run `prettier --write src/**/*.{css,scss,sass}` (or use
      // the `fmt-css` / `lint-css` npm scripts) to format stylesheets.
    },
  },
]);
