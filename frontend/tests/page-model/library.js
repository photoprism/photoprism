import { Selector, t } from "testcafe";

export default class Page {
  constructor() {
    this.openImportFolderSelect = Selector(".input-import-folder input", { timeout: 15000 });
    this.import = Selector(".action-import");
    this.indexFolderSelect = Selector(".input-index-folder input", { timeout: 15000 });
    this.index = Selector(".action-index");
  }
}
