export default class MapStyleControl {
  constructor(styles, defaultStyle, setStyle) {
    this.styles = styles || MapStyleControl.DEFAULT_STYLES;
    this.defaultStyle = defaultStyle || MapStyleControl.DEFAULT_STYLE;
    this.setStyle = setStyle;
    this.onDocumentClick = this.onDocumentClick.bind(this);
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
      styleElement.addEventListener("click", (event) => {
        const srcElement = event.srcElement;
        if (srcElement.classList.contains("active")) {
          return;
        }

        // Set new map style.
        if (typeof this.setStyle === "function") {
          this.setStyle(JSON.parse(srcElement.dataset.style));
        }

        /*
          this.mapStyleContainer.style.display = "none";
          this.styleButton.style.display = "block";
          const elms = this.mapStyleContainer.getElementsByClassName("active");
          while (elms[0]) {
            elms[0].classList.remove("active");
          }
          srcElement.classList.add("active");
        */
      });
      if (style.style === this.defaultStyle) {
        styleElement.classList.add("active");
      }
      this.mapStyleContainer.appendChild(styleElement);
    }
    this.styleButton.classList.add("maplibregl-ctrl-icon");
    this.styleButton.classList.add("maplibregl-style-switcher");
    this.styleButton.addEventListener("click", () => {
      this.styleButton.style.display = "none";
      this.mapStyleContainer.style.display = "block";
    });
    document.addEventListener("click", this.onDocumentClick);
    this.controlContainer.appendChild(this.styleButton);
    this.controlContainer.appendChild(this.mapStyleContainer);
    return this.controlContainer;
  }

  onRemove() {
    if (
      !this.controlContainer ||
      !this.controlContainer.parentNode ||
      !this.map ||
      !this.styleButton
    ) {
      return;
    }
    this.styleButton.removeEventListener("click", this.onDocumentClick);
    this.controlContainer.parentNode.removeChild(this.controlContainer);
    document.removeEventListener("click", this.onDocumentClick);
    this.map = undefined;
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
