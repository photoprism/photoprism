<template>
  <v-dialog
    :model-value="visible"
    persistent
    max-width="350"
    class="p-dialog p-people-merge-dialog"
    @keydown.esc="close"
  >
    <v-card>
      <v-card-title class="d-flex justify-start align-center ga-3">
        <v-icon size="54" color="primary">mdi-account-multiple</v-icon>
        <p class="text-subtitle-1">{{ prompt }}</p>
      </v-card-title>
      <v-card-actions class="dialog-merge action-buttons">
        <v-btn variant="flat" color="button" class="action-cancel" @click.stop="close">
          {{ $gettext(`No`) }}
        </v-btn>
        <v-btn color="highlight" variant="flat" class="action-confirm" @click.stop="confirm">
          {{ $gettext(`Yes`) }}
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>
<script>
import Subject from "model/subject";

export default {
  name: "PPeopleMergeDialog",
  props: {
    visible: {
      type: Boolean,
      default: false,
    },
    subj1: {
      type: Object,
      default: new Subject(),
    },
    subj2: {
      type: Object,
      default: new Subject(),
    },
  },
  data() {
    return {};
  },
  computed: {
    prompt() {
      if (!this.subj1 || !this.subj2) {
        return "";
      }

      return this.$gettextInterpolate(this.$gettext("Merge %{a} with %{b}?"), {
        a: this.subj1.originalValue("Name"),
        b: this.subj2.Name,
      });
    },
  },
  methods: {
    close() {
      this.$emit("close");
    },
    confirm() {
      this.$emit("confirm");
    },
  },
};
</script>
