const glob = require("glob");
const path = require("path");

const projectRoot = __dirname;
const localesPath = path.resolve(projectRoot, "src/locales");

const sourceDirs = (process.env.SRC || "src")
  .split(" ")
  .map((dir) => dir.trim())
  .filter(Boolean);

const includeExtensions = ["vue", "js", "ts"];
const includePatterns = new Set();

sourceDirs.forEach((dir) => {
  const relativeDir = dir.replace(/\\/g, "/");
  includeExtensions.forEach((ext) => {
    includePatterns.add(`${relativeDir}/**/*.${ext}`);
  });
});

if (includePatterns.size === 0) {
  includeExtensions.forEach((ext) => includePatterns.add(`src/**/*.${ext}`));
}

const excludePatterns = ["src/common/gettext.js"];

// Generates a list of existing locales based on the files in src/locales.
const languageCodes = glob.sync(path.join(localesPath, "*.po")).map((filePath) => {
  const fileName = path.basename(filePath);
  return fileName.replace(".po", "");
});

// vue3-gettext merges msgstr entries via msgmerge; disable this when GETTEXT_MERGE=0|false.
const mergeEnv = (process.env.GETTEXT_MERGE || "").toLowerCase();
const shouldMergeLocales = mergeEnv !== "0" && mergeEnv !== "false";

// Generates one JSON file per locale from the gettext *.po files located in src/locales.
module.exports = {
  input: {
    path: ".",
    include: Array.from(includePatterns),
    exclude: excludePatterns,
  },
  output: {
    path: localesPath,
    potPath: "translations.pot",
    jsonPath: "json",
    locales: shouldMergeLocales ? languageCodes : [],
    splitJson: true,
    flat: true,
    linguas: false,
  },
};
