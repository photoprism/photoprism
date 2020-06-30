<template>
  <modal name="video" ref="video" :height="height" :width="width" :reset="true" class="p-video-dialog" @before-close="onClose"
         @before-open="onOpen">
    <p-video-player v-show="show" ref="player" :source="source" :height="height.toString()"
                    :width="width.toString()" :autoplay="true"></p-video-player>
  </modal>
</template>
<script>
    export default {
        name: 'p-video-dialog',
        data() {
            return {
                show: false,
                source: "",
                defaultWidth: 640,
                defaultHeight: 480,
                width: 640,
                height: 480,
                video: null,
                album: null,
            }
        },
        methods: {
            onOpen(ev) {
                this.video = ev.params.video;
                this.album = ev.params.album;
                this.play();
            },
            onClose() {
                this.$refs.player.pause();
                this.show = false;
            },
            play() {
                if (!this.video) {
                    this.$notify.error("no video selected");
                    return;
                }

                let main = this.video.mainFile();
                let file = this.video.videoFile();
                let uri = this.video.videoUrl();

                if (!uri) {
                    this.$notify.error("no video selected");
                    return;
                }

                const vw = Math.max(document.documentElement.clientWidth, window.innerWidth || 0);
                const vh = Math.max(document.documentElement.clientHeight, window.innerHeight || 0);

                let width = 0;
                let height = 0;

                if (file.Width > 0) {
                    width = file.Width;
                } else if (main && main.Width > 0) {
                    width = main.Width;
                } else {
                    width = this.defaultWidth;
                }

                if (file.Height > 0) {
                    height = file.Height;
                } else if (main && main.Height > 0) {
                    height = main.Height;
                } else {
                    height = this.defaultHeight;
                }

                this.width = width;
                this.height = height;

                if (vw < (width + 80)) {
                    let newWidth = vw - 120;
                    this.height = Math.round(newWidth * (height / width));
                    this.width = newWidth;
                }

                if (vh < (this.height + 100)) {
                    let newHeight = vh - 160;
                    this.width = Math.round(newHeight * (width / height));
                    this.height = newHeight;
                }

                // Resize video overlay.
                this.$refs.video.setInitialSize();
                let size = { width: this.width, height: this.height }
                this.$refs.video.onModalResize({size});

                // Play by triggering source change event.
                this.source = uri;
                this.show = true;
            },
        },
    }
</script>
