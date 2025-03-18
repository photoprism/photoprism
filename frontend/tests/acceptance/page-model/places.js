import { Selector, t } from "testcafe";

export default class Page {
  constructor() {
    this.placesSearch = Selector("div.search-control input");
    this.openClusterInSearch = Selector("div.cluster-control-container button.action-browse");
    this.closeCluster = Selector("div.cluster-control-container button.action-close");
    this.clearLocation = Selector("button.action-clear-location");
    this.zoomOut = Selector("button.maplibregl-ctrl-zoom-out");
    this.zoomIn = Selector("button.maplibregl-ctrl-zoom-in");
    this.switchStyle = Selector("button.maplibregl-style-switcher");
  }

  async search(term) {
    await t.typeText(this.placesSearch, term, { replace: true }).pressKey("enter");
  }
}
