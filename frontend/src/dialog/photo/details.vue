<template>
  <div class="p-tab p-tab-photo-details">
    <v-container fluid>
      <v-form lazy-validation dense
              ref="form" class="p-form-photo-details-meta" accept-charset="UTF-8"
              @submit.prevent="save">
        <v-layout row wrap align-top fill-height>
          <v-flex
                  class="p-photo pa-2"
                  xs12 sm4 md2
          >
            <v-card tile
                    class="ma-1 elevation-0"
                    :title="model.Title">
              <v-img :src="model.thumbnailUrl('tile_500')"
                     aspect-ratio="1"
                     class="accent lighten-2 elevation-0 clickable"
                     @click.exact="openPhoto()"
                     v-touch="{left, right}"
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
                        :disabled="disabled"
                        :rules="[textRule]"
                        hide-details
                        :label="labels.title"
                        placeholder=""
                        color="secondary-dark"
                        browser-autocomplete="off"
                        v-model="model.Title"
                        class="input-title"
                ></v-text-field>
              </v-flex>

              <v-flex xs12 sm6 md3 pa-2 class="p-date-select">

                <v-text-field
                        :disabled="disabled"
                        :value="timeLocalFormatted"
                        browser-autocomplete="off"
                        :label="labels.localtime"
                        readonly
                        hide-details
                        color="secondary-dark"
                        class="input-local-time"
                ></v-text-field>

              </v-flex>
              <v-flex xs12 sm6 md3 pa-2 class="p-date-select">
                <v-text-field
                        :disabled="disabled"
                        @change="updateTime"
                        v-model="time"
                        :label="labels.utctime"
                        hide-details return-masked-value
                        mask="##:##:##"
                        color="secondary-dark"
                        class="input-utc-time"
                ></v-text-field>
              </v-flex>
              <v-flex xs12 sm6 md3 class="pa-2 p-date-select">
                <v-menu
                        :disabled="disabled"
                        :close-on-content-click="false"
                        full-width
                        max-width="290"
                >
                  <template v-slot:activator="{ on }">
                    <v-text-field
                            :disabled="disabled"
                            :value="dateFormatted"
                            browser-autocomplete="off"
                            :label="labels.utcdate"
                            readonly
                            hide-details
                            v-on="on"
                            color="secondary-dark"
                            class="input-utc-date"
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
                        @change="updateTime"
                        :disabled="disabled"
                        :label="labels.timezone"
                        hide-details
                        color="secondary-dark"
                        item-value="ID"
                        item-text="Name"
                        v-model="model.TimeZone"
                        :items="options.TimeZones()"
                        class="input-timezone">
                </v-autocomplete>
              </v-flex>

              <v-flex xs12 sm6 md3 class="pa-2">
                <v-text-field
                        :disabled="disabled"
                        hide-details
                        browser-autocomplete="off"
                        :label="labels.latitude"
                        placeholder=""
                        color="secondary-dark"
                        v-model="model.Lat"
                        class="input-latitude"
                ></v-text-field>
              </v-flex>

              <v-flex xs12 sm6 md3 class="pa-2">
                <v-text-field
                        :disabled="disabled"
                        hide-details
                        browser-autocomplete="off"
                        :label="labels.longitude"
                        placeholder=""
                        color="secondary-dark"
                        v-model="model.Lng"
                        class="input-longitude"
                ></v-text-field>
              </v-flex>

              <v-flex xs12 sm6 md3 class="pa-2">
                <v-text-field
                        :disabled="disabled"
                        hide-details
                        browser-autocomplete="off"
                        :label="labels.altitude"
                        placeholder=""
                        color="secondary-dark"
                        v-model="model.Altitude"
                        class="input-altitude"
                ></v-text-field>
              </v-flex>

              <v-flex xs12 sm6 md3 class="pa-2 p-countries-select">
                <v-autocomplete
                        :disabled="disabled"
                        :label="labels.country"
                        hide-details
                        browser-autocomplete="off"
                        color="secondary-dark"
                        item-value="Code"
                        item-text="Name"
                        v-model="model.Country"
                        :items="countries"
                        class="input-country">
                </v-autocomplete>
              </v-flex>

              <v-flex xs12 md6 pa-2 class="p-camera-select">
                <v-select
                        :disabled="disabled"
                        :label="labels.camera"
                        browser-autocomplete="off"
                        hide-details
                        color="secondary-dark"
                        item-value="ID"
                        item-text="Name"
                        v-model="model.CameraID"
                        :items="cameraOptions"
                        class="input-camera">
                </v-select>
              </v-flex>

              <v-flex xs12 sm6 md3 class="pa-2">
                <v-text-field
                        :disabled="disabled"
                        hide-details
                        browser-autocomplete="off"
                        label="ISO"
                        placeholder=""
                        color="secondary-dark"
                        v-model="model.Iso"
                        class="input-iso"
                ></v-text-field>
              </v-flex>

              <v-flex xs12 sm6 md3 class="pa-2">
                <v-text-field
                        :disabled="disabled"
                        hide-details
                        browser-autocomplete="off"
                        :label="labels.exposure"
                        placeholder=""
                        color="secondary-dark"
                        v-model="model.Exposure"
                        class="input-exposure"
                ></v-text-field>
              </v-flex>

              <v-flex xs12 md6 pa-2 class="p-lens-select">
                <v-select
                        :disabled="disabled"
                        :label="labels.lens"
                        browser-autocomplete="off"
                        hide-details
                        color="secondary-dark"
                        item-value="ID"
                        item-text="Name"
                        v-model="model.LensID"
                        :items="lensOptions"
                        class="input-lens">
                </v-select>
              </v-flex>

              <v-flex xs12 sm6 md3 class="pa-2">
                <v-text-field
                        :disabled="disabled"
                        hide-details
                        browser-autocomplete="off"
                        :label="labels.fnumber"
                        placeholder=""
                        color="secondary-dark"
                        v-model="model.FNumber"
                        class="input-fnumber"
                ></v-text-field>
              </v-flex>

              <v-flex xs12 sm6 md3 class="pa-2">
                <v-text-field
                        :disabled="disabled"
                        hide-details
                        browser-autocomplete="off"
                        :label="labels.focallength"
                        placeholder=""
                        color="secondary-dark"
                        v-model="model.FocalLength"
                        class="input-focal-length"
                ></v-text-field>
              </v-flex>

              <v-flex xs12 sm6 md3 class="pa-2">
                <v-textarea
                        :disabled="disabled"
                        :rules="[textRule]"
                        hide-details
                        browser-autocomplete="off"
                        auto-grow
                        :label="labels.subject"
                        placeholder=""
                        :rows="1"
                        color="secondary-dark"
                        v-model="model.Details.Subject"
                        class="input-subject"
                ></v-textarea>
              </v-flex>

              <v-flex xs12 sm6 md3 class="pa-2">
                <v-text-field
                        :disabled="disabled"
                        :rules="[textRule]"
                        hide-details
                        browser-autocomplete="off"
                        :label="labels.artist"
                        placeholder=""
                        color="secondary-dark"
                        v-model="model.Details.Artist"
                        class="input-artist"
                ></v-text-field>
              </v-flex>

              <v-flex xs12 sm6 md3 class="pa-2">
                <v-text-field
                        :disabled="disabled"
                        :rules="[textRule]"
                        hide-details
                        browser-autocomplete="off"
                        :label="labels.copyright"
                        placeholder=""
                        color="secondary-dark"
                        v-model="model.Details.Copyright"
                        class="input-copyright"
                ></v-text-field>
              </v-flex>

              <v-flex xs12 sm6 md3 class="pa-2">
                <v-textarea
                        :disabled="disabled"
                        :rules="[textRule]"
                        hide-details
                        browser-autocomplete="off"
                        auto-grow
                        :label="labels.license"
                        placeholder=""
                        :rows="1"
                        color="secondary-dark"
                        v-model="model.Details.License"
                        class="input-license"
                ></v-textarea>
              </v-flex>

              <v-flex xs12 class="pa-2">
                <v-textarea
                        :disabled="disabled"
                        hide-details
                        browser-autocomplete="off"
                        auto-grow
                        :label="labels.description"
                        placeholder=""
                        :rows="1"
                        color="secondary-dark"
                        v-model="model.Description"
                        class="input-description"
                ></v-textarea>
              </v-flex>

              <v-flex xs12 md6 class="pa-2">
                <v-textarea
                        :disabled="disabled"
                        hide-details
                        browser-autocomplete="off"
                        auto-grow
                        :label="labels.keywords"
                        placeholder=""
                        :rows="1"
                        color="secondary-dark"
                        v-model="model.Details.Keywords"
                        class="input-keywords"
                ></v-textarea>
              </v-flex>

              <v-flex xs12 md6 class="pa-2">
                <v-textarea
                        :disabled="disabled"
                        hide-details
                        browser-autocomplete="off"
                        auto-grow
                        :label="labels.notes"
                        placeholder=""
                        :rows="1"
                        color="secondary-dark"
                        v-model="model.Details.Notes"
                        class="input-notes"
                ></v-textarea>
              </v-flex>

              <v-flex xs12 text-xs-right class="pt-3" v-if="!disabled">
                <v-btn @click.stop="close" depressed color="secondary-light"
                       class="p-photo-dialog-close">
                  <translate key="Close">Close</translate>
                </v-btn>
                <v-btn color="secondary-dark" depressed dark @click.stop="save(false)"
                       class="p-photo-dialog-confirm action-approve">
                  <span v-if="$config.feature('review') && model.Quality < 3"><translate key="Approve">Approve</translate></span>
                  <span v-else><translate key="Apply">Apply</translate></span>
                </v-btn>
                <v-btn color="secondary-dark" depressed dark @click.stop="save(true)"
                       class="p-photo-dialog-confirm hidden-xs-only action-ok">
                  <span><translate key="OK">OK</translate></span>
                  <v-icon right dark>done</v-icon>
                </v-btn>
              </v-flex>
            </v-layout>
          </v-flex>
        </v-layout>

        <div class="mt-5"></div>
      </v-form>
    </v-container>
  </div>
</template>

<script>
    import {DateTime} from "luxon";
    import countries from "resources/countries.json";
    import Thumb from "model/thumb";
    import * as options from "resources/options";

    export default {
        name: 'p-tab-photo-details',
        props: {
            model: Object,
        },
        data() {
            return {
                disabled: !this.$config.feature("edit"),
                config: this.$config.values,
                all: {
                    colors: [{label: this.$gettext("Unknown"), name: ""}],
                },
                readonly: this.$config.get("readonly"),
                options: options,
                countries: countries,
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
                    title: this.$gettext("Title"),
                    localtime: this.$gettext("Local Time"),
                    utctime: this.$gettext("UTC Time"),
                    utcdate: this.$gettext("UTC Date"),
                    latitude: this.$gettext("Latitude"),
                    longitude: this.$gettext("Longitude"),
                    altitude: this.$gettext("Altitude (m)"),
                    exposure: this.$gettext("Exposure"),
                    fnumber: this.$gettext("F Number"),
                    focallength: this.$gettext("Focal Length"),
                    subject: this.$gettext("Subject"),
                    artist: this.$gettext("Artist"),
                    copyright: this.$gettext("Copyright"),
                    license: this.$gettext("License"),
                    description: this.$gettext("Description"),
                    keywords: this.$gettext("Keywords"),
                    notes: this.$gettext("Notes"),
                },
                showDatePicker: false,
                showTimePicker: false,
                date: "",
                time: "",
                dateFormatted: "",
                timeFormatted: "",
                timeLocalFormatted: "",
                textRule: v => v.length <= this.$config.get('clip') || this.$gettext("Text too long"),
            };
        },
        watch: {
            date() {
                if (!this.date) {
                    this.dateFormatted = "";
                    this.timeLocalFormatted = "";
                    return
                }

                this.dateFormatted = DateTime.fromISO(this.date).toLocaleString(DateTime.DATE_FULL);
            },
        },
        computed: {
            cameraOptions() {
                return this.config.cameras;
            },
            lensOptions() {
                return this.config.lenses;
            },
        },
        methods: {
            updateTime() {
                if (!this.time) {
                    this.time = DateTime.fromISO(this.model.TakenAt).toUTC().toFormat("HH:mm:ss");
                    return;
                }

                if (!this.date) {
                    return;
                }

                this.timeFormatted = DateTime.fromISO(this.time).toLocaleString(DateTime.TIME_24_WITH_SECONDS);

                const utcDate = this.date + "T" + this.time + "Z";

                this.model.TakenAt = utcDate;

                this.time = DateTime.fromISO(this.model.TakenAt).toUTC().toFormat("HH:mm:ss");

                let localDate = DateTime.fromISO(utcDate);

                if (this.model.TimeZone) {
                    localDate = localDate.setZone(this.model.TimeZone);
                } else {
                    localDate = localDate.toUTC(0);
                }

                this.model.TakenAtLocal = localDate.toISO({
                    suppressMilliseconds: true,
                    includeOffset: false,
                }) + "Z";

                this.timeLocalFormatted = localDate.toLocaleString(DateTime.TIME_24_WITH_SECONDS);
            },
            left() {
                this.$emit('next');
            },
            right() {
                this.$emit('prev');
            },
            openPhoto() {
                this.$viewer.show(Thumb.fromFiles([this.model]), 0)
            },
            refresh(model) {
                if (!model.hasId()) return;

                if (model.TakenAt) {
                    const date = DateTime.fromISO(model.TakenAt).toUTC();
                    this.date = date.toISODate();
                    this.time = date.toFormat("HH:mm:ss");

                    this.updateTime();
                }
            },
            save(close) {
                if (this.time && this.date) {
                    this.model.TakenAt = this.date + "T" + this.time + "Z";
                }

                this.model.update().then(() => {
                    if (close) {
                        this.$emit('close');
                    } else {
                        this.refresh(this.model);
                    }
                });
            },
            close() {
                this.$emit('close');
            },
        },
    };
</script>
