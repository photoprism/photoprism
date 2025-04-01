const glob = require("glob");
const path = require("path");
const poPath = path.resolve(__dirname, "src/locales");

// Generates a list of existing locales based on the files in src/locales.
const languageCodes = glob.sync(poPath + "/*.po").map((filePath) => {
  const fileName = path.basename(filePath);
  return fileName.replace(".po", "");
});

// Generates one JSON file per locale from the gettext *.po files located in src/locales.
module.exports = {
  output: {
    path: path.resolve(__dirname, "src/locales"),
    potPath: "src/locales/translations.pot",
    jsonPath: "json",
    locales: languageCodes,
    splitJson: true,
    flat: true,
  },
};
