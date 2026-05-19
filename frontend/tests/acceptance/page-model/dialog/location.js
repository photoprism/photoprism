import { Selector } from "testcafe";

export default class LocationDialog {
  constructor() {
    this.root = Selector(".p-meta-location-dialog", { timeout: 15000 });
    this.coordinates = this.root.find(".input-coordinates input");
    this.confirm = this.root.find(".action-confirm");
    this.cancel = this.root.find(".action-cancel");
  }
}
