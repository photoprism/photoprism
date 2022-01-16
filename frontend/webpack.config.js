/*

Copyright (c) 2018 - 2022 Michael Mayer <hello@photoprism.app>

    This program is free software: you can redistribute it and/or modify
    it under the terms of the GNU Affero General Public License as published
    by the Free Software Foundation, either version 3 of the License, or
    (at your option) any later version.

    This program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU Affero General Public License for more details.

    You should have received a copy of the GNU Affero General Public License
    along with this program.  If not, see <https://www.gnu.org/licenses/>.

    PhotoPrism® is a registered trademark of Michael Mayer.  You may use it as required
    to describe our software, run your own server, for educational purposes, but not for
    offering commercial goods, products, or services without prior written permission.
    In other words, please ask.

Feel free to send an e-mail to hello@photoprism.app if you have questions,
want to support our work, or just want to say hello.

Additional information can be found in our Developer Guide:
https://docs.photoprism.app/developer-guide/

*/

const path = require("path");
const MiniCssExtractPlugin = require("mini-css-extract-plugin");
const ESLintPlugin = require("eslint-webpack-plugin");
const OfflinePlugin = require("@lcdp/offline-plugin");
const webpack = require("webpack");
const isDev = process.env.NODE_ENV !== "production";
const { VueLoaderPlugin } = require("vue-loader");

if (isDev) {
  console.log("Building frontend in DEVELOPMENT mode. Please wait.");
} else {
  console.log("Building frontend in PRODUCTION mode. Please wait.");
}

const PATHS = {
  app: path.join(__dirname, "src/app.js"),
  share: path.join(__dirname, "src/share.js"),
  js: path.join(__dirname, "src"),
  css: path.join(__dirname, "src/css"),
  build: path.join(__dirname, "../assets/static/build"),
  public: "/static/build/",
};

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
    filename: "[name].js",
  },
  resolve: {
    modules: [path.join(__dirname, "src"), path.join(__dirname, "node_modules")],
    alias: {
      vue: isDev ? "vue/dist/vue.js" : "vue/dist/vue.min.js",
    },
  },
  plugins: [
    new MiniCssExtractPlugin({
      filename: "[name].css",
      experimentalUseImportModule: false,
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
    maxEntrypointSize: 4000000,
    maxAssetSize: 4000000,
  },
  module: {
    rules: [
      {
        test: /\.vue$/,
        include: PATHS.js,
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
        include: PATHS.js,
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
        include: PATHS.css,
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
        use: {
          loader: "svg-url-loader",
          options: {},
        },
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
