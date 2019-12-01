<template>
    <div class="p-tab p-tab-logs">
        <v-container fluid>
            <p>Work in progress</p>
        </v-container>
    </div>
</template>

<script>
    import Settings from "model/settings";
    import options from "resources/options.json";

    export default {
        name: 'p-tab-logs',
        data() {
            return {
            };
        },
        methods: {
            load() {
            },
        },
        created() {
            this.load();
        },
    };
</script>
