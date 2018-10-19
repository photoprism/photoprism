<template>
    <v-container fluid>
        <h1>Import</h1>
        <vue-dropzone
            ref="dropzone"
            id="dropzone"
            :options="dropzoneOptions"
            :useCustomSlot=true
            :duplicateCheck=true
            @vdropzone-queue-complete="onUploadComplete"
        >
            <div class="dropzone-custom-content">
                <h3 class="dropzone-custom-title">Drag and drop to upload your RAW and JPG files</h3>
            </div>
        </vue-dropzone>

        <v-btn ref="clearBtn" v-on:click="clearQueue" v-bind:disabled="btnDisable">Clear</v-btn>
        <v-btn ref="importBtn" color="info" v-on:click="importFiles" v-bind:disabled="btnDisable">Import</v-btn>
    </v-container>
</template>

<script>
    import vue2Dropzone from 'vue2-dropzone'
    import 'vue2-dropzone/dist/vue2Dropzone.min.css'

    export default {
        name: 'import',
        components: {
            vueDropzone: vue2Dropzone
        },
        data: function () {
            return {
                dropzoneOptions: {
                    url: 'http://localhost:2342/api/v1/photos',
                    autoProcessQueue: false,
                    thumbnailWidth: 200,
                    maxFilesize: 10.0,
                },
                btnDisable: false
            }
        },
        methods: {
            clearQueue: function (event) {
                this.$refs.dropzone.removeAllFiles()
            },
            importFiles: function (event) {
                this.$refs.dropzone.processQueue()
                this.btnDisable = true
            },
            onUploadComplete: function () {
                this.btnDisable = false
            }
        }
    };
</script>

<style scoped>

    h1, h2 {
        font-weight: normal;
    }

    ul {
        list-style-type: none;
        padding: 0;
    }

    li {
        display: inline-block;
        margin: 0 10px;
    }
</style>
