const path = require("path");
const findChrome = require("chrome-finder");

process.env.CHROME_BIN = findChrome();

module.exports = (config) => {
    config.set({
        frameworks: ["mocha"],

        browsers: ["LocalChrome"],

        customLaunchers: {
            LocalChrome: {
                base: "ChromeHeadless",
                flags: ["--disable-translate", "--disable-extensions", "--no-sandbox", "--disable-web-security", "--disable-dev-shm-usage"],
            },
        },

        files: [
            {pattern: "tests/unit/**/*_test.js", watched: false},
        ],

        // Preprocess through webpack
        preprocessors: {
            "tests/unit/**/*_test.js": ["webpack"],
        },

        reporters: ["progress", "html"],

        htmlReporter: {
            outputFile: "tests/unit.html",
        },

        webpack: {
            mode: "development",

            resolve: {
                modules: [
                    path.join(__dirname, "src"),
                    path.join(__dirname, "node_modules"),
                    path.join(__dirname, "tests/unit"),
                ],
                alias: {
                    vue: "vue/dist/vue.js",
                },
            },
            module: {
                rules: [
                    {
                        test: /\.js$/,
                        loader: "babel-loader",
                        exclude: file => (
                            /node_modules/.test(file)
                        ),
                        query: {
                            presets: ["@babel/preset-env"],
                            compact: false,
                        },
                    },
                ],
            },
        },

        singleRun: true,
    });
};
