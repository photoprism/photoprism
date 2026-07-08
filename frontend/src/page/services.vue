<template>
  <div ref="page" tabindex="-1" class="p-page p-page-settings p-settings-services" :class="$config.aclClasses('settings')">
    <v-form validate-on="invalid-input" class="p-services-search p-page__navigation" @submit.prevent="load()">
      <v-toolbar flat :density="$vuetify.display.smAndDown ? 'compact' : 'default'" color="secondary" class="page-toolbar">
        <v-text-field
          v-model="search"
          :density="density"
          hide-details
          clearable
          overflow
          single-line
          rounded="pill"
          variant="solo-filled"
          color="surface-variant"
          :placeholder="$gettext('Search')"
          prepend-inner-icon="mdi-magnify"
          autocomplete="off"
          autocorrect="off"
          autocapitalize="none"
          class="input-search input-search--focus background-inherit elevation-0"
          @keyup.enter="() => load()"
          @click:clear="() => load()"
        ></v-text-field>

        <p-action-menu v-if="$vuetify.display.mdAndUp" :items="menuActions" button-class="ms-1"></p-action-menu>
      </v-toolbar>
    </v-form>

    <v-data-table
      v-model="selected"
      :headers="listColumns"
      :items="results"
      :loading="loading"
      tile
      hover
      hide-default-footer
      item-key="ID"
      :no-data-text="$gettext('No services configured.')"
      :density="$vuetify.display.smAndDown ? 'compact' : 'default'"
      class="elevation-0 account-results list-view"
    >
      <template #item="props">
        <tr :data-name="props.item.AccName">
          <td class="p-account">
            <button class="text-primary text-break" @click.stop.prevent="edit(props.item)">
              {{ props.item.AccName }}
            </button>
          </td>
          <td class="text-center">
            <v-btn icon density="comfortable" variant="plain" :ripple="false" class="action-toggle-share" @click.stop.prevent="editSharing(props.item)">
              <v-icon :icon="props.item.AccShare ? 'mdi-check' : 'mdi-cog'" color="surface-variant"></v-icon>
            </v-btn>
          </td>
          <td class="text-center">
            <v-btn icon density="comfortable" variant="plain" :ripple="false" class="action-toggle-sync" @click.stop.prevent="editSync(props.item)">
              <v-icon v-if="props.item.AccErrors" color="surface-variant" :title="props.item.AccError">mdi-alert </v-icon>
              <v-icon v-else-if="props.item.AccSync" color="surface-variant">mdi-sync</v-icon>
              <v-icon v-else color="surface-variant">mdi-sync-off</v-icon>
            </v-btn>
          </td>
          <td class="hidden-sm-and-down">
            {{ formatDate(props.item.SyncDate) }}
          </td>
          <td class="hidden-xs text-end" nowrap>
            <v-btn
              icon="mdi-delete"
              color="surface-variant"
              density="comfortable"
              variant="plain"
              :ripple="false"
              class="action-remove action-secondary"
              @click.stop.prevent="remove(props.item)"
            ></v-btn>
            <v-btn
              icon="mdi-pencil"
              color="surface-variant"
              density="comfortable"
              variant="plain"
              :ripple="false"
              class="action-edit"
              @click.stop.prevent="edit(props.item)"
            ></v-btn>
          </td>
        </tr>
      </template>
    </v-data-table>
    <div class="pa-5">
      <p class="text-caption pb-1 clickable" @click.stop.prevent="webdavDialog">
        {{ $gettext(`Note:`) }}
        {{ $gettext(`WebDAV clients, like Microsoft’s Windows Explorer or Apple's Finder, can connect directly to PhotoPrism. `) }}
        {{
          $gettext(
            `This mounts the originals folder as a network drive and allows you to open, edit, and delete files from your computer or smartphone as if they were local. `
          )
        }}
      </p>

      <v-form ref="form" validate-on="invalid-input" class="p-form-settings" accept-charset="UTF-8" @submit.prevent="add">
        <div class="action-buttons">
          <v-btn
            v-if="user.hasWebDAV()"
            color="button"
            variant="flat"
            class="action-webdav-dialog"
            :block="$vuetify.display.xs"
            :disabled="isPublic || isDemo"
            @click.stop="webdavDialog"
          >
            {{ $gettext(`Connect via WebDAV`) }}
            <v-icon end>mdi-swap-horizontal</v-icon>
          </v-btn>

          <v-btn color="highlight" class="compact" :block="$vuetify.display.xs" :disabled="isPublic || isDemo" variant="flat" @click.stop="add">
            {{ $gettext(`Connect`) }}
            <v-icon icon="mdi-plus" end></v-icon>
          </v-btn>
        </div>
      </v-form>
    </div>

    <p-service-add :visible="dialog.add" @close="close('add')" @confirm="onAdded"></p-service-add>
    <p-service-remove :visible="dialog.remove" :model="model" @close="close('remove')" @confirm="onRemoved"></p-service-remove>
    <p-service-edit
      :visible="dialog.edit"
      :service="model"
      :scope="editScope"
      @remove="remove(model)"
      @close="close('edit')"
      @confirm="onEdited"
    ></p-service-edit>
    <p-settings-webdav :visible="dialog.webdav" @close="dialog.webdav = false"></p-settings-webdav>
  </div>
</template>

<script>
import Settings from "model/settings";
import Service from "model/service";
import { DateTime } from "luxon";
import links from "common/links";
import PActionMenu from "component/action/menu.vue";
import PServiceAdd from "component/service/add.vue";
import PServiceEdit from "component/service/edit.vue";
import PServiceRemove from "component/service/remove.vue";
import PSettingsWebdav from "component/settings/webdav.vue";

export default {
  name: "PPageServices",
  components: {
    PActionMenu,
    PServiceAdd,
    PServiceEdit,
    PServiceRemove,
    PSettingsWebdav,
  },
  data() {
    return {
      isDemo: this.$config.get("demo"),
      isPublic: this.$config.get("public"),
      config: this.$config.values,
      timeZone: this.$config.getTimeZone(),
      readonly: this.$config.get("readonly"),
      settings: new Settings(this.$config.values.settings),
      model: {},
      results: [],
      loading: false,
      search: "",
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
          headerProps: {
            class: "hidden-sm-and-down",
          },
          key: "SyncDate",
          sortable: false,
          align: "left",
        },
        {
          title: "",
          headerProps: {
            class: "hidden-xs",
          },
          key: "",
          sortable: false,
          align: "right",
        },
      ],
      rtl: this.$isRtl,
    };
  },
  computed: {
    // density returns the input density based on the current display size.
    density() {
      return this.$vuetify.display.smAndDown ? "compact" : "comfortable";
    },
  },
  created() {
    if (this.isPublic && !this.isDemo) {
      this.$router.push({ name: "settings" });
    } else {
      this.load();
    }
  },
  mounted() {
    this.$view.enter(this);
  },
  unmounted() {
    this.$view.leave(this);
  },
  methods: {
    // menuActions returns the toolbar overflow-menu entries (reload the list, open the sync docs).
    menuActions() {
      return [
        {
          name: "reload",
          icon: "mdi-refresh",
          text: this.$gettext("Reload"),
          visible: true,
          click: () => {
            this.load();
          },
        },
        {
          name: "docs",
          icon: "mdi-book-open-page-variant-outline",
          text: this.$gettext("Learn More"),
          visible: true,
          href: links.syncSettings,
          target: "_blank",
        },
      ];
    },
    // webdavDialog opens the WebDAV connection instructions dialog.
    webdavDialog() {
      this.dialog.webdav = true;
    },
    // formatDate renders a sync timestamp in the user's time zone, or "Never" if unset.
    formatDate(d) {
      if (!d || !d.Valid) {
        return this.$gettext("Never");
      }

      const time = d.Time ? d.Time : d;

      return DateTime.fromISO(time, { zone: this.timeZone }).toLocaleString(DateTime.DATE_FULL);
    },
    // load fetches the configured services from the API, filtered by the current search term.
    load() {
      const q = this.search ? this.search.trim() : "";
      this.loading = true;
      Service.search({ count: 2000, q })
        .then((r) => (this.results = r.models))
        .finally(() => (this.loading = false));
    },
    // remove opens the confirmation dialog for deleting a service.
    remove(model) {
      this.model = model.clone();

      this.dialog.edit = false;
      this.dialog.remove = true;
    },
    // onRemoved closes the remove dialog and reloads the list.
    onRemoved() {
      this.dialog.remove = false;
      this.model = {};
      this.load();
    },
    // editSharing opens the edit dialog focused on sharing options.
    editSharing(model) {
      this.model = model.clone();

      this.editScope = "sharing";

      this.dialog.edit = true;
    },
    // editSync opens the edit dialog focused on sync options.
    editSync(model) {
      this.model = model.clone();

      this.editScope = "sync";

      this.dialog.edit = true;
    },
    // edit opens the edit dialog focused on the account settings.
    edit(model) {
      this.model = model.clone();

      this.editScope = "account";

      this.dialog.edit = true;
    },
    // onEdited closes the edit dialog and reloads the list.
    onEdited() {
      this.dialog.edit = false;
      this.model = {};
      this.load();
    },
    // add opens the dialog for connecting a new service.
    add() {
      this.dialog.add = true;
    },
    // onAdded closes the add dialog and reloads the list.
    onAdded() {
      this.dialog.add = false;
      this.load();
    },
    // close hides the named dialog and clears the working model.
    close(name) {
      this.dialog[name] = false;
      this.model = {};
    },
  },
};
</script>
