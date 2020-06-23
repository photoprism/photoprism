/*

Copyright (c) 2018 - 2020 Michael Mayer <hello@photoprism.org>

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

    PhotoPrism™ is a registered trademark of Michael Mayer.  You may use it as required
    to describe our software, run your own server, for educational purposes, but not for
    offering commercial goods, products, or services without prior written permission.
    In other words, please ask.

Feel free to send an e-mail to hello@photoprism.org if you have questions,
want to support our work, or just want to say hello.

Additional information can be found in our Developer Guide:
https://docs.photoprism.org/developer-guide/

*/

const path = require("path");

const MiniCssExtractPlugin = require("mini-css-extract-plugin");

const webpack = require("webpack");

const isDev = process.env.NODE_ENV !== "production";

if(isDev) {
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
};

const config = {
    mode: isDev ? "development" : "production",
    devtool: isDev ? "inline-source-map" : false,
    entry: {
        app: PATHS.app,
        share: PATHS.share,
    },
    output: {
        path: PATHS.build,
        filename: "[name].js",
    },
    resolve: {
        modules: [
            path.join(__dirname, "src"),
            path.join(__dirname, "node_modules"),
        ],
        alias: {
            vue: isDev ? "vue/dist/vue.js" : "vue/dist/vue.min.js",
        },
    },
    plugins: [
        new MiniCssExtractPlugin({
            filename: "[name].css",
        }),
    ],
    node: {
        fs: "empty",
    },
    performance: {
        hints: isDev ? false : "error",
        maxEntrypointSize: 4000000,
        maxAssetSize: 4000000,
    },
    module: {
        rules: [
            {
                test: /\.js$/,
                include: PATHS.app,
                exclude: /node_modules/,
                enforce: "pre",
                loader: "eslint-loader",
                options: {
                    formatter: require("eslint-formatter-pretty"),
                },
            },
            {
                test: /\.vue$/,
                loader: "vue-loader",
                include: PATHS.js,
                options: {
                    loaders: {
                        js: "babel-loader",
                        css: "css-loader",
                    },
                },
            },
            {
                test: /\.js$/,
                loader: "babel-loader",
                include: PATHS.js,
                exclude: file => (
                    /node_modules/.test(file)
                ),
                query: {
                    presets: ["@babel/preset-env"],
                    compact: false,
                },
            },
            {
                test: /\.css$/,
                include: PATHS.css,
                exclude: /node_modules/,
                use: [
                    {
                        loader: MiniCssExtractPlugin.loader,
                        options: {
                            hmr: false,
                            fallback: "vue-style-loader",
                            use: [
                                "style-loader",
                                {
                                    loader: "css-loader",
                                    options: {
                                        importLoaders: 1,
                                        sourceMap: isDev,
                                    },
                                },
                                {
                                    loader: "postcss-loader",
                                    options: {
                                        sourceMap: isDev,
                                        config: {
                                            path: path.resolve(__dirname, "./postcss.config.js"),
                                        },
                                    },
                                },
                                "resolve-url-loader",
                            ],
                            publicPath: PATHS.build,
                        },
                    },
                    "css-loader",
                ],
            },
            {
                test: /\.css$/,
                include: /node_modules/,
                loaders: [
                    "vue-style-loader",
                    "style-loader",
                    {
                        loader: "css-loader",
                        options: { importLoaders: 1, sourceMap: isDev },
                    },
                    {
                        loader: "postcss-loader",
                        options: {
                            sourceMap: isDev,
                            config: {
                                path: path.resolve(__dirname, "./postcss.config.js"),
                            },
                        },
                    },
                    "resolve-url-loader",
                ],
            },
            {
                test: /\.s[c|a]ss$/,
                use: [
                    "vue-style-loader",
                    "style-loader",
                    {
                        loader: "css-loader",
                        options: { importLoaders: 2, sourceMap: isDev },
                    },
                    {
                        loader: "postcss-loader",
                        options: {
                            sourceMap: isDev,
                            config: {
                                path: path.resolve(__dirname, "./postcss.config.js"),
                            },
                        },
                    },
                    "resolve-url-loader",
                    "sass-loader",
                ],

            },
            {
                test: /\.(png|jpg|jpeg|gif)$/,
                loader: "file-loader",
                options: {
                    name: "[hash].[ext]",
                    publicPath: "/static/build/img",
                    outputPath: "img",
                },
            },
            {
                test: /\.(woff(2)?|ttf|eot)(\?v=\d+\.\d+\.\d+)?$/,
                loader: "file-loader",
                options: {
                    name: "[hash].[ext]",
                    publicPath: "/static/build/fonts",
                    outputPath: "fonts",
                },
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

// No sourcemap for production
if (isDev) {
    const devToolPlugin = new webpack.SourceMapDevToolPlugin({
        filename: "[file].map",
    });

    config.plugins.push(devToolPlugin);
}

module.exports = config;
