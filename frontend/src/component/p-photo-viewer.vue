<template>
    <div id="p-photo-viewer" class="p-viewer pswp" tabindex="-1" role="dialog" aria-hidden="true">
        <div class="pswp__bg"></div>
        <div class="pswp__scroll-wrap">
            <div class="pswp__container">
                <div class="pswp__item"></div>
                <div class="pswp__item"></div>
                <div class="pswp__item"></div>
            </div>

            <div class="pswp__ui pswp__ui--hidden">

                <div class="pswp__top-bar">
                    <div class="pswp__counter"></div>

                    <button class="pswp__button pswp__button--close" title="Close (Esc)"></button>

                    <button class="pswp__button pswp__button--share p-photo-download" title="Download"
                            v-if="config.settings.features.download">
                    </button>

                    <button class="pswp__button" style="background: none;" @click.exact="editDialog" title="Edit">
                        <v-icon size="16" color="white">edit</v-icon>
                    </button>

                    <button class="pswp__button pswp__button--fs" title="Toggle fullscreen"></button>

                    <button class="pswp__button pswp__button--zoom" title="Zoom in/out"></button>

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
</template>

<script>
    import 'photoswipe/dist/photoswipe.css'
    import 'photoswipe/dist/default-skin/default-skin.css'
    import Event from "pubsub-js";

    export default {
        name: "p-photo-viewer",
        data() {
            return {
                config: this.$config.values,
            };
        },
        methods: {
            editDialog() {
                const g = this.$viewer.gallery; // Gallery
                let index = 0;

                // remove duplicates
                let filtered = g.items.filter(function (p, i, s) {
                    return !(i > 0 && p.uuid === s[i - 1].uuid);
                });

                let selection = filtered.map((p, i) => {
                    if (g.currItem.uuid === p.uuid) {
                        index = i;
                    }

                    return p.uuid
                });

                let album = null;

                g.close(); // Close Gallery

                Event.publish("dialog.edit", {selection, album, index}); // Open Edit Dialog
            }
        }
    }
</script>
