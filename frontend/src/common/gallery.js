import PhotoSwipe from "photoswipe";
import PhotoSwipeUI_Default from "photoswipe/dist/photoswipe-ui-default.js";

class Gallery {
    constructor() {
        this.photos = [];
        this.el = null;
    }

    photosWithSizes() {
        return this.photos.map(this.createPhotoSizes);
    }

    createPhotoSizes(photo) {
        const createPhotoSize = height => ({
            src: photo.getThumbnailUrl("fit", height),
            w: photo.calculateWidth(height),
            h: height,
            title: photo.PhotoTitle,
        });

        return {
            xxs: createPhotoSize(320),
            xs: createPhotoSize(500),
            s: createPhotoSize(720),
            m: createPhotoSize(1280),
            l: createPhotoSize(1920),
            xl: createPhotoSize(2560),
            xxl: createPhotoSize(3840),
        };
    }

    getEl() {
        if(!this.el) {
            const elements = document.querySelectorAll(".pswp");

            if(elements.length !== 1) {
                let err = "There should be only one PhotoSwipe element";
                console.log(err, elements);
                throw err;
            }

            this.el = elements[0];
        }

        return this.el;
    }

    show(photos, index = 0) {
        if (!Array.isArray(photos) || photos.length === 0 || index >= photos.length) {
            console.log("Array passed to gallery was empty:", photos);
            return;
        }

        this.photos = photos;

        const options = {
            index: index,
            history: false,
            preload: true,
            focus: true,
            modal: true,
            closeEl: true,
            captionEl: true,
            fullscreenEl: true,
            zoomEl: true,
            shareEl: false,
            counterEl: false,
            arrowEl: true,
            preloaderEl: true,
        };

        let photosWithSizes = this.photosWithSizes();
        let gallery = new PhotoSwipe(this.getEl(), PhotoSwipeUI_Default, photosWithSizes, options);
        let realViewportWidth;
        let realViewportHeight;
        let previousSize;
        let nextSize;
        let firstResize = true;
        let photoSrcWillChange;

        gallery.listen("beforeResize", () => {
            realViewportWidth = gallery.viewportSize.x * window.devicePixelRatio;
            realViewportHeight = gallery.viewportSize.y * window.devicePixelRatio;

            if (!previousSize) {
                previousSize = "m";
            }

            nextSize = this.constructor.mapViewportToImageSize(realViewportWidth, realViewportHeight, photosWithSizes[index]);
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
            item.title = item[nextSize].title;
            previousSize = nextSize;
        });

        gallery.init();
    }

    static mapViewportToImageSize(viewportWidth, viewportHeight, item) {
        for (const [sizeKey, photo] of Object.entries(item)) {
            if (photo.w > viewportWidth || photo.h > viewportHeight) {
                return sizeKey;
            }
        }
    }
}

export default Gallery;
