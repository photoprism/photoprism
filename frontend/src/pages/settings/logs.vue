<template>
    <v-container fill-height class="p-tab p-tab-logs terminal elevation-2">
        <div class="p-logs">
            <p v-for="(log, index) in logs" :key="index.id" class="p-log-message" :class="'p-log-' + log.level">
                {{ log.time | luxon:format('yyyy-LL-dd hh:mm:ss') }} {{ level(log) }} <span>{{ log.msg }}</span>
            </p>
        </div>
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
