import { Selector } from "testcafe";

export default class CameraDialog {
  constructor() {
    this.root = Selector(".p-meta-camera-dialog", { timeout: 15000 });
    this.camera = this.root.find(".input-camera");
    this.lens = this.root.find(".input-lens");
    // v-select renders the chosen option in `.v-select__selection-text`.
    this.cameraValue = this.root.find(".input-camera .v-select__selection-text");
    this.lensValue = this.root.find(".input-lens .v-select__selection-text");
    this.iso = this.root.find(".input-iso input");
    this.exposure = this.root.find(".input-exposure input");
    this.fnumber = this.root.find(".input-fnumber input");
    this.focalLength = this.root.find(".input-focal-length input");
    this.confirm = this.root.find(".action-confirm");
    this.cancel = this.root.find(".action-cancel");
  }
}
