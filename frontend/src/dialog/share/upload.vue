<template>
  <v-dialog :model-value="show" persistent max-width="400" class="p-share-upload-dialog" @keydown.esc="cancel">
    <v-card>
      <v-card-title class="d-flex justify-start align-center ga-3">
            <h6 class="text-h6"><translate>WebDAV Upload</translate></h6>
            <v-btn icon variant="text" color="primary" @click.stop="setup">
              <v-icon>mdi-cloud</v-icon>
            </v-btn>
      </v-card-title>
      <v-card-text class="dense">
        <v-row dense>
          <v-col cols="12" class="text-left">
            <v-select v-model="service" hide-details hide-no-data :label="$gettext('Account')" item-title="AccName" item-value="ID" return-object :disabled="loading || noServices" :items="services" @update:model-value="onChange"> </v-select>
          </v-col>
          <v-col cols="12" class="text-left">
            <v-autocomplete
              v-model="path"
              hide-details
              hide-no-data
              autocomplete="off"
              hint="Folder"
              :search.sync="search"
              :items="pathItems"
              :loading="loading"
              :disabled="loading || noServices"
              item-title="abs"
              item-value="abs"
              :label="$gettext('Folder')"
            >
            </v-autocomplete>
          </v-col>
        </v-row>
      </v-card-text>
      <v-card-actions>
        <v-btn variant="flat" color="button" class="action-cancel" @click.stop="cancel">
          <translate>Cancel</translate>
        </v-btn>
        <v-btn v-if="noServices" :disabled="isPublic && !isDemo" color="primary-button" variant="flat" class="action-setup" @click.stop="setup">
          <translate>Setup</translate>
        </v-btn>
        <v-btn v-else :disabled="noServices" color="primary-button" variant="flat" class="action-upload" @click.stop="confirm">
          <translate>Upload</translate>
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>
<script>
import Service from "model/service";
import Selection from "common/selection";

export default {
  name: "PShareUploadDialog",
  props: {
    show: Boolean,
    items: {
      type: Object,
      default: null,
    },
    model: {
      type: Object,
      default: null,
    },
  },
  data() {
    return {
      isDemo: this.$config.get("demo"),
      isPublic: this.$config.get("public"),
      noServices: false,
      loading: true,
      search: null,
      service: {},
      services: [],
      selection: new Selection({}),
      path: "/",
      paths: [{ abs: "/" }],
      pathItems: [],
      newPath: "",
    };
  },
  watch: {
    search(q) {
      if (this.loading) return;

      const exists = this.paths.findIndex((p) => p.value === q);

      if (exists !== -1 || !q) {
        this.pathItems = this.paths;
        this.newPath = "";
      } else {
        this.newPath = q;
        this.pathItems = this.paths.concat([{ abs: q }]);
      }
    },
    show: function (show) {
      if (show) {
        this.load();
      } else if (this.selection) {
        this.selection.clear();
      }
    },
  },
  methods: {
    cancel() {
      this.$emit("cancel");
    },
    setup() {
      this.$router.push({ name: "settings_services" });
    },
    confirm() {
      if (this.noServices) {
        this.$notify.warn(this.$gettext("No servers configured."));
        return;
      } else if (this.loading) {
        this.$notify.busy();
        return;
      }

      this.loading = true;
      this.service
        .Upload(this.selection, this.path)
        .then((files) => {
          this.loading = false;

          if (files.length === 1) {
            this.$notify.success(this.$gettext("One file uploaded"));
          } else {
            this.$notify.success(this.$gettextInterpolate(this.$gettext("%{n} files uploaded"), { n: files.length }));
          }

          this.$emit("confirm", this.service);
        })
        .catch(() => (this.loading = false));
    },
    onChange() {
      this.paths = [{ abs: "/" }];

      this.loading = true;
      this.service
        .Folders()
        .then((p) => {
          for (let i = 0; i < p.length; i++) {
            this.paths.push(p[i]);
          }

          this.pathItems = [...this.paths];
          this.path = this.service.SharePath;
        })
        .finally(() => (this.loading = false));
    },
    load() {
      this.loading = true;

      this.selection.clear().addItems(this.items);

      if (this.selection.isEmpty()) {
        this.selection.addModel(this.model);
      }

      if (this.selection.isEmpty()) {
        this.loading = false;
        this.$emit("cancel");
        return;
      }

      const params = {
        share: true,
        count: 2000,
        offset: 0,
      };

      Service.search(params)
        .then((response) => {
          if (!response.models.length) {
            this.noServices = true;
            this.loading = false;
          } else {
            this.service = response.models[0];
            this.services = response.models;
            this.onChange();
          }
        })
        .catch(() => (this.loading = false));
    },
  },
};
</script>
