import $event from "common/event";

// Opens the lightbox dialog with the specified options.
export class Lightbox {
  open(options) {
    $event.publish("lightbox.open", options);
  }

  openModels(models, index) {
    $event.publish("lightbox.open", { models, index });
  }

  openView(view, index) {
    $event.publish("lightbox.open", { view, index });
  }
}

export const $lightbox = new Lightbox();
