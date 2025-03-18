<template>
  <div class="p-tab p-tab-photo-details">
    <v-form
      ref="form"
      validate-on="invalid-input"
      class="p-form p-form-photo-details-meta"
      accept-charset="UTF-8"
      tabindex="1"
      @submit.prevent="save"
    >
      <div class="form-body">
        <div class="form-controls">
          <v-row dense align="start">
            <v-col cols="3" sm="2" class="form-thumb">
              <div>
                <img
                  :alt="view.model.Title"
                  :src="view.model.thumbnailUrl('tile_500')"
                  class="clickable"
                  @click.stop.prevent.exact="openPhoto()"
                />
              </div>
            </v-col>
            <v-col cols="9" sm="10" class="d-flex align-self-stretch flex-column ga-4">
              <v-text-field
                v-model="view.model.Title"
                :append-inner-icon="view.model.TitleSrc === 'manual' ? 'mdi-check' : ''"
                :disabled="disabled"
                :rules="[textRule]"
                hide-details
                :label="$pgettext('Photo', 'Title')"
                placeholder=""
                autocomplete="off"
                density="comfortable"
                class="input-title"
              ></v-text-field>
              <v-textarea
                v-model="view.model.Caption"
                :append-inner-icon="view.model.CaptionSrc === 'manual' ? 'mdi-check' : ''"
                :disabled="disabled"
                hide-details
                autocomplete="off"
                auto-grow
                :label="$gettext('Caption')"
                placeholder=""
                :rows="1"
                density="comfortable"
                class="input-caption"
              ></v-textarea>
            </v-col>
          </v-row>
          <v-row dense>
            <v-col cols="4" lg="2">
              <v-combobox
                :model-value="view.model.Day > 0 ? view.model.Day : null"
                :disabled="disabled"
                :error="invalidDate"
                :label="$gettext('Day')"
                :placeholder="$gettext('Unknown')"
                :prepend-inner-icon="$vuetify.display.xs ? undefined : 'mdi-calendar'"
                autocomplete="off"
                hide-details
                hide-no-data
                :items="options.Days()"
                item-title="text"
                item-value="value"
                density="comfortable"
                validate-on="input"
                :rules="rules.day(false)"
                class="input-day"
                @update:model-value="setDay"
              >
              </v-combobox>
            </v-col>
            <v-col cols="4" lg="2">
              <v-combobox
                :model-value="view.model.Month > 0 ? view.model.Month : null"
                :disabled="disabled"
                :error="invalidDate"
                :label="$gettext('Month')"
                :placeholder="$gettext('Unknown')"
                autocomplete="off"
                hide-details
                hide-no-data
                :items="options.MonthsShort()"
                item-title="text"
                item-value="value"
                density="comfortable"
                validate-on="input"
                :rules="rules.month(false)"
                class="input-month"
                @update:model-value="setMonth"
              >
              </v-combobox>
            </v-col>
            <v-col cols="4" lg="2">
              <v-combobox
                :model-value="view.model.Year > 0 ? view.model.Year : null"
                :disabled="disabled"
                :error="invalidDate"
                :label="$gettext('Year')"
                :placeholder="$gettext('Unknown')"
                autocomplete="off"
                hide-details
                hide-no-data
                :items="options.Years(1900)"
                item-title="text"
                item-value="value"
                density="comfortable"
                validate-on="input"
                :rules="rules.year(false, 1000)"
                class="input-year"
                @update:model-value="setYear"
              >
              </v-combobox>
            </v-col>
            <v-col cols="6" lg="2">
              <v-text-field
                v-model="time"
                :append-inner-icon="view.model.TakenSrc === 'manual' ? 'mdi-check' : ''"
                :disabled="disabled"
                :label="view.model.timeIsUTC() ? $gettext('Time UTC') : $gettext('Local Time')"
                :prepend-inner-icon="$vuetify.display.xs ? undefined : 'mdi-clock-time-eight-outline'"
                autocomplete="off"
                autocorrect="off"
                autocapitalize="none"
                hide-details
                density="comfortable"
                validate-on="input"
                :rules="rules.time()"
                class="input-local-time"
                @update:model-value="setTime"
              ></v-text-field>
            </v-col>
            <v-col cols="6" lg="4">
              <v-autocomplete
                v-model="view.model.TimeZone"
                :disabled="disabled"
                :label="$gettext('Time Zone')"
                hide-no-data
                item-value="ID"
                item-title="Name"
                :items="options.TimeZones()"
                density="comfortable"
                class="input-timezone"
                @update:model-value="syncTime"
              ></v-autocomplete>
            </v-col>
            <v-col cols="12" sm="8" md="4">
              <v-autocomplete
                v-model="view.model.Country"
                :append-inner-icon="view.model.PlaceSrc === 'manual' ? 'mdi-check' : ''"
                :disabled="disabled"
                :readonly="!!(view.model.Lat || view.model.Lng)"
                :placeholder="$gettext('Country')"
                hide-details
                hide-no-data
                autocomplete="off"
                item-value="Code"
                item-title="Name"
                :items="countries"
                prepend-inner-icon="mdi-map-marker"
                density="comfortable"
                validate-on="input"
                :rules="rules.country(true)"
                class="input-country"
              >
              </v-autocomplete>
            </v-col>
            <v-col cols="4" md="2">
              <v-text-field
                v-model="view.model.Altitude"
                :disabled="disabled"
                hide-details
                flat
                autocomplete="off"
                autocorrect="off"
                autocapitalize="none"
                :label="$gettext('Altitude (m)')"
                placeholder=""
                color="surface-variant"
                density="comfortable"
                validate-on="input"
                :rules="rules.number(false, -10000, 1000000)"
                class="input-altitude"
              ></v-text-field>
            </v-col>
            <v-col cols="4" sm="6" md="3">
              <v-text-field
                v-model="view.model.Lat"
                :append-inner-icon="view.model.PlaceSrc === 'manual' ? 'mdi-check' : ''"
                :disabled="disabled"
                hide-details
                autocomplete="off"
                autocorrect="off"
                autocapitalize="none"
                :label="$gettext('Latitude')"
                placeholder=""
                density="comfortable"
                validate-on="input"
                :rules="rules.lat(false)"
                class="input-latitude"
                @paste="pastePosition"
              ></v-text-field>
            </v-col>
            <v-col cols="4" sm="6" md="3">
              <v-text-field
                v-model="view.model.Lng"
                :append-inner-icon="view.model.PlaceSrc === 'manual' ? 'mdi-check' : ''"
                :disabled="disabled"
                hide-details
                autocomplete="off"
                autocorrect="off"
                autocapitalize="none"
                :label="$gettext('Longitude')"
                placeholder=""
                density="comfortable"
                validate-on="input"
                :rules="rules.lng(false)"
                class="input-longitude"
                @paste="pastePosition"
              ></v-text-field>
            </v-col>
            <v-col cols="12" md="6" class="p-camera-select">
              <v-select
                v-model="view.model.CameraID"
                :append-inner-icon="view.model.CameraSrc === 'manual' ? 'mdi-check' : ''"
                :disabled="disabled"
                :placeholder="$gettext('Camera')"
                :menu-props="{ maxHeight: 346 }"
                autocomplete="off"
                hide-details
                item-value="ID"
                item-title="Name"
                :items="cameraOptions"
                prepend-inner-icon="mdi-camera"
                density="comfortable"
                class="input-camera"
              >
              </v-select>
            </v-col>
            <v-col cols="6" md="3">
              <v-text-field
                v-model="view.model.Iso"
                :disabled="disabled"
                hide-details
                autocomplete="off"
                autocorrect="off"
                autocapitalize="none"
                label="ISO"
                placeholder=""
                density="comfortable"
                validate-on="input"
                :rules="rules.number(false, 0, 1048576)"
                class="input-iso"
              ></v-text-field>
            </v-col>
            <v-col cols="6" md="3">
              <v-text-field
                v-model="view.model.Exposure"
                :disabled="disabled"
                hide-details
                autocomplete="off"
                autocorrect="off"
                autocapitalize="none"
                :label="$gettext('Exposure')"
                placeholder=""
                density="comfortable"
                validate-on="input"
                :rules="rules.text(false, 0, 64)"
                class="input-exposure"
              ></v-text-field>
            </v-col>
            <v-col cols="12" md="6" class="p-lens-select">
              <v-select
                v-model="view.model.LensID"
                :append-inner-icon="view.model.CameraSrc === 'manual' ? 'mdi-check' : ''"
                :disabled="disabled"
                :placeholder="$gettext('Lens')"
                :menu-props="{ maxHeight: 346 }"
                autocomplete="off"
                hide-details
                item-value="ID"
                item-title="Name"
                :items="lensOptions"
                prepend-inner-icon="mdi-camera-iris"
                density="comfortable"
                class="input-lens"
              >
              </v-select>
            </v-col>
            <v-col cols="6" md="3">
              <v-text-field
                v-model="view.model.FNumber"
                :disabled="disabled"
                hide-details
                autocomplete="off"
                autocorrect="off"
                autocapitalize="none"
                :label="$gettext('F Number')"
                placeholder=""
                density="comfortable"
                validate-on="input"
                :rules="rules.number(false, 0, 1048576)"
                class="input-fnumber"
              ></v-text-field>
            </v-col>
            <v-col cols="6" md="3">
              <v-text-field
                v-model="view.model.FocalLength"
                :disabled="disabled"
                hide-details
                autocomplete="off"
                :label="$gettext('Focal Length')"
                placeholder=""
                density="comfortable"
                validate-on="input"
                :rules="rules.number(false, 0, 1048576)"
                class="input-focal-length"
              ></v-text-field>
            </v-col>
          </v-row>
          <v-row dense>
            <v-col cols="12" md="6">
              <v-textarea
                v-model="view.model.Details.Subject"
                :append-inner-icon="view.model.Details.SubjectSrc === 'manual' ? 'mdi-check' : ''"
                :disabled="disabled"
                :rules="[textRule]"
                hide-details
                autocomplete="off"
                auto-grow
                :label="$gettext('Subject')"
                placeholder=""
                :rows="1"
                density="comfortable"
                class="input-subject"
              ></v-textarea>
            </v-col>
            <v-col cols="12" md="6">
              <v-text-field
                v-model="view.model.Details.Copyright"
                :append-inner-icon="view.model.Details.CopyrightSrc === 'manual' ? 'mdi-check' : ''"
                :disabled="disabled"
                :rules="[textRule]"
                hide-details
                autocomplete="off"
                :label="$gettext('Copyright')"
                placeholder=""
                density="comfortable"
                class="input-copyright"
              ></v-text-field>
            </v-col>
            <v-col cols="12" md="6">
              <v-text-field
                v-model="view.model.Details.Artist"
                :append-inner-icon="view.model.Details.ArtistSrc === 'manual' ? 'mdi-check' : ''"
                :disabled="disabled"
                :rules="[textRule]"
                hide-details
                autocomplete="off"
                :label="$gettext('Artist')"
                placeholder=""
                density="comfortable"
                class="input-artist"
              ></v-text-field>
            </v-col>
            <v-col cols="12" md="6">
              <v-textarea
                v-model="view.model.Details.License"
                :append-inner-icon="view.model.Details.LicenseSrc === 'manual' ? 'mdi-check' : ''"
                :disabled="disabled"
                :rules="[textRule]"
                hide-details
                autocomplete="off"
                auto-grow
                :label="$gettext('License')"
                placeholder=""
                :rows="1"
                density="comfortable"
                class="input-license"
              ></v-textarea>
            </v-col>
            <v-col cols="12" md="8">
              <v-textarea
                v-model="view.model.Details.Keywords"
                :append-inner-icon="view.model.Details.KeywordsSrc === 'manual' ? 'mdi-check' : ''"
                :disabled="disabled"
                hide-details
                autocomplete="off"
                auto-grow
                :label="$gettext('Keywords')"
                placeholder=""
                :rows="1"
                density="default"
                class="input-keywords"
              ></v-textarea>
            </v-col>
            <v-col cols="12" md="4">
              <v-textarea
                v-model="view.model.Details.Notes"
                :append-inner-icon="view.model.Details.NotesSrc === 'manual' ? 'mdi-check' : ''"
                :disabled="disabled"
                hide-details
                autocomplete="off"
                auto-grow
                :label="$gettext('Notes')"
                placeholder=""
                :rows="1"
                density="default"
                class="input-notes"
              ></v-textarea>
            </v-col>
          </v-row>
        </div>
      </div>
      <div v-if="!disabled" class="form-actions form-actions--sticky">
        <div class="action-buttons">
          <v-btn color="button" variant="flat" class="action-close" @click.stop="close">
            {{ $gettext(`Close`) }}
          </v-btn>
          <v-btn
            color="highlight"
            variant="flat"
            :disabled="!view.model?.wasChanged() && !inReview"
            class="action-apply action-approve"
            @click.stop="save(false)"
          >
            <span v-if="inReview">{{ $gettext(`Approve`) }}</span>
            <span v-else>{{ $gettext(`Apply`) }}</span>
          </v-btn>
        </div>
      </div>
    </v-form>
  </div>
</template>

<script>
import countries from "options/countries.json";
import Thumb from "model/thumb";
import * as options from "options/options";
import { rules } from "common/form";

export default {
  name: "PTabPhotoDetails",
  props: {
    uid: {
      type: String,
      default: "",
    },
  },
  data() {
    return {
      view: this.$view.data(),
      disabled: !this.$config.feature("edit"),
      config: this.$config.values,
      all: {
        colors: [{ label: this.$gettext("Unknown"), name: "" }],
      },
      readonly: this.$config.get("readonly"),
      options,
      rules,
      countries,
      featReview: this.$config.feature("review"),
      showDatePicker: false,
      showTimePicker: false,
      invalidDate: false,
      time: "",
      textRule: (v) => v.length <= this.$config.get("clip") || this.$gettext("Text too long"),
      rtl: this.$isRtl,
    };
  },
  computed: {
    cameraOptions() {
      return this.config.cameras;
    },
    lensOptions() {
      return this.config.lenses;
    },
    inReview() {
      return this.featReview && this.view.model.Quality < 3;
    },
  },
  watch: {
    uid() {
      this.syncTime();
    },
  },
  created() {
    this.syncTime();
  },
  methods: {
    setDay(v) {
      if (Number.isInteger(v?.value)) {
        this.view.model.Day = v?.value;
        this.syncTime();
      } else if (!v) {
        this.view.model.Day = -1;
      } else if (this.rules.isNumberRange(v, 1, 31)) {
        this.view.model.Day = Number(v);
        this.syncTime();
      }
    },
    setMonth(v) {
      if (Number.isInteger(v?.value)) {
        this.view.model.Month = v?.value;
        this.syncTime();
      } else if (!v) {
        this.view.model.Month = -1;
      } else if (this.rules.isNumberRange(v, 1, 12)) {
        this.view.model.Month = Number(v);
        this.syncTime();
      }
    },
    setYear(v) {
      if (Number.isInteger(v?.value)) {
        this.view.model.Year = v?.value;
        this.syncTime();
      } else if (!v) {
        this.view.model.Year = -1;
      } else if (this.rules.isNumberRange(v, 1000, Number(new Date().getUTCFullYear()))) {
        this.view.model.Year = Number(v);
        this.syncTime();
      }
    },
    setTime() {
      if (this.rules.isTime(this.time)) {
        this.updateModel();
      }
    },
    syncTime() {
      if (!this.view?.model.hasId()) {
        return;
      }

      const taken = this.view.model.getDateTime();
      this.time = taken.toFormat("HH:mm:ss");
    },
    pastePosition(event) {
      // Autofill the lat and lng fields if the text in the clipboard contains two float values.
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
          // Update view.model values.
          this.view.model.Lat = lat;
          this.view.model.Lng = lng;
          // Prevent default action.
          event.preventDefault();
        }
      }
    },
    updateModel() {
      if (!this.view?.model.hasId()) {
        return;
      }

      let localDate = this.view.model.localDate(this.time);

      this.invalidDate = !localDate.isValid;

      if (this.invalidDate) {
        return;
      }

      if (this.view.model.Day === 0) {
        this.view.model.Day = parseInt(localDate.toFormat("d"));
      }

      if (this.view.model.Month === 0) {
        this.view.model.Month = parseInt(localDate.toFormat("L"));
      }

      if (this.view.model.Year === 0) {
        this.view.model.Year = parseInt(localDate.toFormat("y"));
      }

      const isoTime =
        localDate.toISO({
          suppressMilliseconds: true,
          includeOffset: false,
        }) + "Z";

      this.view.model.TakenAtLocal = isoTime;

      if (this.view.model.currentTimeZoneUTC()) {
        this.view.model.TakenAt = isoTime;
      }
    },
    left() {
      this.$emit("next");
    },
    right() {
      this.$emit("prev");
    },
    openPhoto() {
      this.$lightbox.openModels(Thumb.fromFiles([this.view.model]), 0);
    },
    save(close) {
      if (this.invalidDate) {
        this.$notify.error(this.$gettext("Invalid date"));
        return;
      }

      this.updateModel();

      this.view.model.update().then(() => {
        if (close) {
          this.$emit("close");
        }

        this.syncTime();
      });
    },
    close() {
      this.$emit("close");
    },
  },
};
</script>
