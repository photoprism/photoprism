<template>
  <v-dialog
    :model-value="visible"
    persistent
    max-width="400"
    class="p-dialog p-service-upload"
    @keydown.esc="close"
    @after-enter="afterEnter"
    @after-leave="afterLeave"
  >
    <v-form ref="form" validate-on="invalid-input" accept-charset="UTF-8" tabindex="1" @submit.prevent>
      <v-card>
        <v-card-title class="d-flex justify-start align-center ga-3">
          <v-icon size="28" color="primary">mdi-cloud</v-icon>
          <h6 class="text-h6">{{ $gettext(`WebDAV Upload`) }}</h6>
        </v-card-title>
        <v-card-text class="dense">
          <v-row align="center" dense>
            <v-col cols="12">
              <v-select
                v-model="service"
                hide-details
                hide-no-data
                :label="$gettext('Account')"
                item-title="AccName"
                item-value="ID"
                return-object
                :disabled="loading || noServices"
                :items="services"
                @update:model-value="onChange"
              >
              </v-select>
            </v-col>
            <v-col cols="12">
              <v-autocomplete
                :model-value="service ? path : null"
                hide-details
                hide-no-data
                autocomplete="off"
                :hint="$gettext('Folder')"
                :search.sync="search"
                :items="pathItems"
                :loading="loading"
                :disabled="loading || noServices"
                item-title="abs"
                item-value="abs"
                :label="$gettext('Folder')"
                @update:model-value="setPath"
              >
              </v-autocomplete>
            </v-col>
          </v-row>
        </v-card-text>
        <v-card-actions class="action-buttons">
          <v-btn variant="flat" color="button" class="action-cancel action-close" @click.stop="close">
            {{ $gettext(`Cancel`) }}
          </v-btn>
          <v-btn
            v-if="noServices"
            :disabled="!canManage || (isPublic && !isDemo)"
            color="highlight"
            variant="flat"
            class="action-setup"
            @click.stop="setup"
          >
            {{ $gettext(`Setup`) }}
          </v-btn>
          <v-btn
            v-else
            :disabled="!canUpload || noServices || !service"
            color="highlight"
            variant="flat"
            class="action-upload"
            @click.stop="confirm"
          >
            {{ $gettext(`Upload`) }}
          </v-btn>
        </v-card-actions>
      </v-card>
    </v-form>
  </v-dialog>
</template>
<script>
import Service from "model/service";
import Selection from "common/selection";

export default {
  name: "PServiceUpload",
  props: {
    visible: {
      type: Boolean,
      default: false,
    },
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
      canManage: this.$config.feature("services") && this.$config.allow("services", "manage"),
      canUpload: this.$config.feature("services") && this.$config.allow("services", "upload"),
      noServices: false,
      loading: false,
      search: null,
      service: null,
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
      if (this.loading) {
        return;
      }

      const exists = this.paths.findIndex((p) => p.value === q);

      if (exists !== -1 || !q) {
        this.pathItems = this.paths;
        this.newPath = "";
      } else {
        this.newPath = q;
        this.pathItems = this.paths.concat([{ abs: q }]);
      }
    },
    visible: function (show) {
      if (show) {
        this.loading = false;
        this.load();
      } else if (this.selection) {
        this.selection.clear();
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
      this.$emit("close");
    },
    setup() {
      if (!this.canManage) {
        this.$notify.error(this.$gettext("Not allowed"));
        return;
      }

      this.$emit("close");

      this.$nextTick(() => {
        this.$router.push({ name: "settings_services" });
      });
    },
    confirm() {
      if (!this.canUpload) {
        this.$notify.error(this.$gettext("Not allowed"));
        return;
      } else if (this.noServices) {
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
    setPath(path) {
      this.path = path ? path : "/";
    },
    onChange() {
      if (this.loading) {
        return;
      }

      this.loading = true;

      this.paths = [{ abs: "/" }];
      this.pathItems = [];

      this.service
        .Folders()
        .then((paths) => {
          this.paths = [...paths];
          this.pathItems = [...paths];
          this.path = this.service.SharePath;
        })
        .finally(() => (this.loading = false));
    },
    load() {
      if (this.loading) {
        return;
      }

      this.loading = true;

      this.selection.clear().addItems(this.items);

      if (this.selection.isEmpty()) {
        this.selection.addModel(this.model);
      }

      if (this.selection.isEmpty()) {
        this.loading = false;
        this.$emit("close");
        return;
      }

      const params = {
        share: true,
        count: 2000,
        offset: 0,
      };

      Service.search(params)
        .then((response) => {
          this.loading = false;
          if (!response.models.length) {
            this.noServices = true;
            this.services.length = 0;
            this.service = null;
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
