<template>
    <div class="p-tab p-tab-logs p-logs">
        <v-container fluid>
            <v-list dense class="transparent">
                <template v-for="(log, index) in logs">
                    <v-list-tile
                            :key="index.id"
                            avatar
                            ripple
                            :class="logClass(log.level)"
                    >
                        <v-list-tile-content>
                            <v-list-tile-sub-title class="p-log-message black--text">{{ log.msg }}</v-list-tile-sub-title>
                        </v-list-tile-content>

                        <v-list-tile-action>
                            <v-list-tile-action-text class="p-log-time">{{ log.time | luxon:format('MMM d, yyyy hh:mm:ss') }}</v-list-tile-action-text>
                        </v-list-tile-action>
                    </v-list-tile>

                    <v-divider
                            v-if="index + 1 < logs.length"
                            :key="index.id"
                    ></v-divider>
                </template>
            </v-list>
        </v-container>
    </div>
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
            logClass(level) {
                switch (level) {
                    case "fatal":
                    case "critical":
                    case "error": return "pink lighten-5";
                    case "warning": return "amber lighten-4";
                    default:
                        return "blue-grey lighten-5";
                }
            },
        },
        created() {
        },
    };
</script>
