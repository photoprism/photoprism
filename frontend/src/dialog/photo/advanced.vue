<template>
    <div class="p-tab p-tab-photo-advanced">
        <div class="v-table__overflow">
            <table class="v-datatable v-table theme--light">
                <tbody>
                <tr>
                    <td>UID</td>
                    <td>{{ model.UID | uppercase }}</td>
                </tr>
                <tr>
                    <td>Path</td>
                    <td>{{ model.Path }}</td>
                </tr>
                <tr>
                    <td>Name</td>
                    <td>{{ model.Name }}</td>
                </tr>
                <tr>
                    <td>Original</td>
                    <td><v-text-field
                            @change="save"
                            flat solo dense hide-details v-model="model.OriginalName"
                            color="secondary-dark"
                            style="font-weight: 400; font-size: 13px;"
                    ></v-text-field></td>
                </tr>
                <tr>
                    <td>Country</td>
                    <td>{{ model.countryName() }}</td>
                </tr>
                <tr>
                    <td>Year</td>
                    <td><v-text-field
                            @change="save"
                            flat solo dense hide-details v-model="model.Year"
                            color="secondary-dark"
                            style="font-weight: 400; font-size: 13px;"
                    ></v-text-field></td>
                </tr>
                <tr>
                    <td>Month</td>
                    <td><v-select @change="save"
                                  label="Month"
                                  flat solo dense hide-details
                                  color="secondary-dark"
                                  style="font-weight: 400; font-size: 13px;"
                                  item-value="Month"
                                  item-text="Name"
                                  v-model="model.Month"
                                  :items="monthOptions">
                    </v-select></td>
                </tr>
                </tbody>
            </table>
        </div>
    </div>
</template>

<script>
    import Thumb from "model/thumb";
    import {DateTime, Info} from "luxon";

    export default {
        name: 'p-tab-photo-advanced',
        props: {
            model: Object,
        },
        data() {
            return {
                config: this.$config.values,
                readonly: this.$config.get("readonly"),
            };
        },
        computed: {
            monthOptions() {
                let result = [
                    {"Month": -1, "Name": this.$gettext("Unknown")},
                ];

                const months = Info.months("long");

                for (let i = 0; i < months.length; i++) {
                    result.push({"Month": i + 1, "Name": months[i]});
                }

                return result;
            },
        },
        methods: {
            save() {
                this.model.update();
            },
            close() {
                this.$emit('close');
            },
            openPhoto() {
                this.$viewer.show(Thumb.fromFiles([this.model]), 0)
            },
        },
    };
</script>
