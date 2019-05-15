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
        const result = {
            title: photo.PhotoTitle,
            download_url: photo.getDownloadUrl(),
        };

        const thumbs = window.appConfig.thumbnails;

        for (let i = 0; i < thumbs.length; i++) {
            let size = photo.calculateSize(thumbs[i].Width, thumbs[i].Height);

            result[thumbs[i].Name] = {
                src: photo.getThumbnailUrl(thumbs[i].Name),
                w: size.width,
                h: size.height,
            };
        }

        return result;
    }

    getEl() {
        if (!this.el) {
            const elements = document.querySelectorAll(".pswp");

            if (elements.length !== 1) {
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


        const shareButtons = [
            {id: "download", label: "Download image", url: "{{raw_image_url}}", download: true},
        ];

        const options = {
            index: index,
            history: false,
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
            getImageURLForShare: function() { return gallery.currItem.download_url;},
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
        const thumbs = window.appConfig.thumbnails;

        for (let i = 0; i < thumbs.length; i++) {
            if (thumbs[i].Width >= viewportWidth || thumbs[i].Height >= viewportHeight) {
                return thumbs[i].Name;
            }
        }

        return "fit_720";
    }
}

export default Gallery;
