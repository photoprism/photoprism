<template>
  <div id="p-photo-viewer" class="p-viewer pswp" tabindex="-1" role="dialog" aria-hidden="true">
    <div class="pswp__bg"></div>
    <div class="pswp__scroll-wrap">
      <div class="pswp__container" v-bind:class="{ 'slideshow': slideshow.active }">
        <div class="pswp__item"></div>
        <div class="pswp__item"></div>
        <div class="pswp__item"></div>
      </div>

      <div class="pswp__ui pswp__ui--hidden">

        <div class="pswp__top-bar">
          <div class="pswp__taken hidden-xs-only">{{ item.taken }}</div>

          <div class="pswp__counter"></div>

          <button class="pswp__button pswp__button--close action-close" title="Close (Esc)"></button>

          <button class="pswp__button action-download" style="background: none;" @click.exact="onDownload"
                  title="Download" v-if="config.settings.features.download">
            <v-icon size="16" color="white">get_app</v-icon>
          </button>

          <button class="pswp__button action-edit hidden-shared-only" style="background: none;" @click.exact="onEdit"
                  title="Edit">
            <v-icon size="16" color="white">edit</v-icon>
          </button>

          <button class="pswp__button action-like hidden-shared-only" style="background: none;"
                  @click.exact="onLike" title="Like">
            <v-icon v-if="item.favorite" size="16" color="white">favorite</v-icon>
            <v-icon v-else size="16" color="white">favorite_border</v-icon>
          </button>

          <button class="pswp__button pswp__button--fs action-toogle-fullscreen" title="Toggle fullscreen"></button>

          <button class="pswp__button pswp__button--zoom action-zoom" title="Zoom in/out"></button>

          <button class="pswp__button" style="background: none;" @click.exact="onSlideshow" title="Slideshow">
            <v-icon v-show="!interval" size="18" color="white">play_arrow</v-icon>
            <v-icon v-show="interval" size="16" color="white">pause</v-icon>
          </button>

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

        <button class="pswp__button pswp__button--arrow--left action-previous" title="Previous (arrow left)">
        </button>

        <button class="pswp__button pswp__button--arrow--right action-next" title="Next (arrow right)">
        </button>

        <div class="pswp__caption" @click="onPlay">
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
    import Thumb from "model/thumb";
    import Photo from "model/photo";
    import Notify from "common/notify";

    export default {
        name: "p-photo-viewer",
        data() {
            return {
                config: this.$config.values,
                item: new Thumb(),
                subscriptions: [],
                interval: false,
                slideshow: {
                    active: false,
                    next: 0,
                },
            };
        },
        created() {
            this.subscriptions['viewer.change'] = Event.subscribe('viewer.change', this.onChange);
            this.subscriptions['viewer.pause'] = Event.subscribe('viewer.pause', this.onPause);
        },
        destroyed() {
            this.onPause();

            for (let i = 0; i < this.subscriptions.length; i++) {
                Event.unsubscribe(this.subscriptions[i]);
            }
        },
        methods: {
            onChange(ev, data) {
                const psp = this.$viewer.gallery;

                if(psp && this.slideshow.next !== psp.getCurrentIndex()) {
                    this.onPause();
                }

                this.item = data.item;
            },
            onLike() {
                this.item.toggleLike();
            },
            onPlay() {
                if (this.item && this.item.playable) {
                    let photo = new Photo();
                    photo.find(this.item.uid).then((p) => {
                        this.$modal.show('video', {video: p, album: null});
                    });
                }
            },
            onPause() {
                this.slideshow.active = false;

                if (this.interval) {
                    clearInterval(this.interval);
                    this.interval = false;
                }
            },
            onSlideshow() {
                if (this.interval) {
                    this.onPause();
                    return;
                }

                this.slideshow.active = true;

                const self = this;
                const psp = this.$viewer.gallery;

                self.interval = setInterval(() => {
                    if (psp && typeof psp.next === "function") {
                        psp.next();
                        this.slideshow.next = psp.getCurrentIndex();
                    } else {
                        this.onPause();
                    }
                }, 5000);
            },
            onDownload() {
                this.onPause();

                if (!this.item || !this.item.download_url) {
                    console.warn("photo viewer: no download url");
                    return;
                }

                Notify.success(this.$gettext("Downloading…"));
                let photo = new Photo();
                photo.find(this.item.uid).then((p) => {
                    p.downloadAll();
                });
            },
            onEdit() {
                this.onPause();

                const g = this.$viewer.gallery; // Gallery
                let index = 0;

                // remove duplicates
                let filtered = g.items.filter(function (p, i, s) {
                    return !(i > 0 && p.uid === s[i - 1].uid);
                });

                let selection = filtered.map((p, i) => {
                    if (g.currItem.uid === p.uid) {
                        index = i;
                    }

                    return p.uid
                });

                let album = null;

                g.close(); // Close Gallery

                Event.publish("dialog.edit", {selection, album, index}); // Open Edit Dialog
            }
        }
    }
</script>
