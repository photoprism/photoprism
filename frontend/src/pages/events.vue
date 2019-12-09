<template>
    <div>
        <v-toolbar flat color="secondary">
            <v-toolbar-title>Events</v-toolbar-title>

            <v-spacer></v-spacer>
        </v-toolbar>

        <v-container>
            <v-layout wrap>
                <v-flex
                        xs12
                        class="mb-3"
                >
                    <v-sheet height="500">
                        <v-calendar
                                ref="calendar"
                                v-model="start"
                                :type="type"
                                :end="end"
                                color="primary"
                        ></v-calendar>
                    </v-sheet>
                </v-flex>

                <v-flex
                        sm4
                        xs12
                        class="text-sm-left text-xs-center"
                >
                    <v-btn @click="$refs.calendar.prev()">
                        <v-icon
                                dark
                                left
                        >
                            keyboard_arrow_left
                        </v-icon>
                        <translate>Prev</translate>
                    </v-btn>
                </v-flex>
                <v-flex
                        sm4
                        xs12
                        class="text-xs-center"
                >
                    <v-select
                            v-model="type"
                            :items="typeOptions"
                            :label="labels.type"
                    ></v-select>
                </v-flex>
                <v-flex
                        sm4
                        xs12
                        class="text-sm-right text-xs-center"
                >
                    <v-btn @click="$refs.calendar.next()">
                        <translate>Next</translate>
                        <v-icon
                                right
                                dark
                        >
                            keyboard_arrow_right
                        </v-icon>
                    </v-btn>
                </v-flex>
            </v-layout>
        </v-container>
    </div>
</template>

<script>
    export default {
        name: 'calendar',
        data: () => ({
            type: 'month',
            start: '2019-01-01',
            end: '2019-01-06',
            typeOptions: [
                {text: this.$gettext('Day'), value: 'day'},
                {text: this.$gettext('4 Day'), value: '4day'},
                {text: this.$gettext('Week'), value: 'week'},
                {text: this.$gettext('Month'), value: 'month'},
                {text: this.$gettext('Custom Daily'), value: 'custom-daily'},
                {text: this.$gettext('Custom Weekly'), value: 'custom-weekly'}
            ],
            labels: {
                type: this.$gettext("Type"),
            },
        }),
        methods: {}
    };
</script>
