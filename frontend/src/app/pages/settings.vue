<template>
    <v-container fluid>
        <h1 class="md-display-1">Settings</h1>
        <p class="md-subheading">

        </p>
        <button v-on:click="showSuccess()">Show Success</button>
        <button v-on:click="showInfo()">Show Info</button>
        <button v-on:click="showError()">Show Error</button>
        <button v-on:click="showWarning()">Show Warning</button>
    </v-container>
</template>

<script>
    export default {
        name: 'dashboard',
        data() {
            return {
            };
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
