<template>
  <v-dialog
    ref="dialog"
    :model-value="visible"
    :max-width="480"
    :fullscreen="$vuetify.display.xs"
    persistent
    scrim
    scrollable
    class="p-datetime-dialog"
    @keydown.esc.exact="close"
    @after-enter="afterEnter"
    @after-leave="afterLeave"
  >
    <v-card ref="content" tabindex="-1" :tile="$vuetify.display.xs">
      <v-toolbar flat color="navigation" density="comfortable">
        <v-toolbar-title>
          {{ $gettext("Edit Date & Time") }}
        </v-toolbar-title>
        <v-btn icon :aria-label="$gettext('Close')" @click.stop="close">
          <v-icon>mdi-close</v-icon>
        </v-btn>
      </v-toolbar>
      <v-card-text class="pb-3">
        <v-row dense>
          <v-col cols="4">
            <v-autocomplete
              :model-value="day > 0 ? day : null"
              :error="invalidDate"
              :label="$gettext('Day')"
              :placeholder="$gettext('Unknown')"
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
            </v-autocomplete>
          </v-col>
          <v-col cols="4">
            <v-autocomplete
              :model-value="month > 0 ? month : null"
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
            </v-autocomplete>
          </v-col>
          <v-col cols="4">
            <v-autocomplete
              :model-value="year > 0 ? year : null"
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
            </v-autocomplete>
          </v-col>
          <v-col cols="6">
            <v-text-field
              v-model="time"
              :label="timeLabel"
              prepend-inner-icon="mdi-clock-time-eight-outline"
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
          <v-col cols="6">
            <v-autocomplete
              v-model="timeZone"
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
        </v-row>
        <div class="action-buttons mt-4 d-flex justify-end ga-2">
          <v-btn variant="flat" color="button" class="action-cancel" min-width="100" @click.stop="close">
            {{ $gettext("Cancel") }}
          </v-btn>
          <v-btn color="info" class="action-confirm" min-width="100" :disabled="invalidDate" @click="confirm">
            {{ $gettext("Confirm") }}
          </v-btn>
        </div>
      </v-card-text>
    </v-card>
  </v-dialog>
</template>

<script>
import { DateTime } from "luxon";
import * as options from "options/options";
import { rules } from "common/form";

export default {
  name: "PDateTimeDialog",
  props: {
    visible: {
      type: Boolean,
      default: false,
    },
    photo: {
      type: Object,
      default: null,
    },
  },
  emits: ["close", "confirm"],
  data() {
    return {
      options,
      rules,
      day: 0,
      month: 0,
      year: 0,
      time: "",
      timeZone: "",
      invalidDate: false,
    };
  },
  computed: {
    timeLabel() {
      if (this.photo && this.photo.timeIsUTC()) {
        return this.$gettext("Time UTC");
      }
      return this.$gettext("Local Time");
    },
  },
  watch: {
    visible(show) {
      if (show) {
        this.loadFromPhoto();
      }
    },
  },
  methods: {
    afterEnter() {
      this.$view.enter(this);
    },
    afterLeave() {
      this.$view.leave(this);
    },
    loadFromPhoto() {
      if (!this.photo) return;

      this.day = this.photo.Day;
      this.month = this.photo.Month;
      this.year = this.photo.Year;
      this.timeZone = this.photo.TimeZone || "";

      const taken = this.photo.getDateTime();
      this.time = taken.toFormat("HH:mm:ss");
      this.invalidDate = false;
    },
    effectiveYear() {
      if (this.year > 0) return this.year;
      const y = this.photo?.TakenAtLocal ? parseInt(this.photo.TakenAtLocal.substring(0, 4)) : new Date().getUTCFullYear();
      return isNaN(y) ? new Date().getUTCFullYear() : y;
    },
    effectiveMonth() {
      if (this.month > 0) return this.month;
      const m = this.photo?.TakenAtLocal ? parseInt(this.photo.TakenAtLocal.substring(5, 7)) : new Date().getUTCMonth() + 1;
      return isNaN(m) ? new Date().getUTCMonth() + 1 : m;
    },
    clampDayToValidRange() {
      if (this.day <= 0) return;
      const maxDay = new Date(Date.UTC(this.effectiveYear(), this.effectiveMonth(), 0)).getUTCDate();
      if (this.day > maxDay) {
        this.day = maxDay;
      }
    },
    setDay(v) {
      if (Number.isInteger(v?.value)) {
        this.day = v.value;
        this.clampDayToValidRange();
        this.syncTime();
      } else if (!v) {
        this.day = -1;
        this.year = -1;
        this.updateLocalDate();
      } else if (rules.isNumberRange(v, 1, 31)) {
        this.day = Number(v);
        this.clampDayToValidRange();
        this.syncTime();
      }
    },
    setMonth(v) {
      if (Number.isInteger(v?.value)) {
        this.month = v.value;
        this.clampDayToValidRange();
        this.syncTime();
      } else if (!v) {
        this.month = -1;
        this.year = -1;
        this.syncTime();
      } else if (rules.isNumberRange(v, 1, 12)) {
        this.month = Number(v);
        this.clampDayToValidRange();
        this.syncTime();
      }
    },
    setYear(v) {
      if (Number.isInteger(v?.value)) {
        this.year = v.value;
        this.clampDayToValidRange();
        this.syncTime();
      } else if (!v) {
        this.year = -1;
        this.syncTime();
      } else if (rules.isNumberRange(v, 1000, Number(new Date().getUTCFullYear()))) {
        this.year = Number(v);
        this.clampDayToValidRange();
        this.syncTime();
      }
    },
    setTime() {
      if (rules.isTime(this.time)) {
        this.updateLocalDate();
      }
    },
    syncTime() {
      this.updateLocalDate();
    },
    localYearString() {
      if (this.year <= 0) return "";
      return this.year.toString().padStart(4, "0");
    },
    localMonthString() {
      if (this.month <= 0) return "01";
      return this.month.toString().padStart(2, "0");
    },
    localDayString() {
      if (this.day <= 0) return "01";
      return this.day.toString().padStart(2, "0");
    },
    updateLocalDate() {
      if (!this.photo) return;

      const yearStr = this.localYearString();
      if (!yearStr) return;

      const date = yearStr + "-" + this.localMonthString() + "-" + this.localDayString();
      const time = this.time || "12:00:00";

      const zone = this.timeZone || "UTC";
      const localDate = DateTime.fromISO(`${date}T${time}`, { zone });

      this.invalidDate = !localDate.isValid;
    },
    close() {
      this.$emit("close");
    },
    confirm() {
      if (this.invalidDate) return;

      this.$emit("confirm", {
        Day: this.day,
        Month: this.month,
        Year: this.year,
        TimeZone: this.timeZone,
        time: this.time,
      });
    },
  },
};
</script>
