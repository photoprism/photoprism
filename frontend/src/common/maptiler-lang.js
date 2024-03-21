/* @preserve
 * https://github.com/klokantech/openmaptiles-language
 * (c) 2018 Klokan Technologies GmbH
 */
import maplibregl from "maplibre-gl";
import { config } from "app/session";

const langFallbackDecorate = function (style, cfg) {
  let layers = style.layers;
  let lf = cfg["layer-filter"];
  let decorators = cfg["decorators"];
  let lfProp = lf[1];
  let lfValues = lf.slice(2);

  for (let i = layers.length - 1; i >= 0; i--) {
    let layer = layers[i];
    if (!(lf[0] === "in" && lfProp === "layout.text-field" && layer.layout && layer.layout["text-field"] && lfValues.indexOf(layer.layout["text-field"]) >= 0)) {
      continue;
    }
    for (let j = decorators.length - 1; j >= 0; j--) {
      let decorator = decorators[j];
      let postfix = decorator["layer-name-postfix"] || "";
      postfix = postfix.replace(/(^-+|-+$)/g, "");
      let newLayer;
      if (j > 0) {
        newLayer = JSON.parse(JSON.stringify(layer));
        layers.splice(i + 1, 0, newLayer);
      } else {
        newLayer = layer;
      }
      newLayer.id += postfix ? "-" + postfix : "";
      newLayer.layout["text-field"] = decorator["layout.text-field"];
      if (newLayer.layout["symbol-placement"] === "line") {
        newLayer.layout["text-field"] = newLayer.layout["text-field"].replace("\n", " ");
      }
      let filterPart = decorator["filter-all-part"].concat();
      if (!newLayer.filter) {
        newLayer.filter = filterPart;
      } else if (newLayer.filter[0] === "all") {
        newLayer.filter.push(filterPart);
      } else {
        newLayer.filter = ["all", newLayer.filter, filterPart];
      }
    }
  }
};

let langEnabled = true;

let setStyleMutex = false;
let origSetStyle = maplibregl.Map.prototype.setStyle;
maplibregl.Map.prototype.setStyle = function () {
  origSetStyle.apply(this, arguments);

  if (langEnabled && !setStyleMutex) {
    if (this.styleUndecorated) {
      this.styleUndecorated = undefined;
    }
    this.once(
      "styledata",
      function () {
        if (this.languageOptions) {
          this.setLanguage(this.languageOptions.language, this.languageOptions.noAlt);
        }
      }.bind(this)
    );
  }
};

maplibregl.Map.prototype.setLanguageEnabled = function (enable) {
  langEnabled = enable;
};

maplibregl.Map.prototype.setLanguage = function (language, noAlt) {
  this.languageOptions = {
    language: language,
    noAlt: noAlt,
  };
  if (!this.styleUndecorated) {
    try {
      this.styleUndecorated = this.getStyle();
    } catch (e) {
      // Ignore.
    }
  }
  if (!this.styleUndecorated) {
    return;
  }

  let isNonlatin = ["ar", "hy", "be", "bg", "zh", "ka", "el", "he", "ja", "ja_kana", "kn", "kk", "ko", "mk", "ru", "sr", "th", "uk"].indexOf(language) >= 0;

  let style = JSON.parse(JSON.stringify(this.styleUndecorated));
  let langCfg = {
    "layer-filter": ["in", "layout.text-field", "{name}", "{name_de}", "{name_en}", "{name:latin}", "{name:latin} {name:nonlatin}", "{name:latin}\n{name:nonlatin}"],
    decorators: [
      {
        "layout.text-field": isNonlatin ? "{name:nonlatin}" + (noAlt ? "" : "\n{name:latin}") : "{name:latin}" + (noAlt ? "" : "\n{name:nonlatin}"),
        "filter-all-part": ["!has", "name:" + language],
      },
      {
        "layer-name-postfix": language,
        "layout.text-field": "{name:" + language + "}" + (noAlt ? "" : "\n{name:" + (isNonlatin ? "latin" : "nonlatin") + "}"),
        "filter-all-part": ["has", "name:" + language],
      },
    ],
  };
  if (language === "native") {
    langCfg["decorators"] = [
      {
        "layout.text-field": "{name}",
        "filter-all-part": ["all"],
      },
    ];
  }
  langFallbackDecorate(style, langCfg);

  setStyleMutex = true;
  this.setStyle(style);
  setStyleMutex = false;
};

maplibregl.Map.prototype.autodetectLanguage = function (opt_fallback) {
  this.setLanguage(config.values.settings.ui.language.split("-")[0] || opt_fallback || "native");
};

// Add plugin to support right-to-left languages such as Arabic and Hebrew.
maplibregl.setRTLTextPlugin(
  `${config.staticUri}/plugins/maplibre-gl-rtl-text/v0.2.3/maplibre-gl-rtl-text.js`,
  null,
  true // Lazy load the plugin
);
