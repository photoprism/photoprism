<template>
    <v-dialog lazy v-model="show" :scrollable="false" :max-width="width" class="p-video-dialog">
        <p-video-player v-show="show" ref="player" :source="source" :height="height.toString()" :width="width.toString()" :autoplay="true"></p-video-player>
    </v-dialog>
</template>
<script>
    export default {
        name: 'p-video-dialog',
        props: {
            play: Object,
            album: Object,
        },
        data() {
            return {
                show: false,
                source: "",
                defaultWidth: 640,
                defaultHeight: 480,
                width: 640,
                height: 480,
            }
        },
        methods: {
            load(video) {
                if (!video) {
                    this.$notify.error("no video selected");
                    return;
                }

                let main = video.mainFile();
                let file = video.videoFile();
                let uri = video.videoUri();

                if (!uri) {
                    this.$notify.error("no video selected");
                    return;
                }

                const vw = Math.max(document.documentElement.clientWidth, window.innerWidth || 0);
                const vh = Math.max(document.documentElement.clientHeight, window.innerHeight || 0);

                let width = 0;
                let height = 0;

                if(file.FileWidth > 0) {
                    width = file.FileWidth;
                } else if(main && main.FileWidth > 0) {
                    width = main.FileWidth;
                } else {
                    width = this.defaultWidth;
                }

                if(file.FileHeight > 0) {
                    height = file.FileHeight;
                } else if(main && main.FileHeight > 0) {
                    height = main.FileHeight;
                } else {
                    height = this.defaultHeight;
                }

                this.width = width;
                this.height = height;

                if(vw < (width + 80)) {
                    let newWidth =  vw - 120;
                    this.height = Math.round(newWidth * (height / width));
                    this.width = newWidth;
                }

                if(vh < (this.height + 100)) {
                    let newHeight = vh - 160;
                    this.width = Math.round(newHeight * (width / height));
                    this.height = newHeight;
                }

                this.source = uri;
                this.show = true;
            },
        },
        watch: {
            play: function (play) {
                if (play) {
                    this.load(play);
                }
            },
            show: function(show) {
                if(!show) {
                    this.$refs.player.pause();
                }
            }
        },
    }
</script>
