<template>
  <v-dialog :model-value="show" persistent max-width="350" class="p-people-merge-dialog" @keydown.esc="cancel">
    <v-card elevation="24">
      <v-container fluid class="pb-2 pr-2 pl-2">
        <v-row>
          <v-col cols="3" class="text-center">
            <v-icon size="54" color="secondary-dark lighten-1">mdi-account-multiple</v-icon>
          </v-col>
          <v-col cols="9" class="text-left" align-self="center">
            <div class="text-subtitle-1 pr-1">
              {{ prompt }}
            </div>
          </v-col>
          <v-col cols="12" class="text-right pt-6">
            <v-btn variant="flat" color="secondary-light" class="action-cancel" @click.stop="cancel">
              <translate key="No">No</translate>
            </v-btn>
            <v-btn color="primary-button" variant="flat" theme="dark" class="action-confirm" @click.stop="confirm">
              <translate key="Yes">Yes</translate>
            </v-btn>
          </v-col>
        </v-row>
      </v-container>
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
