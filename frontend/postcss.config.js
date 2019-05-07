module.exports = ({ file, options, env }) => ({
    plugins: {
        "postcss-import": {},
        "postcss-preset-env": true,
        "cssnano": env === "production",
    }
});
