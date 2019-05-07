<template>
    <div>
        <v-toolbar flat color="blue-grey lighten-4">
            <v-toolbar-title>Settings</v-toolbar-title>

            <v-spacer></v-spacer>
        </v-toolbar>

        <v-container>
            <p>
                <v-btn v-on:click="showSuccess()">Show Success</v-btn>
                <v-btn v-on:click="showInfo()">Show Info</v-btn>
                <v-btn v-on:click="showError()">Show Error</v-btn>
                <v-btn v-on:click="showWarning()">Show Warning</v-btn>
            </p>
        </v-container>
    </div>
</template>

<script>
    export default {
        name: 'settings',
        data() {
            return {};
        },
        methods: {
            showSuccess() {
                this.$alert.success('Action successfully completed');
            },
            showInfo() {
                this.$alert.info('This is an info');
            },
            showError() {
                this.$alert.error('An error has occurred - please check logs');
            },
            showWarning() {
                this.$alert.warning('This is a warning');
            },
        }
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
