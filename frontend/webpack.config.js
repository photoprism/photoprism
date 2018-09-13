const path = require('path');

const ExtractTextPlugin = require('extract-text-webpack-plugin');

const PATHS = {
    app: path.join(__dirname, 'src/app.js'),
    css: path.join(__dirname, 'css'),
    build: path.join(__dirname, '../server/assets/public/build'),
};

const cssPlugin = new ExtractTextPlugin({
    filename: '[name].css',
});

// See https://github.com/webpack/loader-utils/issues/56
process.noDeprecation = true;

const config = {
    entry: {
        app: PATHS.app,
    },
    output: {
        path: PATHS.build,
        filename: '[name].js',
    },
    resolve: {
        modules: [
            path.join(__dirname, 'src'),
            path.join(__dirname, 'node_modules'),
        ],
        alias: {
            vue: 'vue/dist/vue.js',
        },
    },
    plugins: [
        cssPlugin,
    ],
    node: {
        fs: 'empty',
    },
    module: {
        rules: [
            {
                test: /\.(js)$/,
                include: PATHS.app,
                enforce: 'pre',
                loader: 'eslint-loader',
            },
            {
                test: /\.js$/,
                loader: 'babel-loader',
                query: {
                    presets: ['es2015'],
                },
            },
            {
                test: /\.vue$/,
                loader: 'vue-loader',
                options: {
                    loaders: {
                        js: 'babel-loader?presets[]=es2015',
                    },
                },
            },
            {
                test: /\.css$/,
                include: PATHS.css,
                exclude: /node_modules/,
                use: cssPlugin.extract({
                    use: 'css-loader',
                    fallback: 'style-loader',
                }),
            },
            {
                test: /\.css$/,
                include: /node_modules/,
                loaders: ['style-loader', 'css-loader']
            },
            {
                test: /\.scss$/,
                loaders: ['style-loader', 'css-loader', 'sass-loader']
            },
            {
                test: /\.(png|jpg|jpeg|gif|svg|woff|woff2)$/,
                loader: 'url-loader',
            },
            {
                test: /\.(wav|mp3|eot|ttf)$/,
                loader: 'file-loader',
            },
            {
                test: /\.svg/,
                use: {
                    loader: 'svg-url-loader',
                    options: {},
                },
            },
        ],
    },
};

// No sourcemap for production
if (process.env.NODE_ENV === "production") {
    config.devtool = "";
}

module.exports = config;