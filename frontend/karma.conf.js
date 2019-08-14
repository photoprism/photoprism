const path = require("path");
const findChrome = require("chrome-finder");

process.env.CHROME_BIN = findChrome();

module.exports = (config) => {
    config.set({
        logLevel: config.LOG_INFO,

        webpackMiddleware: {
            stats: "errors-only",
        },

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

        reporters: ["progress", "html", "coverage-istanbul"],

        htmlReporter: {
            outputFile: "tests/unit.html",
        },

        coverageIstanbulReporter: {
            // reports can be any that are listed here: https://github.com/istanbuljs/istanbuljs/tree/aae256fb8b9a3d19414dcf069c592e88712c32c6/packages/istanbul-reports/lib
            reports: ["html", "lcovonly", "text-summary"],

            // base output directory. If you include %browser% in the path it will be replaced with the karma browser name
            dir: path.join(__dirname, "coverage"),

            // Combines coverage information from multiple browsers into one report rather than outputting a report
            // for each browser.
            combineBrowserReports: true,

            // if using webpack and pre-loaders, work around webpack breaking the source path
            fixWebpackSourcePaths: true,

            // Omit files with no statements, no functions and no branches from the report
            skipFilesWithNoCoverage: true,

            // Most reporters accept additional config options. You can pass these through the `report-config` option
            "report-config": {
                // all options available at: https://github.com/istanbuljs/istanbuljs/blob/aae256fb8b9a3d19414dcf069c592e88712c32c6/packages/istanbul-reports/lib/html/index.js#L135-L137
                html: {
                    // outputs the report in ./coverage/html
                    subdir: "html",
                },
            },

            // enforce percentage thresholds
            // anything under these percentages will cause karma to fail with an exit code of 1 if not running in watch mode
            thresholds: {
                emitWarning: true, // set to `true` to not fail the test command when thresholds are not met
                // thresholds for all files
                global: {
                    //statements: 90,
                    lines: 90,
                    //branches: 90,
                    //functions: 90,
                },
                // thresholds per file
                each: {
                    //statements: 90,
                    lines: 90,
                    //branches: 90,
                    //functions: 90,
                    overrides: {
                        "src/common/viewer.js": {
                            lines: 0,
                            functions: 0,
                        },
                    },
                },
            },

            verbose: false, // output config used by istanbul for debugging
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
                    vue: "vue/dist/vue.min.js",
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
