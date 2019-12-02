<template>
    <v-container fluid fill-height class="pa-0 ma-0 p-tab p-tab-logs">
        <v-layout row wrap fill-height class="pa-0 ma-2">
            <v-flex grow xs12 class="pa-2 terminal elevation-2 p-logs">
                <p v-for="(log, index) in logs" :key="index.id" class="p-log-message" :class="'p-log-' + log.level">
                    {{ log.time | luxon:format('yyyy-LL-dd hh:mm:ss') }} {{ level(log) }} <span>{{ log.msg }}</span>
                </p>
            </v-flex>
        </v-layout>
    </v-container>
</template>

<script>
    export default {
        name: 'p-tab-logs',
        data() {
            return {
                logs: this.$log.logs,
            };
        },
        methods: {
            level(log) {
                return log.level.substr(0, 4).toUpperCase();
            },
        },
        created() {
        },
    };
</script>
