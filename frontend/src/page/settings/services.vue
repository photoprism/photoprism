<template>
  <div class="p-tab p-settings-services">
    <v-data-table
        v-model="selected"
        :headers="listColumns"
        :items="results"
        hide-actions
        disable-initial-sort
        class="elevation-0 account-results list-view"
        item-key="ID"
        :no-data-text="$gettext('No services configured.')"
    >
      <template #items="props">
        <tr :data-name="props.item.AccName">
          <td class="p-account">
            <button class="secondary-dark--text text--lighten-1" @click.stop.prevent="edit(props.item)">
              {{ props.item.AccName }}
            </button>
          </td>
          <td class="text-xs-center">
            <v-btn icon small flat :ripple="false"
                   class="action-toggle-share"
                   color="transparent"
                   @click.stop.prevent="editSharing(props.item)">
              <v-icon v-if="props.item.AccShare" color="secondary-dark">check</v-icon>
              <v-icon v-else color="secondary-dark">settings</v-icon>
            </v-btn>
          </td>
          <td class="text-xs-center">
            <v-btn icon small flat :ripple="false"
                   class="action-toggle-sync"
                   color="transparent"
                   @click.stop.prevent="editSync(props.item)">
              <v-icon v-if="props.item.AccErrors" color="secondary-dark" :title="props.item.AccError">report_problem
              </v-icon>
              <v-icon v-else-if="props.item.AccSync" color="secondary-dark">sync</v-icon>
              <v-icon v-else color="secondary-dark">sync_disabled</v-icon>
            </v-btn>
          </td>
          <td class="hidden-sm-and-down">{{ formatDate(props.item.SyncDate) }}</td>
          <td class="hidden-xs-only text-xs-right" nowrap>
            <v-btn icon small flat :ripple="false"
                   class="action-remove action-secondary"
                   color="transparent"
                   @click.stop.prevent="remove(props.item)">
              <v-icon color="secondary-dark">delete</v-icon>
            </v-btn>
            <v-btn icon small flat :ripple="false"
                   class="action-edit"
                   color="transparent"
                   @click.stop.prevent="edit(props.item)">
              <v-icon color="secondary-dark">edit</v-icon>
            </v-btn>
          </td>
        </tr>
      </template>
    </v-data-table>
    <v-container fluid>
      <p class="caption pa-0 clickable" @click.stop.prevent="webdavDialog">
        <translate>Note:</translate>
        <translate>WebDAV clients, like Microsoft’s Windows Explorer or Apple's Finder, can connect directly to
          PhotoPrism.
        </translate>
        <translate>This mounts the originals folder as a network drive and allows you to open, edit, and delete files
          from your computer or smartphone as if they were local.
        </translate>
      </p>

      <v-form ref="form" lazy-validation
              dense class="p-form-settings mt-2" accept-charset="UTF-8"
              @submit.prevent="add">

        <v-btn v-if="user.WebDAV" depressed color="secondary-light" class="action-webdav-dialog compact ml-0 my-2 mr-2"
               :block="$vuetify.breakpoint.xsOnly"
               :disabled="isPublic || isDemo" @click.stop="webdavDialog">
          <translate>Connect via WebDAV</translate>
          <v-icon :right="!rtl" :left="rtl" dark>sync_alt</v-icon>
        </v-btn>

        <v-btn color="primary-button"
               class="white--text compact ml-0 my-2 mr-2"
               :block="$vuetify.breakpoint.xsOnly"
               :disabled="isPublic || isDemo"
               depressed @click.stop="add">
          <translate>Connect</translate>
          <v-icon :right="!rtl" :left="rtl" dark>add</v-icon>
        </v-btn>
      </v-form>
    </v-container>

    <p-service-add-dialog :show="dialog.add" @cancel="onCancel('add')"
                          @confirm="onAdded"></p-service-add-dialog>
    <p-service-remove-dialog :show="dialog.remove" :model="model" @cancel="onCancel('remove')"
                             @confirm="onRemoved"></p-service-remove-dialog>
    <p-service-edit-dialog :show="dialog.edit" :model="model" :scope="editScope" @remove="remove(model)"
                           @cancel="onCancel('edit')"
                           @confirm="onEdited"></p-service-edit-dialog>
    <p-webdav-dialog :show="dialog.webdav" @close="dialog.webdav = false"></p-webdav-dialog>
  </div>
</template>

<script>
import Settings from "model/settings";
import Service from "model/service";
import {DateTime} from "luxon";

export default {
  name: 'PSettingsServices',
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
        {text: this.$gettext('Name'), value: 'AccName', sortable: false, align: 'left'},
        {text: this.$gettext('Upload'), value: 'AccShare', sortable: false, align: 'center'},
        {text: this.$gettext('Sync'), value: 'AccSync', sortable: false, align: 'center'},
        {
          text: this.$gettext('Last Sync'),
          value: 'SyncDate',
          sortable: false,
          class: 'hidden-sm-and-down',
          align: 'left'
        },
        {text: '', value: '', sortable: false, class: 'hidden-xs-only', align: 'right'},
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
        return this.$gettext('Never');
      }

      const time = d.Time ? d.Time : d;

      return DateTime.fromISO(time).toLocaleString(DateTime.DATE_FULL);
    },
    load() {
      Service.search({count: 2000}).then(r => this.results = r.models);
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
    onCancel(name) {
      this.dialog[name] = false;
      this.model = {};
    },
  },
};
</script>
