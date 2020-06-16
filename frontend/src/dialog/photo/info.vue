<template>
    <div class="p-tab p-tab-photo-advanced">
        <div class="v-table__overflow">
            <table class="v-datatable v-table theme--light">
                <tbody>
                <tr>
                    <td><translate key="UID">UID</translate></td>
                    <td>{{ model.UID | uppercase }}</td>
                </tr>
                <tr v-if="model.DocumentID">
                    <td><translate key="Document ID">Document ID</translate></td>
                    <td>{{ model.DocumentID | uppercase }}</td>
                </tr>
                <tr>
                    <td><translate key="Type">Type</translate></td>
                    <td>{{ model.Type | capitalize }}</td>
                </tr>
                <tr v-if="model.Path">
                    <td><translate key="Path">Path</translate></td>
                    <td>{{ model.Path }}</td>
                </tr>
                <tr>
                    <td><translate key="Name">Name</translate></td>
                    <td>{{ model.Name }}</td>
                </tr>
                <tr v-if="model.OriginalName">
                    <td><translate key="Original Name">Original Name</translate></td>
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
                    <td><translate key="Title">Title</translate></td>
                    <td>{{ model.Title }}</td>
                </tr>
                <tr v-if="model.TitleSrc">
                    <td><translate key="Title Source">Title Source</translate></td>
                    <td>{{ model.TitleSrc | capitalize }}</td>
                </tr>
                <tr v-if="model.TakenAcc">
                    <td><translate key="Year">Year</translate></td>
                    <td>
                        <v-text-field
                                flat solo dense hide-details v-model="model.Year"
                                color="secondary-dark"
                                style="font-weight: 400; font-size: 13px;"
                        ></v-text-field>
                    </td>
                </tr>
                <tr v-if="model.TakenAcc">
                    <td><translate key="Month">Month</translate></td>
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
                    <td><translate key="Quality Score">Quality Score</translate></td>
                    <td>
                        <v-rating v-model="model.Quality" :length="7" readonly small></v-rating>
                    </td>
                </tr>
                <tr>
                    <td><translate key="Resolution">Resolution</translate></td>
                    <td>{{ model.Resolution }} MP</td>
                </tr>
                <tr v-if="model.CameraSerial">
                    <td><translate key="Camera Serial">Camera Serial</translate></td>
                    <td>{{ model.CameraSerial }}
                    </td>
                </tr>
                <tr>
                    <td><translate key="Favorite">Favorite</translate></td>
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
                    <td><translate key="Private">Private</translate></td>
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
                    <td><translate key="Created">Created</translate></td>
                    <td>
                        {{ model.CreatedAt | luxon:format('http') }}
                    </td>
                </tr>
                <tr>
                    <td><translate key="Updated">Updated</translate></td>
                    <td>
                        {{ model.UpdatedAt | luxon:format('http') }}
                    </td>
                </tr>
                <tr v-if="model.EditedAt">
                    <td><translate key="Edited">Edited</translate></td>
                    <td>
                        {{ model.EditedAt | luxon:format('http') }}
                    </td>
                </tr>
                <tr v-if="model.MaintainedAt">
                    <td><translate key="Maintained">Maintained</translate></td>
                    <td>
                        {{ model.MaintainedAt | luxon:format('http') }}
                    </td>
                </tr>
                <tr v-if="model.DeletedAt">
                    <td><translate key="Archived">Archived</translate></td>
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
