/*

Copyright (c) 2018 - 2024 PhotoPrism UG. All rights reserved.

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
const isDev = process.env.NODE_ENV !== "production";
const isCustom = !!process.env.CUSTOM_SRC;
const appName = process.env.CUSTOM_NAME ? process.env.CUSTOM_NAME : "PhotoPrism";
const { VueLoaderPlugin } = require("vue-loader");

const PATHS = {
  src: path.join(__dirname, "src"),
  css: path.join(__dirname, "src/css"),
  modules: path.join(__dirname, "node_modules"),
  app: path.join(__dirname, "src/app.js"),
  share: path.join(__dirname, "src/share.js"),
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
      vue: isDev ? "vue/dist/vue.js" : "vue/dist/vue.min.js",
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
  ],
  performance: {
    hints: isDev ? false : "error",
    maxEntrypointSize: 5000000,
    maxAssetSize: 5000000,
  },
  module: {
    rules: [
      {
        test: /\.vue$/,
        include: isCustom ? [PATHS.custom, PATHS.src] : [PATHS.src],
        use: [
          {
            loader: "vue-loader",
            options: {
              loaders: {
                js: "babel-loader",
                css: "css-loader",
              },
            },
          },
        ],
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
              plugins: [
                "@babel/plugin-proposal-object-rest-spread",
                "@babel/plugin-proposal-class-properties",
              ],
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

  const esLintPlugin = new ESLintPlugin({
    formatter: require("eslint-formatter-pretty"),
    extensions: ["js"],
  });

  config.plugins.push(esLintPlugin);
}

module.exports = config;
