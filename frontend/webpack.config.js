/*

Copyright (c) 2018 - 2025 PhotoPrism UG. All rights reserved.

    This program is free software: you can redistribute it and/or modify
    it under Version 3 of the GNU Affero General Public License (the "AGPL"):
    <https://docs.photoprism.app/license/agpl>

    This program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU Affero General Public License for more details.

    The AGPL is supplemented by our Trademark and Brand Guidelines,
    which describe how our Brand Assets may be used:
    <https://www.photoprism.app/trademark>

Feel free to send an email to hello@photoprism.app if you have questions,
want to support our work, or just want to say hello.

Additional information can be found in our Developer Guide:
<https://docs.photoprism.app/developer-guide/>

*/

const path = require("path");
const MiniCssExtractPlugin = require("mini-css-extract-plugin");
const ESLintPlugin = require("eslint-webpack-plugin");
const { WebpackManifestPlugin } = require("webpack-manifest-plugin");
const OfflinePlugin = require("@lcdp/offline-plugin");
const webpack = require("webpack");
const isAnalyze = process.env?.BUILD_ENV === "analyze" || process.env?.NODE_ENV === "analyze";
const isDev = isAnalyze || process.env?.BUILD_ENV === "development" || process.env?.NODE_ENV === "development";
const isCustom = !!process.env.CUSTOM_SRC;
const appName = process.env.CUSTOM_NAME ? process.env.CUSTOM_NAME : "PhotoPrism";
const { VueLoaderPlugin } = require("vue-loader");
const { VuetifyPlugin } = require("webpack-plugin-vuetify");
const { DefinePlugin } = require("webpack");
const BundleAnalyzerPlugin = require("webpack-bundle-analyzer").BundleAnalyzerPlugin;

const PATHS = {
  src: path.join(__dirname, "src"),
  css: path.join(__dirname, "src/css"),
  modules: path.join(__dirname, "node_modules"),
  app: path.join(__dirname, "src/app.js"),
  share: path.join(__dirname, "src/share.js"),
  splash: path.join(__dirname, "src/splash.js"),
  build: path.join(__dirname, "../assets/static/build"),
  public: "./",
};

if (isCustom) {
  PATHS.custom = path.join(__dirname, process.env.CUSTOM_SRC);
}

if (isDev) {
  console.log(`Starting ${appName} DEVELOPMENT build. Please wait.`);
} else {
  console.log(`Starting ${appName} PRODUCTION build. Please wait.`);
}

const config = {
  mode: isDev ? "development" : "production",
  devtool: isDev ? "inline-source-map" : false,
  optimization: {
    minimize: !isDev,
  },
  entry: {
    app: PATHS.app,
    share: PATHS.share,
    splash: PATHS.splash,
  },
  output: {
    path: PATHS.build,
    publicPath: PATHS.public,
    filename: "[name].[contenthash].js",
    clean: true,
  },
  resolve: {
    modules: isCustom ? [PATHS.custom, PATHS.src, PATHS.modules] : [PATHS.src, PATHS.modules],
    preferRelative: true,
    alias: {
      // TODO: change it
      vue$: "vue/dist/vue.runtime.esm-bundler.js",
      // vue: isDev ? "vue/dist/vue.js" : "vue/dist/vue.min.js",
    },
  },
  plugins: [
    new MiniCssExtractPlugin({
      filename: "[name].[contenthash].css",
      experimentalUseImportModule: false,
    }),
    new WebpackManifestPlugin({
      fileName: "assets.json",
      publicPath: "",
    }),
    new webpack.ProgressPlugin(),
    new VueLoaderPlugin(),
    new OfflinePlugin({
      relativePaths: false,
      publicPath: "/",
      excludes: ["**/*.txt", "**/*.css", "**/*.js", "**/*.*"],
      rewrites: function (asset) {
        return "/static/build/" + asset;
      },
    }),
    new VuetifyPlugin({ autoImport: true }),
    new DefinePlugin({
      __VUE_OPTIONS_API__: JSON.stringify(true), // Change to false as needed
      __VUE_PROD_DEVTOOLS__: JSON.stringify(false), // Change to true to enable in production
      __VUE_PROD_HYDRATION_MISMATCH_DETAILS__: JSON.stringify(false), // Change to true for detailed warnings
    }),
  ],
  performance: {
    hints: isDev ? false : "warning",
    maxEntrypointSize: 7500000,
    maxAssetSize: 7500000,
  },
  module: {
    rules: [
      {
        test: /\.vue$/,
        loader: "vue-loader",
        options: {
          loaders: {
            js: "babel-loader",
            css: "css-loader",
          },
          compilerOptions: {
            whitespace: "preserve",
          },
        },
      },
      {
        test: /\.js$/,
        include: PATHS.src,
        exclude: (file) => /node_modules/.test(file),
        use: [
          {
            loader: "babel-loader",
            options: {
              sourceMap: isDev,
              compact: false,
              presets: ["@babel/preset-env"],
              plugins: [],
            },
          },
        ],
      },
      {
        test: /\.css$/,
        include: isCustom ? [PATHS.custom, PATHS.css] : [PATHS.css],
        exclude: /node_modules/,
        use: [
          {
            loader: MiniCssExtractPlugin.loader,
            options: {
              publicPath: PATHS.public,
            },
          },
          {
            loader: "css-loader",
            options: {
              sourceMap: true,
              importLoaders: 1,
            },
          },
          "resolve-url-loader",
          {
            loader: "postcss-loader",
            options: {
              sourceMap: true,
              postcssOptions: {
                config: path.resolve(__dirname, "./postcss.config.js"),
              },
            },
          },
        ],
      },
      {
        test: /\.css$/,
        include: /node_modules/,
        use: [
          {
            loader: MiniCssExtractPlugin.loader,
            options: {
              publicPath: PATHS.public,
            },
          },
          {
            loader: "css-loader",
            options: {
              sourceMap: true,
              importLoaders: 1,
            },
          },
          "resolve-url-loader",
          {
            loader: "postcss-loader",
            options: {
              sourceMap: true,
              postcssOptions: {
                config: path.resolve(__dirname, "./postcss.config.js"),
              },
            },
          },
        ],
      },
      {
        test: /\.s[c|a]ss$/,
        use: [
          {
            loader: MiniCssExtractPlugin.loader,
            options: {
              publicPath: PATHS.public,
            },
          },
          {
            loader: "css-loader",
            options: {
              sourceMap: true,
              importLoaders: 1,
            },
          },
          "resolve-url-loader",
          {
            loader: "postcss-loader",
            options: {
              sourceMap: true,
              postcssOptions: {
                config: path.resolve(__dirname, "./postcss.config.js"),
              },
            },
          },
          "sass-loader",
        ],
      },
      {
        test: /\.(png|jpg|jpeg|gif)$/,
        type: "asset/resource",
        dependency: { not: ["url"] },
      },
      {
        test: /\.(woff(2)?|ttf|eot)(\?v=\d+\.\d+\.\d+)?$/,
        type: "asset/resource",
        dependency: { not: ["url"] },
      },
      {
        test: /\.svg/,
        type: "asset/resource",
        dependency: { not: ["url"] },
      },
    ],
  },
};

// Don't create sourcemap for production builds.
if (isDev) {
  const devToolPlugin = new webpack.SourceMapDevToolPlugin({
    filename: "[file].map",
  });

  config.plugins.push(devToolPlugin);

  import("eslint-formatter-pretty").then(() => {
    const esLintPlugin = new ESLintPlugin({
      formatter: "eslint-formatter-pretty",
      extensions: ["js"],
    });
    config.plugins.push(esLintPlugin);
  });
}

// Analyze bundle contents with https://www.npmjs.com/package/webpack-bundle-analyzer.
if (isAnalyze) {
  const analyzerPlugin = new BundleAnalyzerPlugin();

  config.plugins.push(analyzerPlugin);
}

module.exports = config;
