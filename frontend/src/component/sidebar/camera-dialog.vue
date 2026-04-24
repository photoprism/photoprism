<template>
  <v-dialog
    ref="dialog"
    :model-value="visible"
    :max-width="480"
    :fullscreen="$vuetify.display.xs"
    persistent
    scrim
    scrollable
    class="p-camera-dialog"
    @keydown.esc.exact="close"
    @after-enter="afterEnter"
    @after-leave="afterLeave"
  >
    <v-card ref="content" tabindex="-1" :tile="$vuetify.display.xs">
      <v-toolbar flat color="navigation" density="comfortable">
        <v-toolbar-title>
          {{ $gettext("Edit Camera Details") }}
        </v-toolbar-title>
        <v-btn icon :aria-label="$gettext('Close')" @click.stop="close">
          <v-icon>mdi-close</v-icon>
        </v-btn>
      </v-toolbar>
      <v-card-text class="pb-3">
        <v-row dense>
          <v-col cols="12">
            <v-select
              v-model="cameraID"
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
          <v-col cols="12">
            <v-select
              v-model="lensID"
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
          <v-col cols="6">
            <v-text-field
              v-model="iso"
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
          <v-col cols="6">
            <v-text-field
              v-model="exposure"
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
          <v-col cols="6">
            <v-text-field
              v-model="fNumber"
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
          <v-col cols="6">
            <v-text-field
              v-model="focalLength"
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
        <div class="action-buttons mt-4 d-flex justify-end ga-2">
          <v-btn variant="flat" color="button" class="action-cancel" min-width="100" @click.stop="close">
            {{ $gettext("Cancel") }}
          </v-btn>
          <v-btn color="info" class="action-confirm" min-width="100" @click="confirm">
            {{ $gettext("Confirm") }}
          </v-btn>
        </div>
      </v-card-text>
    </v-card>
  </v-dialog>
</template>

<script>
import { rules } from "common/form";

export default {
  name: "PCameraDialog",
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
      rules,
      cameraID: 0,
      lensID: 0,
      iso: "",
      exposure: "",
      fNumber: "",
      focalLength: "",
    };
  },
  computed: {
    cameraOptions() {
      return this.$config.values.cameras || [];
    },
    lensOptions() {
      return this.$config.values.lenses || [];
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

      this.cameraID = this.photo.CameraID || 0;
      this.lensID = this.photo.LensID || 0;
      this.iso = this.photo.Iso || "";
      this.exposure = this.photo.Exposure || "";
      this.fNumber = this.photo.FNumber || "";
      this.focalLength = this.photo.FocalLength || "";
    },
    close() {
      this.$emit("close");
    },
    confirm() {
      this.$emit("confirm", {
        CameraID: this.cameraID,
        LensID: this.lensID,
        Iso: this.iso,
        Exposure: this.exposure,
        FNumber: this.fNumber,
        FocalLength: this.focalLength,
      });
    },
  },
};
</script>
