<template>
  <div class="p-tab p-settings-services">
    <v-container class="width-lg">
      <v-data-table v-model="selected" :headers="listColumns" :items="results" hide-default-footer class="elevation-0 account-results list-view" item-key="ID" :no-data-text="$gettext('No services configured.')">
        <template #item="props">
          <tr :data-name="props.item.AccName">
            <td class="p-account">
              <button class="surface-variant--text" @click.stop.prevent="edit(props.item)">
                {{ props.item.AccName }}
              </button>
            </td>
            <td class="text-center">
              <v-btn icon density="comfortable" variant="plain" :ripple="false" class="action-toggle-share" @click.stop.prevent="editSharing(props.item)">
                <v-icon v-if="props.item.AccShare" color="surface-variant">mdi-check</v-icon>
                <v-icon v-else color="surface-variant">mdi-cog</v-icon>
              </v-btn>
            </td>
            <td class="text-center">
              <v-btn icon density="comfortable" variant="plain" :ripple="false" class="action-toggle-sync" @click.stop.prevent="editSync(props.item)">
                <v-icon v-if="props.item.AccErrors" color="surface-variant" :title="props.item.AccError">mdi-alert </v-icon>
                <!-- TODO: change icon -->
                <v-icon v-else-if="props.item.AccSync" color="surface-variant">sync</v-icon>
                <!-- TODO: change icon -->
                <v-icon v-else color="surface-variant">sync_disabled</v-icon>
              </v-btn>
            </td>
            <td class="hidden-sm-and-down">
              {{ formatDate(props.item.SyncDate) }}
            </td>
            <td class="hidden-xs text-right" nowrap>
              <v-btn icon density="comfortable" variant="plain" :ripple="false" class="action-remove action-secondary" @click.stop.prevent="remove(props.item)">
                <v-icon color="surface-variant">mdi-delete</v-icon>
              </v-btn>
              <v-btn icon density="comfortable" variant="plain" :ripple="false" class="action-edit" @click.stop.prevent="edit(props.item)">
                <v-icon color="surface-variant">mdi-pencil</v-icon>
              </v-btn>
            </td>
          </tr>
        </template>
      </v-data-table>

      <p class="text-caption pt-3 clickable" @click.stop.prevent="webdavDialog">
        <translate>Note:</translate>
        <translate>WebDAV clients, like Microsoft’s Windows Explorer or Apple's Finder, can connect directly to PhotoPrism. </translate>
        <translate>This mounts the originals folder as a network drive and allows you to open, edit, and delete files from your computer or smartphone as if they were local. </translate>
      </p>

      <v-form ref="form" validate-on="blur" class="p-form-settings" accept-charset="UTF-8" @submit.prevent="add">
        <div class="action-buttons">
          <v-btn v-if="user.hasWebDAV()" color="button" variant="flat" class="action-webdav-dialog compact" :block="$vuetify.display.xs" :disabled="isPublic || isDemo" @click.stop="webdavDialog">
            <translate>Connect via WebDAV</translate>
            <v-icon :end="!rtl" :start="rtl">mdi-swap-horizontal</v-icon>
          </v-btn>

          <v-btn color="primary-button" class="compact" :block="$vuetify.display.xs" :disabled="isPublic || isDemo" variant="flat" @click.stop="add">
            <translate>Connect</translate>
            <v-icon :end="!rtl" :start="rtl">mdi-plus</v-icon>
          </v-btn>
        </div>
      </v-form>
    </v-container>

    <p-service-add-dialog :show="dialog.add" @cancel="close('add')" @confirm="onAdded"></p-service-add-dialog>
    <p-service-remove-dialog :show="dialog.remove" :model="model" @cancel="close('remove')" @confirm="onRemoved"></p-service-remove-dialog>
    <p-service-edit-dialog :show="dialog.edit" :model="model" :scope="editScope" @remove="remove(model)" @cancel="close('edit')" @confirm="onEdited"></p-service-edit-dialog>
    <p-webdav-dialog :show="dialog.webdav" @close="dialog.webdav = false"></p-webdav-dialog>
  </div>
</template>

<script>
import Settings from "model/settings";
import Service from "model/service";
import { DateTime } from "luxon";

export default {
  name: "PSettingsServices",
  data() {
    return {
      isDemo: this.$config.get("demo"),
      isPublic: this.$config.get("public"),
      config: this.$config.values,
      readonly: this.$config.get("readonly"),
      settings: new Settings(this.$config.values.settings),
      model: {},
      results: [],
      labels: {},
      selected: [],
      user: this.$session.getUser(),
      dialog: {
        add: false,
        remove: false,
        webdav: false,
      },
      editScope: "main",
      listColumns: [
        { title: this.$gettext("Name"), key: "AccName", sortable: false, align: "left" },
        { title: this.$gettext("Upload"), key: "AccShare", sortable: false, align: "center" },
        { title: this.$gettext("Sync"), key: "AccSync", sortable: false, align: "center" },
        {
          title: this.$gettext("Last Sync"),
          key: "SyncDate",
          sortable: false,
          class: "hidden-sm-and-down",
          align: "left",
        },
        { title: "", key: "", sortable: false, class: "hidden-xs", align: "right" },
      ],
      rtl: this.$rtl,
    };
  },
  created() {
    if (this.isPublic && !this.isDemo) {
      this.$router.push({ name: "settings" });
    } else {
      this.load();
    }
  },
  methods: {
    webdavDialog() {
      this.dialog.webdav = true;
    },
    formatDate(d) {
      if (!d || !d.Valid) {
        return this.$gettext("Never");
      }

      const time = d.Time ? d.Time : d;

      return DateTime.fromISO(time).toLocaleString(DateTime.DATE_FULL);
    },
    load() {
      Service.search({ count: 2000 }).then((r) => (this.results = r.models));
    },
    remove(model) {
      this.model = model.clone();

      this.dialog.edit = false;
      this.dialog.remove = true;
    },
    onRemoved() {
      this.dialog.remove = false;
      this.model = {};
      this.load();
    },
    editSharing(model) {
      this.model = model.clone();

      this.editScope = "sharing";

      this.dialog.edit = true;
    },
    editSync(model) {
      this.model = model.clone();

      this.editScope = "sync";

      this.dialog.edit = true;
    },
    edit(model) {
      this.model = model.clone();

      this.editScope = "account";

      this.dialog.edit = true;
    },
    onEdited() {
      this.dialog.edit = false;
      this.model = {};
      this.load();
    },
    add() {
      this.dialog.add = true;
    },
    onAdded() {
      this.dialog.add = false;
      this.load();
    },
    close(name) {
      this.dialog[name] = false;
      this.model = {};
    },
  },
};
</script>
