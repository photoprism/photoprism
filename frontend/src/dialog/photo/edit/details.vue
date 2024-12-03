<template>
  <div class="p-tab p-tab-photo-details">
    <v-form ref="form" lazy-validation class="p-form-photo-details-meta" accept-charset="UTF-8" @submit.prevent="save">
      <v-row class="pa-2 d-flex align-stretch" align="start">
        <v-col class="pa-2 p-photo d-flex align-stretch" cols="12" sm="4" md="2">
          <v-card tile color="background" class="pa-0 ma-0 elevation-0 flex-grow-1">
            <v-img v-touch="{ left, right }" :src="model.thumbnailUrl('tile_500')" aspect-ratio="1" class="card elevation-0 clickable" @click.exact="openPhoto()">
</v-img>
          </v-card>
        </v-col>
        <v-col cols="12" sm="8" md="10" class="d-flex pa-0" align-self="stretch">
          <v-row>
            <v-col cols="12" lg="6" class="pa-2">
              <v-text-field
                v-model="model.Title"
                :append-icon="model.TitleSrc === 'mdi-human-male' ? 'mdi-check' : ''"
                :disabled="disabled"
                :rules="[textRule]"
                hide-details
                :label="$pgettext('Photo', 'Title')"
                placeholder=""
                autocomplete="off"
                class="input-title"
              ></v-text-field>
            </v-col>

            <v-col cols="4" lg="1" class="pa-2">
              <v-autocomplete
                v-model="model.Day"
                :append-icon="model.TakenSrc === 'mdi-human-male' ? 'mdi-check' : ''"
                :disabled="disabled"
                :error="invalidDate"
                :label="$gettext('Day')"
                autocomplete="off"
                hide-details
                hide-no-data
                :items="options.Days()"
                item-title="text"
                item-value="value"
                class="input-day"
                @update:model-value="updateTime"
              >
              </v-autocomplete>
            </v-col>
            <v-col cols="4" lg="1" class="pa-2">
              <v-autocomplete
                v-model="model.Month"
                :append-icon="model.TakenSrc === 'mdi-human-male' ? 'mdi-check' : ''"
                :disabled="disabled"
                :error="invalidDate"
                :label="$gettext('Month')"
                autocomplete="off"
                hide-details
                hide-no-data
                :items="options.MonthsShort()"
                item-title="text"
                item-value="value"
                class="input-month"
                @update:model-value="updateTime"
              >
              </v-autocomplete>
            </v-col>
            <v-col cols="4" lg="2" class="pa-2">
              <v-autocomplete
                v-model="model.Year"
                :append-icon="model.TakenSrc === 'mdi-human-male' ? 'mdi-check' : ''"
                :disabled="disabled"
                :error="invalidDate"
                :label="$gettext('Year')"
                autocomplete="off"
                hide-details
                hide-no-data
                :items="options.Years()"
                item-title="text"
                item-value="value"
                class="input-year"
                @update:model-value="updateTime"
              >
              </v-autocomplete>
            </v-col>

            <v-col cols="6" lg="2" class="pa-2">
              <!-- TODO: check property return-masked-value TEST -->
              <v-text-field
                v-model="time"
                :append-icon="model.TakenSrc === 'mdi-human-male' ? 'mdi-check' : ''"
                :disabled="disabled"
                :label="model.timeIsUTC() ? $gettext('Time UTC') : $gettext('Local Time')"
                autocomplete="off"
                autocorrect="off"
                autocapitalize="none"
                hide-details
                return-masked-value
                mask="##:##:##"
                class="input-local-time"
              ></v-text-field>
            </v-col>

            <v-col cols="6" sm="6" md="6" lg="3" class="pa-2">
              <v-autocomplete v-model="model.TimeZone" :disabled="disabled" :label="$gettext('Time Zone')" hide-details flat hide-no-data color="surface-variant" item-value="ID" item-title="Name" :items="options.TimeZones()" class="input-timezone" @update:model-value="updateTime"> </v-autocomplete>
            </v-col>

            <v-col cols="12" sm="8" md="4" lg="3" class="pa-2">
              <v-autocomplete
                v-model="model.Country"
                :append-icon="model.PlaceSrc === 'mdi-human-male' ? 'mdi-check' : ''"
                :disabled="disabled"
                :readonly="!!(model.Lat || model.Lng)"
                :label="$gettext('Country')"
                hide-details
                hide-no-data
                autocomplete="off"
                item-value="Code"
                item-title="Name"
                :items="countries"
                class="input-country"
              >
              </v-autocomplete>
            </v-col>

            <v-col cols="4" md="2" lg="2" class="pa-2">
              <v-text-field v-model="model.Altitude" :disabled="disabled" hide-details flat autocomplete="off" autocorrect="off" autocapitalize="none" :label="$gettext('Altitude (m)')" placeholder="" color="surface-variant" class="input-altitude"></v-text-field>
            </v-col>

            <v-col cols="4" sm="6" md="3" lg="2" class="pa-2">
              <v-text-field
                v-model="model.Lat"
                :append-icon="model.PlaceSrc === 'mdi-human-male' ? 'mdi-check' : ''"
                :disabled="disabled"
                hide-details
                autocomplete="off"
                autocorrect="off"
                autocapitalize="none"
                :label="$gettext('Latitude')"
                placeholder=""
                class="input-latitude"
                @paste="pastePosition"
              ></v-text-field>
            </v-col>

            <v-col cols="4" sm="6" md="3" lg="2" class="pa-2">
              <v-text-field
                v-model="model.Lng"
                :append-icon="model.PlaceSrc === 'mdi-human-male' ? 'mdi-check' : ''"
                :disabled="disabled"
                hide-details
                autocomplete="off"
                autocorrect="off"
                autocapitalize="none"
                :label="$gettext('Longitude')"
                placeholder=""
                class="input-longitude"
                @paste="pastePosition"
              ></v-text-field>
            </v-col>

            <v-col cols="12" md="6" class="pa-2 p-camera-select">
              <v-select
                v-model="model.CameraID"
                :append-icon="model.CameraSrc === 'mdi-human-male' ? 'mdi-check' : ''"
                :disabled="disabled"
                :label="$gettext('Camera')"
                :menu-props="{ maxHeight: 346 }"
                autocomplete="off"
                hide-details
                item-value="ID"
                item-title="Name"
                :items="cameraOptions"
                class="input-camera"
              >
              </v-select>
            </v-col>

            <v-col cols="6" md="3" class="pa-2">
              <v-text-field v-model="model.Iso" :disabled="disabled" hide-details autocomplete="off" autocorrect="off" autocapitalize="none" label="ISO" placeholder="" class="input-iso"></v-text-field>
            </v-col>

            <v-col cols="6" md="3" class="pa-2">
              <v-text-field v-model="model.Exposure" :disabled="disabled" hide-details autocomplete="off" autocorrect="off" autocapitalize="none" :label="$gettext('Exposure')" placeholder="" class="input-exposure"></v-text-field>
            </v-col>

            <v-col cols="12" md="6" class="pa-2 p-lens-select">
              <v-select
                v-model="model.LensID"
                :append-icon="model.CameraSrc === 'mdi-human-male' ? 'mdi-check' : ''"
                :disabled="disabled"
                :label="$gettext('Lens')"
                :menu-props="{ maxHeight: 346 }"
                autocomplete="off"
                hide-details
                item-value="ID"
                item-title="Name"
                :items="lensOptions"
                class="input-lens"
              >
              </v-select>
            </v-col>

            <v-col cols="6" md="3" class="pa-2">
              <v-text-field v-model="model.FNumber" f :disabled="disabled" hide-details autocomplete="off" autocorrect="off" autocapitalize="none" :label="$gettext('F Number')" placeholder="" class="input-fnumber"></v-text-field>
            </v-col>

            <v-col cols="6" md="3" class="pa-2">
              <v-text-field v-model="model.FocalLength" :disabled="disabled" hide-details autocomplete="off" :label="$gettext('Focal Length')" placeholder="" class="input-focal-length"></v-text-field>
            </v-col>

            <v-col cols="12" md="6" class="pa-2">
              <v-text-field
                v-model="model.Details.Artist"
                :append-icon="model.Details.ArtistSrc === 'mdi-human-male' ? 'mdi-check' : ''"
                :disabled="disabled"
                :rules="[textRule]"
                hide-details
                autocomplete="off"
                :label="$gettext('Artist')"
                placeholder=""
                class="input-artist"
              ></v-text-field>
            </v-col>

            <v-col cols="6" md="3" class="pa-2">
              <v-text-field
                v-model="model.Details.Copyright"
                :append-icon="model.Details.CopyrightSrc === 'mdi-human-male' ? 'mdi-check' : ''"
                :disabled="disabled"
                :rules="[textRule]"
                hide-details
                autocomplete="off"
                :label="$gettext('Copyright')"
                placeholder=""
                class="input-copyright"
              ></v-text-field>
            </v-col>

            <v-col cols="6" md="3" class="pa-2">
              <v-textarea
                v-model="model.Details.License"
                :append-icon="model.Details.LicenseSrc === 'mdi-human-male' ? 'mdi-check' : ''"
                :disabled="disabled"
                :rules="[textRule]"
                hide-details
                autocomplete="off"
                auto-grow
                :label="$gettext('License')"
                placeholder=""
                :rows="1"
                class="input-license"
              ></v-textarea>
            </v-col>

            <v-col cols="12" class="pa-2">
              <v-textarea
                v-model="model.Details.Subject"
                :append-icon="model.Details.SubjectSrc === 'mdi-human-male' ? 'mdi-check' : ''"
                :disabled="disabled"
                :rules="[textRule]"
                hide-details
                autocomplete="off"
                auto-grow
                :label="$gettext('Subject')"
                placeholder=""
                :rows="1"
                class="input-subject"
              ></v-textarea>
            </v-col>

            <v-col cols="12" class="pa-2">
              <v-textarea
                v-model="model.Description"
                :append-icon="model.DescriptionSrc === 'mdi-human-male' ? 'mdi-check' : ''"
                :disabled="disabled"
                hide-details
                autocomplete="off"
                auto-grow
                :label="$gettext('Description')"
                placeholder=""
                :rows="1"
                class="input-description"
              ></v-textarea>
            </v-col>

            <v-col cols="12" md="8" class="pa-2">
              <v-textarea
                v-model="model.Details.Keywords"
                :append-icon="model.Details.KeywordsSrc === 'mdi-human-male' ? 'mdi-check' : ''"
                :disabled="disabled"
                hide-details
                autocomplete="off"
                auto-grow
                :label="$gettext('Keywords')"
                placeholder=""
                :rows="1"
                class="input-keywords"
              ></v-textarea>
            </v-col>

            <v-col cols="12" md="4" class="pa-2">
              <v-textarea
                v-model="model.Details.Notes"
                :append-icon="model.Details.NotesSrc === 'mdi-human-male' ? 'mdi-check' : ''"
                :disabled="disabled"
                hide-details
                autocomplete="off"
                auto-grow
                :label="$gettext('Notes')"
                placeholder=""
                :rows="1"
                class="input-notes"
              ></v-textarea>
            </v-col>

            <v-col v-if="!disabled" cols="12" :class="rtl ? 'text-left' : 'text-right'" class="pt-6">
              <v-btn color="button" variant="flat" class="compact action-close ma-1" @click.stop="close">
                <translate>Close</translate>
              </v-btn>
              <v-btn color="primary-button" variant="flat" class="compact action-apply action-approve ma-1" @click.stop="save(false)">
                <span v-if="$config.feature('review') && model.Quality < 3"><translate>Approve</translate></span>
                <span v-else><translate>Apply</translate></span>
              </v-btn>
              <v-btn color="primary-button" variant="flat" class="compact action-done hidden-xs ma-1" @click.stop="save(true)">
                <translate>Done</translate>
              </v-btn>
            </v-col>
          </v-row>
        </v-col>
      </v-row>
      <div class="mt-1 clear"></div>
    </v-form>
  </div>
</template>

<script>
import countries from "options/countries.json";
import Thumb from "model/thumb";
import Photo from "model/photo";
import * as options from "options/options";

export default {
  name: "PTabPhotoDetails",
  props: {
    model: {
      type: Object,
      default: () => new Photo(false),
    },
    uid: {
      type: String,
      default: "",
    },
  },
  data() {
    return {
      disabled: !this.$config.feature("edit"),
      config: this.$config.values,
      all: {
        colors: [{ label: this.$gettext("Unknown"), name: "" }],
      },
      readonly: this.$config.get("readonly"),
      options: options,
      countries: countries,
      showDatePicker: false,
      showTimePicker: false,
      invalidDate: false,
      time: "",
      textRule: (v) => v.length <= this.$config.get("clip") || this.$gettext("Text too long"),
      rtl: this.$rtl,
    };
  },
  computed: {
    cameraOptions() {
      return this.config.cameras;
    },
    lensOptions() {
      return this.config.lenses;
    },
  },
  watch: {
    model() {
      this.updateTime();
    },
    uid() {
      this.updateTime();
    },
  },
  created() {
    this.updateTime();
  },
  methods: {
    updateTime() {
      if (!this.model.hasId()) {
        return;
      }

      const taken = this.model.getDateTime();

      this.time = taken.toFormat("HH:mm:ss");
    },
    pastePosition(event) {
      // Auto-fills the lat and lng fields if the text in the clipboard contains two float values.
      const clipboard = event.clipboardData ? event.clipboardData : window.clipboardData;

      if (!clipboard) {
        return;
      }

      // Get values from browser clipboard.
      const text = clipboard.getData("text");

      // Trim spaces before splitting by whitespace and/or commas.
      const val = text.trim().split(/[ ,]+/);

      // Two values found?
      if (val.length >= 2) {
        // Parse values.
        const lat = parseFloat(val[0]);
        const lng = parseFloat(val[1]);

        // Lat and long must be valid floating point numbers.
        if (!isNaN(lat) && lat >= -90 && lat <= 90 && !isNaN(lng) && lng >= -180 && lng <= 180) {
          // Update model values.
          this.model.Lat = lat;
          this.model.Lng = lng;
          // Prevent default action.
          event.preventDefault();
        }
      }
    },
    updateModel() {
      if (!this.model.hasId()) {
        return;
      }

      let localDate = this.model.localDate(this.time);

      this.invalidDate = !localDate.isValid;

      if (this.invalidDate) {
        return;
      }

      if (this.model.Day === 0) {
        this.model.Day = parseInt(localDate.toFormat("d"));
      }

      if (this.model.Month === 0) {
        this.model.Month = parseInt(localDate.toFormat("L"));
      }

      if (this.model.Year === 0) {
        this.model.Year = parseInt(localDate.toFormat("y"));
      }

      const isoTime =
        localDate.toISO({
          suppressMilliseconds: true,
          includeOffset: false,
        }) + "Z";

      this.model.TakenAtLocal = isoTime;

      if (this.model.currentTimeZoneUTC()) {
        this.model.TakenAt = isoTime;
      }
    },
    left() {
      this.$emit("next");
    },
    right() {
      this.$emit("prev");
    },
    openPhoto() {
      this.$viewer.show(Thumb.fromFiles([this.model]), 0);
    },
    save(close) {
      if (this.invalidDate) {
        this.$notify.error(this.$gettext("Invalid date"));
        return;
      }

      this.updateModel();

      this.model.update().then(() => {
        if (close) {
          this.$emit("close");
        }

        this.updateTime();
      });
    },
    close() {
      this.$emit("close");
    },
  },
};
</script>
