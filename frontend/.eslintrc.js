module.exports = {
  env: {
    browser: true,
    commonjs: true,
    es2021: true,
    node: true,
    mocha: true,
  },
  extends: [
    "eslint:recommended",
    "plugin:vue/recommended",
    "plugin:prettier/recommended",
    "plugin:vue/base",
    "plugin:vuetify/base",
  ],
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
  parserOptions: {
    ecmaVersion: "latest",
    sourceType: "module",
  },
  rules: {
    // 'comma-dangle': ['error', 'always-multiline'],
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
};
