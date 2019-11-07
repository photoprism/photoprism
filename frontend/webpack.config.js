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
    js: path.join(__dirname, "src"),
    css: path.join(__dirname, "src/css"),
    build: path.join(__dirname, "../assets/resources/static/build"),
};

const config = {
    mode: isDev ? "development" : "production",
    devtool: isDev ? "inline-source-map" : false,
    entry: {
        app: PATHS.app,
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
        maxEntrypointSize: 2000000,
        maxAssetSize: 2000000,
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
