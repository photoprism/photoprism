<template>
    <div class="p-tab p-tab-photo-edit-details">
        <v-container fluid>
            <v-form lazy-validation dense
                    ref="form" class="p-form-photo-edit-meta" accept-charset="UTF-8"
                    @submit.prevent="save">
                <v-layout row wrap align-top fill-height>
                    <v-flex
                            class="p-photo pa-2"
                            xs12 sm4 md2
                    >
                        <v-card tile
                                class="ma-1 elevation-0"
                                :title="model.PhotoTitle">
                            <v-img :src="model.getThumbnailUrl('tile_500')"
                                   aspect-ratio="1"
                                   class="accent lighten-2 elevation-0"
                                   style="cursor: pointer"
                                   @click.exact="openPhoto()"
                            >
                                <v-layout
                                        slot="placeholder"
                                        fill-height
                                        align-center
                                        justify-center
                                        ma-0
                                >
                                    <v-progress-circular indeterminate
                                                         color="accent lighten-5"></v-progress-circular>
                                </v-layout>
                            </v-img>

                        </v-card>
                    </v-flex>
                    <v-flex xs12 sm8 md10 fill-height>
                        <v-layout row wrap>
                            <v-flex xs12 class="pa-2">
                                <v-text-field
                                        hide-details
                                        label="Title"
                                        placeholder=""
                                        color="secondary-dark"
                                        v-model="model.PhotoTitle"
                                ></v-text-field>
                            </v-flex>

                            <v-flex xs12 sm6 md3 pa-2 class="p-date-select">

                                <v-text-field
                                        :value="timeLocalFormatted"
                                        label="Local Time"
                                        readonly
                                        hide-details
                                        color="secondary-dark"
                                ></v-text-field>

                            </v-flex>
                            <v-flex xs12 sm6 md3 pa-2 class="p-date-select">
                                <v-menu
                                        v-model="showTimePicker"
                                        :close-on-content-click="false"
                                        full-width
                                        max-width="290"
                                >
                                    <template v-slot:activator="{ on }">
                                        <v-text-field
                                                :value="timeFormatted"
                                                label="UTC Time"
                                                readonly
                                                hide-details
                                                v-on="on"
                                                color="secondary-dark"
                                        ></v-text-field>
                                    </template>
                                    <v-time-picker
                                            color="secondary-dark"
                                            v-model="time"
                                            format="24hr"
                                            use-seconds
                                            @change="showTimePicker = false"
                                    ></v-time-picker>
                                </v-menu>
                            </v-flex>
                            <v-flex xs12 sm6 md3 class="pa-2 p-date-select">
                                <v-menu
                                        :close-on-content-click="false"
                                        full-width
                                        max-width="290"
                                >
                                    <template v-slot:activator="{ on }">
                                        <v-text-field
                                                :value="dateFormatted"
                                                label="UTC Date"
                                                readonly
                                                hide-details
                                                v-on="on"
                                                color="secondary-dark"
                                        ></v-text-field>
                                    </template>
                                    <v-date-picker
                                            color="secondary-dark"
                                            v-model="date"
                                            @change="showDatePicker = false"
                                    ></v-date-picker>
                                </v-menu>
                            </v-flex>

                            <v-flex xs12 sm6 md3 class="pa-2 p-timezone-select">
                                <v-autocomplete
                                        :label="labels.timezone"
                                        hide-details
                                        color="secondary-dark"
                                        item-value="code"
                                        item-text="name"
                                        v-model="model.TimeZone"
                                        :items="timeZones">
                                </v-autocomplete>
                            </v-flex>

                            <v-flex xs12 md6 pa-2 class="p-camera-select">
                                <v-select
                                        :label="labels.camera"
                                        hide-details
                                        color="secondary-dark"
                                        item-value="ID"
                                        item-text="CameraModel"
                                        v-model="model.CameraID"
                                        :items="cameraOptions">
                                </v-select>
                            </v-flex>
                            <v-flex xs12 md6 pa-2 class="p-lens-select">
                                <v-select
                                        :label="labels.lens"
                                        hide-details
                                        color="secondary-dark"
                                        item-value="ID"
                                        item-text="LensModel"
                                        v-model="model.LensID"
                                        :items="lensOptions">
                                </v-select>
                            </v-flex>

                            <v-flex xs12 sm6 md3 class="pa-2">
                                <v-text-field
                                        hide-details
                                        label="Latitude"
                                        placeholder=""
                                        color="secondary-dark"
                                        v-model="model.PhotoLat"
                                ></v-text-field>
                            </v-flex>
                            <v-flex xs12 sm6 md3 class="pa-2">
                                <v-text-field
                                        hide-details
                                        label="Longitude"
                                        placeholder=""
                                        color="secondary-dark"
                                        v-model="model.PhotoLng"
                                ></v-text-field>
                            </v-flex>
                            <v-flex xs12 sm6 md3 class="pa-2">
                                <v-text-field
                                        hide-details
                                        label="Altitude (m)"
                                        placeholder=""
                                        color="secondary-dark"
                                        v-model="model.PhotoAltitude"
                                ></v-text-field>
                            </v-flex>

                            <v-flex xs12 sm6 md3 class="pa-2 p-countries-select">
                                <v-select
                                        :label="labels.country"
                                        hide-details
                                        color="secondary-dark"
                                        item-value="code"
                                        item-text="name"
                                        v-model="model.PhotoCountry"
                                        :items="countryOptions">
                                </v-select>
                            </v-flex>

                            <v-flex xs12 sm6 md3 class="pa-2">
                                <v-text-field
                                        hide-details
                                        label="Focal Length"
                                        placeholder=""
                                        color="secondary-dark"
                                        v-model="model.PhotoFocalLength"
                                ></v-text-field>
                            </v-flex>

                            <v-flex xs12 sm6 md3 class="pa-2">
                                <v-text-field
                                        hide-details
                                        label="F Number"
                                        placeholder=""
                                        color="secondary-dark"
                                        v-model="model.PhotoFNumber"
                                ></v-text-field>
                            </v-flex>

                            <v-flex xs12 sm6 md3 class="pa-2">
                                <v-text-field
                                        hide-details
                                        label="ISO"
                                        placeholder=""
                                        color="secondary-dark"
                                        v-model="model.PhotoIso"
                                ></v-text-field>
                            </v-flex>

                            <v-flex xs12 sm6 md3 class="pa-2">
                                <v-text-field
                                        hide-details
                                        label="Exposure"
                                        placeholder=""
                                        color="secondary-dark"
                                        v-model="model.PhotoExposure"
                                ></v-text-field>
                            </v-flex>

                            <v-flex xs12 class="pa-2">
                                <v-textarea
                                        hide-details
                                        auto-grow
                                        label="Description"
                                        placeholder=""
                                        :rows="1"
                                        color="secondary-dark"
                                        v-model="model.PhotoDescription"
                                ></v-textarea>
                            </v-flex>

                            <v-flex xs12 sm6 md3 class="pa-2">
                                <v-text-field
                                        hide-details
                                        label="Copyright"
                                        placeholder=""
                                        color="secondary-dark"
                                        v-model="model.PhotoCopyright"
                                ></v-text-field>
                            </v-flex>

                            <v-flex xs12 sm6 md3 class="pa-2">
                                <v-text-field
                                        hide-details
                                        label="Artist"
                                        placeholder=""
                                        color="secondary-dark"
                                        v-model="model.PhotoArtist"
                                ></v-text-field>
                            </v-flex>

                            <v-flex xs12 md6 class="pa-2">
                                <v-textarea
                                        hide-details
                                        auto-grow
                                        label="Notes"
                                        placeholder=""
                                        :rows="1"
                                        color="secondary-dark"
                                        v-model="model.PhotoNotes"
                                ></v-textarea>
                            </v-flex>

                            <v-flex xs12 text-xs-right class="pt-3">
                                <v-btn @click.stop="close" depressed color="secondary-light"
                                       class="p-photo-dialog-close">
                                    <translate>Close</translate>
                                </v-btn>
                                <v-btn color="secondary-dark" depressed dark @click.stop="save"
                                       class="p-photo-dialog-confirm">
                                    <span>Save</span>
                                    <v-icon right dark>save</v-icon>
                                </v-btn>
                            </v-flex>
                        </v-layout>
                    </v-flex>
                </v-layout>
            </v-form>
        </v-container>
    </div>
</template>

<script>
    import options from "resources/options.json";
    import {DateTime} from "luxon";
    import moment from "moment-timezone"

    export default {
        name: 'p-tab-photo-edit-details',
        props: {
            model: Object,
        },
        data() {
            return {
                config: this.$config.values,
                all: {
                    countries: [{code: "", name: this.$gettext("Unknown")}],
                    cameras: [{ID: 0, CameraModel: this.$gettext("Unknown")}],
                    lenses: [{ID: 0, LensModel: "Unknown"}],
                    colors: [{label: "Unknown", name: ""}],
                },
                readonly: this.$config.getValue("readonly"),
                options: options,
                labels: {
                    search: this.$gettext("Search"),
                    view: this.$gettext("View"),
                    country: this.$gettext("Country"),
                    camera: this.$gettext("Camera"),
                    lens: this.$gettext("Lens"),
                    year: this.$gettext("Year"),
                    color: this.$gettext("Color"),
                    category: this.$gettext("Category"),
                    sort: this.$gettext("Sort By"),
                    before: this.$gettext("Taken before"),
                    after: this.$gettext("Taken after"),
                    language: this.$gettext("Language"),
                    timezone: this.$gettext("Time Zone"),
                },
                showDatePicker: false,
                showTimePicker: false,
                date: "",
                time: "",
            };
        },
        computed: {
            dateFormatted() {
                if (!this.date) {
                    return "";
                }

                return DateTime.fromISO(this.date).toLocaleString(DateTime.DATE_FULL);
            },
            timeFormatted() {
                if (!this.time) {
                    return "";
                }

                return DateTime.fromISO(this.time).toLocaleString(DateTime.TIME_24_WITH_SECONDS);
            },
            timeLocalFormatted() {
                if (!this.time || !this.date) {
                    return ""
                }

                const utcDate = this.date + "T" + this.time + "Z";

                this.model.TakenAt = utcDate;

                let localDate = DateTime.fromISO(utcDate);

                if(this.model.TimeZone) {
                    localDate = localDate.setZone(this.model.TimeZone);
                } else {
                    localDate = localDate.toUTC(0);
                }

                this.model.TakenAtLocal = localDate.toISO({
                    suppressMilliseconds: true,
                    includeOffset: false,
                }) + "Z";

                return localDate.toLocaleString(DateTime.TIME_24_WITH_SECONDS);
            },
            countryOptions() {
                return this.all.countries.concat(this.config.countries);
            },
            cameraOptions() {
                return this.all.cameras.concat(this.config.cameras);
            },
            lensOptions() {
                return this.all.lenses.concat(this.config.lenses);
            },
            colorOptions() {
                return this.all.colors.concat(this.config.colors);
            },
            timeZones() {
                return moment.tz.names();
            },
        },
        methods: {
            openPhoto() {
                this.$viewer.show([this.model], 0)
            },
            refresh(model) {
                if (!model.hasId()) return;

                if (model.TakenAt) {
                    const date = DateTime.fromISO(model.TakenAt).toUTC();
                    this.date = date.toISODate();
                    this.time = date.toFormat("HH:mm:ss");
                }
            },
            save() {
                if (this.time && this.date) {
                    this.model.TakenAt = this.date + "T" + this.time + "Z";
                }

                this.model.update();
            },
            close() {
                this.$emit('cancel');
            },
        },
    };
</script>
