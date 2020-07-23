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
                        :append-icon="model.TitleSrc === 'manual' ? 'check' : ''"
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

              <v-flex xs4 md2 pa-2>
                <v-autocomplete
                        :append-icon="model.TakenSrc === 'manual' ? 'check' : ''"
                        @change="updateTime"
                        :disabled="disabled"
                        :error="invalidDate"
                        :label="$gettext('Day')"
                        browser-autocomplete="off"
                        hide-details hide-no-data
                        color="secondary-dark"
                        v-model="model.Day"
                        :items="options.Days()"
                        class="input-day">
                </v-autocomplete>
              </v-flex>
              <v-flex xs4 md2 pa-2>
                <v-autocomplete
                        :append-icon="model.TakenSrc === 'manual' ? 'check' : ''"
                        @change="updateTime"
                        :disabled="disabled"
                        :error="invalidDate"
                        :label="$gettext('Month')"
                        browser-autocomplete="off"
                        hide-details hide-no-data
                        color="secondary-dark"
                        v-model="model.Month"
                        :items="options.MonthsShort()"
                        class="input-month">
                </v-autocomplete>
              </v-flex>
              <v-flex xs4 md2 pa-2>
                <v-autocomplete
                        :append-icon="model.TakenSrc === 'manual' ? 'check' : ''"
                        @change="updateTime"
                        :disabled="disabled"
                        :error="invalidDate"
                        :label="$gettext('Year')"
                        browser-autocomplete="off"
                        hide-details hide-no-data
                        color="secondary-dark"
                        v-model="model.Year"
                        :items="options.Years()"
                        class="input-year">
                </v-autocomplete>
              </v-flex>

              <v-flex xs6 md2 class="pa-2">
                <v-text-field
                        :append-icon="model.TakenSrc === 'manual' ? 'check' : ''"
                        :disabled="disabled"
                        @change="updateTime"
                        v-model="localTime"
                        :label="$gettext('Local Time')"
                        browser-autocomplete="off"
                        hide-details return-masked-value
                        mask="##:##:##"
                        color="secondary-dark"
                        class="input-local-time"
                ></v-text-field>
              </v-flex>

              <v-flex xs6 sm6 md2 pa-2>
                <v-text-field
                        :disabled="disabled"
                        @change="updateTime"
                        v-model="utcTime"
                        :label="$gettext('Time UTC')"
                        browser-autocomplete="off"
                        readonly hide-details return-masked-value
                        mask="##:##:##"
                        color="secondary-dark"
                        class="input-utc-time"
                ></v-text-field>
              </v-flex>

              <v-flex xs12 sm6 md2 class="pa-2">
                <v-autocomplete
                        @change="updateTime"
                        :disabled="disabled"
                        :label="labels.timezone"
                        browser-autocomplete="off"
                        hide-details hide-no-data
                        color="secondary-dark"
                        item-value="ID"
                        item-text="Name"
                        v-model="model.TimeZone"
                        :items="options.TimeZones()"
                        class="input-timezone">
                </v-autocomplete>
              </v-flex>

              <v-flex xs12 sm6 md4 class="pa-2">
                <v-autocomplete
                        :append-icon="model.PlaceSrc === 'manual' ? 'check' : ''"
                        :disabled="disabled"
                        :readonly="!!(model.Lat || model.Lng)"
                        :label="labels.country"
                        hide-details hide-no-data
                        browser-autocomplete="off"
                        color="secondary-dark"
                        item-value="Code"
                        item-text="Name"
                        v-model="model.Country"
                        :items="countries"
                        class="input-country">
                </v-autocomplete>
              </v-flex>

              <v-flex xs4 md2 class="pa-2">
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

              <v-flex xs4 md3 class="pa-2">
                <v-text-field
                        :append-icon="model.PlaceSrc === 'manual' ? 'check' : ''"
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

              <v-flex xs4 md3 class="pa-2">
                <v-text-field
                        :append-icon="model.PlaceSrc === 'manual' ? 'check' : ''"
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

              <v-flex xs12 md6 pa-2 class="p-camera-select">
                <v-select
                        :append-icon="model.CameraSrc === 'manual' ? 'check' : ''"
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

              <v-flex xs6 md3 class="pa-2">
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

              <v-flex xs6 md3 class="pa-2">
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
                        :append-icon="model.CameraSrc === 'manual' ? 'check' : ''"
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

              <v-flex xs6 md3 class="pa-2">
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

              <v-flex xs6 md3 class="pa-2">
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
                        :append-icon="model.DescriptionSrc === 'manual' ? 'check' : ''"
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
                       class="action-close">
                  <translate>Close</translate>
                </v-btn>
                <v-btn color="secondary-dark" depressed dark @click.stop="save(false)"
                       class="action-apply action-approve">
                  <span v-if="$config.feature('review') && model.Quality < 3"><translate>Approve</translate></span>
                  <span v-else><translate>Apply</translate></span>
                </v-btn>
                <v-btn color="secondary-dark" depressed dark @click.stop="save(true)"
                       class="action-done hidden-xs-only">
                  <translate>Done</translate>
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
    import countries from "options/countries.json";
    import Thumb from "model/thumb";
    import * as options from "options/options";

    export default {
        name: 'p-tab-photo-details',
        props: {
            model: Object,
            uid: String,
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
                    sort: this.$gettext("Sort Order"),
                    before: this.$gettext("Taken before"),
                    after: this.$gettext("Taken after"),
                    language: this.$gettext("Language"),
                    timezone: this.$gettext("Time Zone"),
                    title: this.$gettext("Title"),
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
                invalidDate: false,
                utcTime: "",
                localTime: "",
                textRule: v => v.length <= this.$config.get('clip') || this.$gettext("Text too long"),
            };
        },
        created() {
            this.updateTime();
        },
        watch: {
            model() {
                this.updateTime();
            },
            uid() {
                this.updateTime();
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
                if (!this.model.hasId()) {
                    return;
                }

                let localDate = this.model.localDate(this.localTime);

                this.invalidDate = !localDate.isValid

                if(this.invalidDate) {
                    return;
                }

                const utcDate = localDate.toUTC();

                this.localTime = localDate.toFormat("HH:mm:ss");
                this.utcTime = utcDate.toFormat("HH:mm:ss");

                if(this.model.Day === 0) {
                    this.model.Day = parseInt(localDate.toFormat("d"));
                }

                if(this.model.Month === 0) {
                    this.model.Month = parseInt(localDate.toFormat("L"));
                }

                if(this.model.Year === 0) {
                    this.model.Year = parseInt(localDate.toFormat("y"));
                }

                this.model.TakenAtLocal = localDate.toISO({
                    suppressMilliseconds: true,
                    includeOffset: false,
                }) + "Z";

                this.model.TakenAt = localDate.toUTC().toISO({
                    suppressMilliseconds: true,
                    includeOffset: false,
                }) + "Z";
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
            save(close) {
                if(this.invalidDate) {
                    this.$notify.error(this.$gettext("Invalid date"));
                    return;
                }

                this.model.update().then(() => {
                    if (close) {
                        this.$emit('close');
                    }

                    this.updateTime();
                });
            },
            close() {
                this.$emit('close');
            },
        },
    };
</script>
