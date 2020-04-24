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

                    <button class="pswp__button" style="background: none;" @click.exact="onEdit" title="Edit">
                        <v-icon size="16" color="white">edit</v-icon>
                    </button>

                    <button class="pswp__button" style="background: none;" @click.exact="toggleLike" title="Like">
                        <v-icon v-if="item.favorite" size="16" color="white">favorite</v-icon>
                        <v-icon v-else size="16" color="white">favorite_border</v-icon>
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
    import Thumb from "../model/thumb";

    export default {
        name: "p-photo-viewer",
        data() {
            return {
                config: this.$config.values,
                item: new Thumb(),
                subscriptions: [],
            };
        },
        created() {
            this.subscriptions['viewer.change'] = Event.subscribe('viewer.change', this.onChange);
        },
        destroyed() {
            for (let i = 0; i < this.subscriptions.length; i++) {
                Event.unsubscribe(this.subscriptions[i]);
            }
        },
        methods: {
            onChange(ev, data) {
                this.item = data.item;
            },
            toggleLike() {
                this.item.toggleLike();
            },
            onEdit() {
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
