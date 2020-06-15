<template>
    <div class="p-tab p-tab-photo-advanced">
        <div class="v-table__overflow">
            <table class="v-datatable v-table theme--light">
                <tbody>
                <tr>
                    <td><translate>UID</translate></td>
                    <td>{{ model.UID | uppercase }}</td>
                </tr>
                <tr v-if="model.DocumentID">
                    <td><translate>Document ID</translate></td>
                    <td>{{ model.DocumentID | uppercase }}</td>
                </tr>
                <tr>
                    <td><translate>Type</translate></td>
                    <td>{{ model.Type | capitalize }}</td>
                </tr>
                <tr v-if="model.Path">
                    <td><translate>Path</translate></td>
                    <td>{{ model.Path }}</td>
                </tr>
                <tr>
                    <td><translate>Name</translate></td>
                    <td>{{ model.Name }}</td>
                </tr>
                <tr v-if="model.OriginalName">
                    <td><translate>Original Name</translate></td>
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
                    <td><translate>Title</translate></td>
                    <td>{{ model.Title }}</td>
                </tr>
                <tr v-if="model.TitleSrc">
                    <td><translate>Title Source</translate></td>
                    <td>{{ model.TitleSrc | capitalize }}</td>
                </tr>
                <tr v-if="model.TakenAcc">
                    <td><translate>Year</translate></td>
                    <td>
                        <v-text-field
                                flat solo dense hide-details v-model="model.Year"
                                color="secondary-dark"
                                style="font-weight: 400; font-size: 13px;"
                        ></v-text-field>
                    </td>
                </tr>
                <tr v-if="model.TakenAcc">
                    <td><translate>Month</translate></td>
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
                    <td><translate>Quality Score</translate></td>
                    <td>
                        <v-rating v-model="model.Quality" :length="7" readonly small></v-rating>
                    </td>
                </tr>
                <tr>
                    <td><translate>Resolution</translate></td>
                    <td>{{ model.Resolution }} MP</td>
                </tr>
                <tr v-if="model.CameraSerial">
                    <td><translate>Camera Serial</translate></td>
                    <td>{{ model.CameraSerial }}
                    </td>
                </tr>
                <tr>
                    <td><translate>Favorite</translate></td>
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
                    <td><translate>Private</translate></td>
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
                    <td><translate>Created</translate></td>
                    <td>
                        {{ model.CreatedAt | luxon:format('http') }}
                    </td>
                </tr>
                <tr>
                    <td><translate>Updated</translate></td>
                    <td>
                        {{ model.UpdatedAt | luxon:format('http') }}
                    </td>
                </tr>
                <tr v-if="model.EditedAt">
                    <td><translate>Edited</translate></td>
                    <td>
                        {{ model.EditedAt | luxon:format('http') }}
                    </td>
                </tr>
                <tr v-if="model.MaintainedAt">
                    <td><translate>Maintained</translate></td>
                    <td>
                        {{ model.MaintainedAt | luxon:format('http') }}
                    </td>
                </tr>
                <tr v-if="model.DeletedAt">
                    <td><translate>Archived</translate></td>
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
