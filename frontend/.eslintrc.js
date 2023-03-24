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
    "prettier",
    "plugin:prettier-vue/recommended",
  ],

  settings: {
    "prettier-vue": {
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
    indent: ["error", 2, { SwitchCase: 1 }],
    "linebreak-style": ["error", "unix"],
    quotes: ["off", "double"], // Easier for Go developers!
    semi: ["error", "always"],
    "no-unused-vars": ["warn"],
    "no-console": 0,
    "no-case-declarations": 0,
    "no-prototype-builtins": 0,
    "vue/no-v-text-v-html-on-component": 0,
    "vue/valid-model-definition": 0,
    "vue/valid-attribute-name": 0,
    "vue/first-attribute-linebreak": [
      "error",
      {
        singleline: "ignore",
        multiline: "ignore",
      },
    ],
    "prettier-vue/prettier": [
      "error",
      {
        // Override all options of `prettier` here
        // @see https://prettier.io/docs/en/options.html
        printWidth: 100,
        singleQuote: false,
        semi: true,
        trailingComma: "es5",
        htmlWhitespaceSensitivity: "strict",
      },
    ],
  },
};
