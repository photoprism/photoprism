export default class MapStyleControl {
  constructor(styles, defaultStyle, setStyle) {
    this.styles = styles || MapStyleControl.DEFAULT_STYLES;
    this.defaultStyle = defaultStyle || MapStyleControl.DEFAULT_STYLE;
    this.setStyle = setStyle;
    this.onDocumentClick = this.onDocumentClick.bind(this);
    // Keep refs to per-element handlers so they can be removed onRemove().
    this._styleElementHandlers = [];
    // Handler for the style button (set in onAdd).
    this._styleButtonHandler = null;
  }

  getDefaultPosition() {
    return "top-right";
  }

  onAdd(map) {
    this.map = map;
    this.controlContainer = document.createElement("div");
    this.controlContainer.classList.add("maplibregl-ctrl");
    this.controlContainer.classList.add("maplibregl-ctrl-group");
    this.mapStyleContainer = document.createElement("div");
    this.styleButton = document.createElement("button");
    this.styleButton.type = "button";
    this.mapStyleContainer.classList.add("maplibregl-style-list");
    for (const style of this.styles) {
      const styleElement = document.createElement("button");
      styleElement.type = "button";
      styleElement.innerText = style.title;
      styleElement.classList.add(style.style, "_");
      styleElement.dataset.style = JSON.stringify(style.style);

      // Create a named handler so we can remove it later.
      const styleValue = style.style;
      const handler = (event) => {
        const srcElement = event.currentTarget || event.target || event.srcElement;
        if (!srcElement || !srcElement.classList) {
          return;
        }

        if (srcElement.classList.contains("active")) {
          this.mapStyleContainer.style.display = "none";
          this.styleButton.style.display = "block";
          return;
        }

        // Set new map style.
        if (typeof this.setStyle === "function") {
          this.setStyle(styleValue);
        }
      };

      styleElement.addEventListener("click", handler);
      this._styleElementHandlers.push({ el: styleElement, handler });

      if (style.style === this.defaultStyle) {
        styleElement.classList.add("active");
      }

      this.mapStyleContainer.appendChild(styleElement);
    }

    this.styleButton.classList.add("maplibregl-ctrl-icon");
    this.styleButton.classList.add("maplibregl-style-switcher");
    this._styleButtonHandler = (/* event */) => {
      this.styleButton.style.display = "none";
      this.mapStyleContainer.style.display = "block";
    };

    this.styleButton.addEventListener("click", this._styleButtonHandler);
    document.addEventListener("click", this.onDocumentClick);
    this.controlContainer.appendChild(this.styleButton);
    this.controlContainer.appendChild(this.mapStyleContainer);
    return this.controlContainer;
  }

  onRemove() {
    document.removeEventListener("click", this.onDocumentClick);

    // Remove per-style element handlers.
    for (const item of this._styleElementHandlers) {
      if (item?.el && typeof item.el.removeEventListener === "function" && typeof item.handler === "function") {
        item.el.removeEventListener("click", item.handler);
      }
    }
    this._styleElementHandlers = [];

    // Remove style button handler.
    if (this.styleButton && this._styleButtonHandler) {
      this.styleButton.removeEventListener("click", this._styleButtonHandler);
      this._styleButtonHandler = null;
    }

    if (this.controlContainer?.parentNode) {
      this.controlContainer.parentNode.removeChild(this.controlContainer);
    }

    this.map = undefined;
    this.controlContainer = undefined;
    this.mapStyleContainer = undefined;
    this.styleButton = undefined;
  }

  onDocumentClick(event) {
    if (
      this.controlContainer &&
      !this.controlContainer.contains(event.target) &&
      this.mapStyleContainer &&
      this.styleButton
    ) {
      this.mapStyleContainer.style.display = "none";
      this.styleButton.style.display = "block";
    }
  }
}

MapStyleControl.DEFAULT_STYLE = "default";
MapStyleControl.DEFAULT_STYLES = [
  {
    title: "Default",
    style: "default",
  },
];
