<template>
  <v-dialog
    ref="dialog"
    :model-value="visible"
    persistent
    max-width="610"
    class="p-dialog modal-dialog p-settings-connected-apps"
    @keydown.esc.exact="close"
    @after-enter="afterEnter"
    @after-leave="afterLeave"
  >
    <v-card>
      <v-card-title class="d-flex justify-start align-center ga-3">
        <v-icon size="28" color="primary">mdi-connection</v-icon>
        <h6 class="text-h6">{{ $gettext(`Connected Apps`) }}</h6>
      </v-card-title>
      <v-card-text>
        <v-row class="align-start" no-gutters>
          <v-col cols="12">
            <v-data-table
              :headers="listColumns"
              :items="results"
              :items-per-page-options="[]"
              :loading="loading"
              hide-default-footer
              item-key="UID"
              :no-data-text="noDataText"
              density="compact"
              class="elevation-0 user-results list-view"
            >
              <template #item="props">
                <tr :data-uid="props.item.UID">
                  <td class="text-selectable text-break text-start">
                    {{ props.item.ClientName || props.item.ClientID }}
                    <div class="text-caption text-secondary">{{ props.item.Origin }}</div>
                  </td>
                  <td class="text-start text-break hidden-xs">
                    {{ props.item.Scope }}
                  </td>
                  <td class="text-start text-break" nowrap>
                    {{ formatDateTime(props.item.UsedAt) }}
                  </td>
                  <td>
                    <div class="table-actions">
                      <v-btn
                        icon="mdi-delete"
                        color="surface-variant"
                        density="compact"
                        variant="plain"
                        :ripple="false"
                        :aria-label="$gettext('Disconnect')"
                        class="action-remove action-secondary"
                        @click.stop.prevent="onRevoke(props.item)"
                      ></v-btn>
                    </div>
                  </td>
                </tr>
              </template>
            </v-data-table>
          </v-col>
        </v-row>
      </v-card-text>
      <v-card-actions class="action-buttons">
        <v-btn variant="flat" color="button" class="action-close" @click.stop="close">
          {{ $gettext(`Close`) }}
        </v-btn>
      </v-card-actions>
    </v-card>
    <p-confirm-dialog :visible="revoke.dialog" icon="mdi-delete-outline" @close="revoke.dialog = false" @confirm="onRevoked"></p-confirm-dialog>
  </v-dialog>
</template>
<script>
import { DateTime } from "luxon";
import PConfirmDialog from "component/confirm/dialog.vue";

// PSettingsConnectedApps lists the apps connected to an account with OAuth sign-in and lets the user disconnect them.
export default {
  name: "PSettingsConnectedApps",
  components: {
    PConfirmDialog,
  },
  props: {
    visible: {
      type: Boolean,
      default: false,
    },
    model: {
      type: Object,
      default: () => {},
    },
  },
  emits: ["close"],
  data() {
    return {
      loading: false,
      failed: false,
      busy: false,
      results: [],
      revoke: {
        uid: "",
        dialog: false,
      },
      timeZone: this.$config.getTimeZone(),
      listColumns: [
        { title: this.$gettext("Name"), key: "ClientName", sortable: false, align: "left" },
        { title: this.$gettext("Scope"), headerProps: { class: "hidden-xs" }, key: "Scope", sortable: false, align: "left" },
        { title: this.$gettext("Last Used"), headerProps: { class: "text-no-wrap" }, key: "UsedAt", sortable: false, align: "left" },
        { title: "", key: "", sortable: false, align: "right" },
      ],
    };
  },
  computed: {
    // noDataText returns the text of an empty list, which reports a failed load.
    noDataText() {
      return this.failed ? this.$gettext("Something went wrong, try again") : this.$gettext("No apps are connected to your account.");
    },
  },
  watch: {
    visible: function (show) {
      if (show) {
        this.find();
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
    close() {
      if (this.busy) {
        return;
      }

      this.$emit("close");
    },
    formatDateTime(d) {
      if (!d) {
        return "–";
      }

      return DateTime.fromISO(d, { zone: this.timeZone }).toLocaleString(DateTime.DATETIME_SHORT);
    },
    // find loads the connected apps of the account.
    find() {
      if (!this.model || typeof this.model.findGrants !== "function") {
        return;
      }

      this.loading = true;
      this.failed = false;

      this.model
        .findGrants()
        .then((results) => {
          this.results = Array.isArray(results) ? results : [];
        })
        .catch(() => {
          this.results = [];
          this.failed = true;
        })
        .finally(() => {
          this.loading = false;
        });
    },
    onRevoke(app) {
      if (this.busy) {
        return;
      }

      this.revoke.uid = app.UID;
      this.revoke.dialog = true;
    },
    // onRevoked disconnects the selected app after the user confirmed it.
    onRevoked() {
      if (this.busy || !this.revoke.uid) {
        return;
      }

      this.busy = true;

      this.model
        .revokeGrant(this.revoke.uid)
        .then(() => {
          this.$notify.info(this.$gettext("Successfully deleted"));
          this.revoke.uid = "";
          this.revoke.dialog = false;
          this.find();
        })
        .catch(() => {
          this.revoke.dialog = false;
        })
        .finally(() => {
          this.busy = false;
        });
    },
  },
};
</script>
