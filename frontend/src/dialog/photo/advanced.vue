<template>
    <div class="p-tab p-tab-photo-advanced">
        <div class="v-table__overflow">
            <table class="v-datatable v-table theme--light">
                <tbody>
                <tr>
                    <td>UID</td>
                    <td>{{ model.UID | uppercase }}</td>
                </tr>
                <tr v-if="model.DocumentID">
                    <td>Document ID</td>
                    <td>{{ model.DocumentID | uppercase }}</td>
                </tr>
                <tr>
                    <td>Type</td>
                    <td>{{ model.Type | capitalize }}</td>
                </tr>
                <tr v-if="model.Path">
                    <td>Path</td>
                    <td>{{ model.Path }}</td>
                </tr>
                <tr>
                    <td>Name</td>
                    <td>{{ model.Name }}</td>
                </tr>
                <tr v-if="model.OriginalName">
                    <td>Original Name</td>
                    <td>
                        <v-text-field
                                @change="save"
                                flat solo dense hide-details v-model="model.OriginalName"
                                color="secondary-dark"
                                style="font-weight: 400; font-size: 13px;"
                        ></v-text-field>
                    </td>
                </tr>
                <tr>
                    <td>Title</td>
                    <td>{{ model.Title }}</td>
                </tr>
                <tr v-if="model.TitleSrc">
                    <td>Title Source</td>
                    <td>{{ model.TitleSrc | capitalize }}</td>
                </tr>
                <tr v-if="model.TakenAcc">
                    <td>Year</td>
                    <td>
                        <v-text-field
                                flat solo dense hide-details v-model="model.Year"
                                color="secondary-dark"
                                style="font-weight: 400; font-size: 13px;"
                        ></v-text-field>
                    </td>
                </tr>
                <tr v-if="model.TakenAcc">
                    <td>Month</td>
                    <td>
                        <v-select
                                label="Month"
                                flat solo dense hide-details
                                color="secondary-dark"
                                style="font-weight: 400; font-size: 13px;"
                                item-value="Month"
                                item-text="Name"
                                v-model="model.Month"
                                :items="monthOptions">
                        </v-select>
                    </td>
                </tr>
                <tr>
                    <td>Quality Score</td>
                    <td>
                        <v-rating v-model="model.Quality" :length="7" readonly small></v-rating>
                    </td>
                </tr>
                <tr>
                    <td>Resolution</td>
                    <td>{{ model.Resolution }} MP</td>
                </tr>
                <tr v-if="model.CameraSerial">
                    <td>Camera Serial</td>
                    <td>{{ model.CameraSerial }}
                    </td>
                </tr>
                <tr>
                    <td>Favorite</td>
                    <td>
                        <v-switch
                                @change="save"
                                hide-details
                                v-model="model.Favorite"
                                :label="model.Favorite ? 'Yes' : 'No'"
                        ></v-switch>
                    </td>
                </tr>
                <tr>
                    <td>Private</td>
                    <td>
                        <v-switch
                                @change="save"
                                hide-details
                                v-model="model.Private"
                                :label="model.Private ? 'Yes' : 'No'"
                        ></v-switch>
                    </td>
                </tr>
                <tr>
                    <td>Created</td>
                    <td>
                        {{ model.CreatedAt | luxon:format('http') }}
                    </td>
                </tr>
                <tr>
                    <td>Updated</td>
                    <td>
                        {{ model.UpdatedAt | luxon:format('http') }}
                    </td>
                </tr>
                <tr v-if="model.EditedAt">
                    <td>Edited</td>
                    <td>
                        {{ model.EditedAt | luxon:format('http') }}
                    </td>
                </tr>
                <tr v-if="model.MaintainedAt">
                    <td>Maintained</td>
                    <td>
                        {{ model.MaintainedAt | luxon:format('http') }}
                    </td>
                </tr>
                <tr v-if="model.DeletedAt">
                    <td>Archived</td>
                    <td>
                        {{ model.DeletedAt | luxon:format('http') }}
                    </td>
                </tr>
                </tbody>
            </table>
        </div>
    </div>
</template>

<script>
    import Thumb from "model/thumb";
    import {Info} from "luxon";

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
