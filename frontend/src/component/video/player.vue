<template>
  <video class="p-video-player" ref="player" :height="height" :width="width" :autoplay="autoplay"
         :preload="preload"></video>
</template>

<script>
    import "mediaelement";

    export default {
        name: "p-photo-player",
        props: {
            show: {
                type: Boolean,
                required: false,
                default: false
            },
            source: {
                type: String,
                required: true,
                default: ""
            },
            width: {
                type: String,
                required: false,
                default: "auto"
            },
            height: {
                type: String,
                required: false,
                default: "auto"
            },
            preload: {
                type: String,
                required: false,
                default: "none"
            },
            autoplay: {
                type: Boolean,
                required: false,
                default: false
            },
            success: {
                type: Function,
                default() {
                    return false;
                }
            },
            error: {
                type: Function,
                default() {
                    return false;
                }
            }
        },
        data: () => ({
            refresh: false,
            player: null,
        }),
        mounted() {
            this.render();
        },
        methods: {
            render() {
                const {MediaElementPlayer} = global;

                const self = this;
                this.player = new MediaElementPlayer(this.$el, {
                    videoWidth: this.width,
                    videoHeight: this.height,
                    pluginPath: "/static/build/",
                    shimScriptAccess: "always",
                    forceLive: false,
                    loop: false,
                    stretching: true,
                    autoplay: true,
                    setDimensions: true,
                    success: (mediaElement, originalNode, instance) => {
                        instance.setSrc(self.source);
                        this.success(mediaElement, originalNode, instance);
                        mediaElement.addEventListener(Hls.Events.MEDIA_ATTACHED, function () {
                        });
                    },
                    error: (e) => {
                        // console.log(e);
                    }
                });
            },
            remove() {
                if (this.player) {
                    this.player.pause();
                    this.player.remove();
                    this.player = null;
                }
            },
            setSource(src) {
                if (!this.player) {
                    console.log('source: player not initialized');
                    return;
                }

                if (!src) {
                    return;
                }

                this.player.height = this.height;
                this.player.width = this.width;
                this.player.videoHeight = this.height;
                this.player.videoWidth = this.width;
                this.$el.style.cssText = "width: " +  this.width + "px; height: " + this.height + "px;"

                this.player.setSrc(src);
                this.player.setPoster("");
                this.player.load();
            },
            pause() {
                if (this.player) {
                    this.player.pause();
                }
            },
        },
        beforeDestroy() {
            this.remove();
        },
        watch: {
            source: function (source) {
                if (source) {
                    this.setSource(source);
                }
            },
        },
    }
</script>
