<template>


    <div class="dashboard">
        <h1 class="md-display-1">Welcome to PhotoPrism</h1>
        <p class="md-subheading">
            A user-friendly tool for importing, filtering and archiving large amounts of JPEG and RAW files
        </p>
    </div>
</template>

<script>
    export default {
        name: 'dashboard',
        data() {
            return {
                msg: 'Welcome to PhotoPrism',
                username: this.$session.getFullName(),
            };
        },
        methods: {
            open() {
                this.$refs.snackbar.open();
            },

            showNotification() {
                this.$alert.success('Yeah wuwuwu');
            },

            logout() {
                this.$session.logout();
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
