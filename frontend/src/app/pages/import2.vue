<template>
    <div>
        <v-toolbar flat color="blue-grey lighten-4">
            <h1>Import</h1>
        </v-toolbar>
    <v-container fluid>
        <v-form>
        <p class="md-subheading">
            You have two possibilities to get your photos into photoprism.</p>
            <h2>Import & Index</h2>
            <p>Importing means the photos you upload are renamed (the naming schema you can define in settings), and moved to the originals folder sorted by year and month.
                Additionally duplicates are removed and images get tagged and metadata (like location, camera model etc.) will be extracted. In case you have not supported file types
                (e.g. videos) within the folder you import --> those are ignored. </p>
            <v-btn color="success" @click="$refs.inputUpload.click()" type="file" class="importbtn" disabled>Import & Index</v-btn>
            <input v-show="false" ref="inputUpload" type="file">
                <v-flex xs12 sm6 offset-sm3>
                    <v-card class="card">
                        <v-card-title primary-title>
                            <div>
                                <div>598 JPEG and 432 RAW files found</div>
                            </div>
                        </v-card-title>

                        <v-card-actions>
                            <v-btn flat color="success">Start</v-btn>
                            <v-btn flat color="success">Cancel</v-btn>
                        </v-card-actions>
                    </v-card>
                </v-flex>
        </v-form>
    </v-container>
        <v-container fluid>
            <v-form>
            <h2>Index</h2>
            <p>In case you already have a nice folder structure you can only index the photos. Therefore in settings you need to set the base directory to the directory your photos
            are in. The index functionality will then just tag the images and extract the metadata.
        </p>
             <v-btn color="success" @click="$refs.inputUpload.click()" type="file" class="importbtn" disabled>Index</v-btn>
          </v-form>
    </v-container>
</div>
</template>

<script>
    export default {
        name: 'import',
        props: {},
        data() {
            return {}
        },
    };
</script>
