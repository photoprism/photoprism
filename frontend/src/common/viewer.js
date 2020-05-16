import PhotoSwipe from "photoswipe";
import PhotoSwipeUI_Default from "photoswipe/dist/photoswipe-ui-default.js";
import Event from "pubsub-js";

const thumbs = window.clientConfig.thumbnails;

class Viewer {
    constructor() {
        this.el = null;
        this.gallery = null;
    }

    getEl() {
        if (!this.el) {
            this.el = document.getElementById("p-photo-viewer");

            if (this.el === null) {
                let err = "no photo viewer element found";
                console.warn(err);
                throw err;
            }
        }

        return this.el;
    }

    show(items, index = 0) {
        if (!Array.isArray(items) || items.length === 0 || index >= items.length) {
            console.log("Array passed to gallery was empty:", items);
            return;
        }

        const shareButtons = [
            {id: "fit_720", template: "Tiny (size)", label: "Tiny", url: "{{raw_image_url}}", download: true},
            {id: "fit_1280", template: "Small (size)", label: "Small", url: "{{raw_image_url}}", download: true},
            {id: "fit_2048", template: "Medium (size)", label: "Medium", url: "{{raw_image_url}}", download: true},
            {id: "fit_2560", template: "Large (size)", label: "Large", url: "{{raw_image_url}}", download: true},
            {id: "original", template: "Original (size)", label: "Original", url: "{{raw_image_url}}", download: true},
        ];

        const options = {
            index: index,
            history: true,
            preload: [1, 1],
            focus: true,
            modal: true,
            closeEl: true,
            captionEl: true,
            fullscreenEl: true,
            zoomEl: true,
            shareEl: true,
            shareButtons: shareButtons,
            counterEl: false,
            arrowEl: true,
            preloaderEl: true,
            getImageURLForShare: function (button) {
                const item = gallery.currItem;

                if (!item.original_w) {
                    button.label = button.template.replace("size", "not available");
                    return item.download_url;
                }

                if(button.id === "original") {
                    button.label = button.template.replace("size", item.original_w + " × " + item.original_h);
                    return item.download_url;
                } else {
                    button.label = button.template.replace("size", item[button.id].w + " × " + item[button.id].h);
                    return item[button.id].src + "?download=1";
                }
            },
        };

        let gallery = new PhotoSwipe(this.getEl(), PhotoSwipeUI_Default, items, options);
        let realViewportWidth;
        let realViewportHeight;
        let previousSize;
        let nextSize;
        let firstResize = true;
        let photoSrcWillChange;

        this.gallery = gallery;

        gallery.listen("beforeChange", function() {
            Event.publish("viewer.change", {gallery: gallery, item: gallery.currItem});
        });

        gallery.listen("beforeResize", () => {
            realViewportWidth = gallery.viewportSize.x * window.devicePixelRatio;
            realViewportHeight = gallery.viewportSize.y * window.devicePixelRatio;

            if (!previousSize) {
                previousSize = "tile_720";
            }

            nextSize = this.constructor.mapViewportToImageSize(realViewportWidth, realViewportHeight);

            if (nextSize !== previousSize) {
                photoSrcWillChange = true;
            }

            if (photoSrcWillChange && !firstResize) {
                gallery.invalidateCurrItems();
            }

            if (firstResize) {
                firstResize = false;
            }

            photoSrcWillChange = false;
        });

        gallery.listen("gettingData", function (index, item) {
            item.src = item[nextSize].src;
            item.w = item[nextSize].w;
            item.h = item[nextSize].h;
            previousSize = nextSize;
        });

        gallery.init();
    }

    static mapViewportToImageSize(viewportWidth, viewportHeight) {
        for (let i = 0; i < thumbs.length; i++) {
            if (thumbs[i].Width >= viewportWidth || thumbs[i].Height >= viewportHeight) {
                return thumbs[i].Name;
            }
        }

        return "fit_720";
    }
}

export default Viewer;
