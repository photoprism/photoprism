<template>
  <v-dialog
    :value="show"
    lazy
    persistent
    max-width="350"
    class="p-people-merge-dialog"
    @keydown.esc="cancel"
  >
    <v-card raised elevation="24">
      <v-container fluid class="pb-2 pr-2 pl-2">
        <v-layout row wrap>
          <v-flex xs3 text-xs-center>
            <v-icon size="54" color="secondary-dark lighten-1">people</v-icon>
          </v-flex>
          <v-flex xs9 text-xs-left align-self-center>
            <div class="subheading pr-1">
              {{ prompt }}
            </div>
          </v-flex>
          <v-flex xs12 text-xs-right class="pt-3">
            <v-btn depressed color="secondary-light" class="action-cancel" @click.stop="cancel">
              <translate key="No">No</translate>
            </v-btn>
            <v-btn
              color="primary-button"
              depressed
              dark
              class="action-confirm"
              @click.stop="confirm"
            >
              <translate key="Yes">Yes</translate>
            </v-btn>
          </v-flex>
        </v-layout>
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
