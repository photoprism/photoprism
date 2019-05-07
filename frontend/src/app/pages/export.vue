<template>
    <div>
        <v-toolbar flat color="blue-grey lighten-4">
            <v-toolbar-title>Export</v-toolbar-title>

            <v-spacer></v-spacer>
        </v-toolbar>

        <v-container>
            <p>
                Issues labeled <a href="https://github.com/photoprism/photoprism/labels/help%20wanted">help wanted</a> /
                <a href="https://github.com/photoprism/photoprism/labels/easy">easy</a> can be good (first)
                contributions.
                Our <a href="https://github.com/photoprism/photoprism/wiki">Developer Guide</a> contains all information
                necessary to get you started.
            </p>
        </v-container>
    </div>
</template>

<script>
    export default {
        name: 'todo',
        data() {
            return {};
        },
        methods: {}
    };
</script>

<style scoped>
    a {
        color: #00B8D4;
    }

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
