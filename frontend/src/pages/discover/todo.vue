<template>
    <div class="p-tab p-tab-discover-todo">
        <v-container grid-list-xs text-xs-center fluid>
            <p class="subheading pb-3">
                This is a very first draft for a "Discover" area where you can find photos by
                similarity, by color, by season or just randomly. Feedback welcome.
            </p>
        </v-container>
    </div>
</template>

<script>

    export default {
        name: 'p-tab-discover-todo',
        data() {
            return {
                readonly: this.$config.getValue("readonly"),
                config: this.$config.values,
                labels: {},
            };
        },
        methods: {
        },
        created() {
        },
    };
</script>
