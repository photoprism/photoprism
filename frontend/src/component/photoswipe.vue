<template>
    <div>
        <slot v-bind:openGallery="openGallery"></slot>
        <!-- Root element of PhotoSwipe. Must have class pswp. -->
        <div class="pswp" tabindex="-1" role="dialog" aria-hidden="true">

            <!-- Background of PhotoSwipe.
                 It's a separate element as animating opacity is faster than rgba(). -->
            <div class="pswp__bg"></div>

            <!-- Slides wrapper with overflow:hidden. -->
            <div class="pswp__scroll-wrap">

                <!-- Container that holds slides.
                    PhotoSwipe keeps only 3 of them in the DOM to save memory.
                    Don't modify these 3 pswp__item elements, data is added later on. -->
                <div class="pswp__container">
                    <div class="pswp__item"></div>
                    <div class="pswp__item"></div>
                    <div class="pswp__item"></div>
                </div>

                <!-- Default (PhotoSwipeUI_Default) interface on top of sliding area. Can be changed. -->
                <div class="pswp__ui pswp__ui--hidden">

                    <div class="pswp__top-bar">

                        <!--  Controls are self-explanatory. Order can be changed. -->

                        <div class="pswp__counter"></div>

                        <button class="pswp__button pswp__button--close" title="Close (Esc)"></button>

                        <button class="pswp__button pswp__button--share" title="Share"></button>

                        <button class="pswp__button pswp__button--fs" title="Toggle fullscreen"></button>

                        <button class="pswp__button pswp__button--zoom" title="Zoom in/out"></button>

                        <!-- Preloader demo https://codepen.io/dimsemenov/pen/yyBWoR -->
                        <!-- element will get class pswp__preloader--active when preloader is running -->
                        <div class="pswp__preloader">
                            <div class="pswp__preloader__icn">
                                <div class="pswp__preloader__cut">
                                    <div class="pswp__preloader__donut"></div>
                                </div>
                            </div>
                        </div>
                    </div>

                    <div class="pswp__share-modal pswp__share-modal--hidden pswp__single-tap">
                        <div class="pswp__share-tooltip"></div>
                    </div>

                    <button class="pswp__button pswp__button--arrow--left" title="Previous (arrow left)">
                    </button>

                    <button class="pswp__button pswp__button--arrow--right" title="Next (arrow right)">
                    </button>

                    <div class="pswp__caption">
                        <div class="pswp__caption__center"></div>
                    </div>

                </div>

            </div>
        </div>
    </div>
</template>

<script>
    import PhotoSwipe from 'photoswipe'
    import PhotoSwipeUI_Default from 'photoswipe/dist/photoswipe-ui-default.js'
    import 'photoswipe/dist/photoswipe.css'
    import 'photoswipe/dist/default-skin/default-skin.css'

    export default {
        name: "photoswipe",
        props: {
            images: Array
        },
        computed: {
            imagesWithSizes: function() {
                return this.images.map(this.createPhotoSizes);
            }
        },
        methods: {
            createPhotoSizes(photo) {
                const createPhotoSize = height => ({
                    src: photo.getThumbnailUrl('fit', height),
                    w: photo.calculateWidth(height),
                    h: height,
                    title: photo.PhotoTitle
                });

                return {
                    xxs: createPhotoSize(320),
                    xs: createPhotoSize(500),
                    s: createPhotoSize(720),
                    m: createPhotoSize(1280),
                    l: createPhotoSize(1920),
                    xl: createPhotoSize(2560),
                    xxl: createPhotoSize(3840)
                }
            },

            mapViewportToImageSize(viewportWidth, viewportHeight, item) {
                for (const [sizeKey, photo] of Object.entries(item)) {
                    if (photo.w > viewportWidth || photo.h > viewportHeight) {
                        return sizeKey
                    }
                }
            },

            openGallery: function () {
            },

            openPhoto: function (index = 0) {
                if (this.$props.images.length === 0) {
                    return
                }

                const pswpElement = document.querySelectorAll('.pswp')[0];

                const options = {
                    index
                };

                let gallery = new PhotoSwipe(pswpElement, PhotoSwipeUI_Default, this.imagesWithSizes, options);
                let realViewportWidth;
                let realViewportHeight;
                let previousSize;
                let nextSize;
                let firstResize = true;
                let imageSrcWillChange;

                gallery.listen('beforeResize', () => {
                    realViewportWidth = gallery.viewportSize.x * window.devicePixelRatio;
                    realViewportHeight = gallery.viewportSize.y * window.devicePixelRatio;

                    if (!previousSize) {
                        previousSize = 'm'
                    }

                    nextSize = this.mapViewportToImageSize(realViewportWidth, realViewportHeight, this.imagesWithSizes[index])
                    if (nextSize !== previousSize) {
                        imageSrcWillChange = true
                    }

                    if (imageSrcWillChange && !firstResize) {
                        gallery.invalidateCurrItems();
                    }

                    if (firstResize) {
                        firstResize = false;
                    }

                    imageSrcWillChange = false;
                });


                gallery.listen('gettingData', function (index, item) {
                    item.src = item[nextSize].src;
                    item.w = item[nextSize].w;
                    item.h = item[nextSize].h;
                    item.title = item[nextSize].title;
                    previousSize = nextSize;
                });

                gallery.init();
            }
        }
    }
</script>
