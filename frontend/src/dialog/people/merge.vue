<template>
  <v-dialog :model-value="show" persistent max-width="350" class="p-people-merge-dialog" @keydown.esc="cancel">
    <v-card>
      <v-card-title class="d-flex justify-start align-center ga-3">
        <v-icon size="54" color="primary">mdi-account-multiple</v-icon>
        <p class="text-subtitle-1">{{ prompt }}</p>
      </v-card-title>
      <v-card-actions>
        <v-btn variant="flat" color="button" class="action-cancel" @click.stop="cancel">
          <translate key="No">No</translate>
        </v-btn>
        <v-btn color="primary-button" variant="flat" class="action-confirm" @click.stop="confirm">
          <translate key="Yes">Yes</translate>
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
    show: Boolean,
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
    cancel() {
      this.$emit("cancel");
    },
    confirm() {
      this.$emit("confirm");
    },
  },
};
</script>
