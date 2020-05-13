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
                    this.$notify.error("no video file found");
                    return;
                }

                if(file.FileWidth > 0) {
                    this.width = file.FileWidth;
                } else if(main.FileWidth > 0) {
                    this.width = main.FileWidth;
                } else {
                    this.width = this.defaultWidth;
                }

                if(window.innerWidth < (this.width + 50)) {
                    this.width = window.innerWidth - 50;
                }

                if(file.FileHeight > 0) {
                    this.height = file.FileHeight;
                } else if(main.FileHeight > 0) {
                    this.height = main.FileHeight;
                } else {
                    this.height = this.defaultHeight;
                }

                this.$el.style.height = this.height;
                this.$el.style.width = this.width;

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
