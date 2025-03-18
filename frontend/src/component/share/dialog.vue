<template>
  <v-dialog
    :model-value="visible"
    persistent
    max-width="540"
    class="p-dialog p-share-dialog"
    @keydown.esc="close"
    @after-enter="afterEnter"
    @after-leave="afterLeave"
  >
    <v-card>
      <v-card-title class="d-flex justify-space-between align-center ga-3">
        <h6 class="text-h6">{{ $gettext(`Share %{s}`, { s: model.modelName() }) }}</h6>
        <v-btn
          icon="mdi-link-plus"
          variant="text"
          color="primary"
          :title="$gettext('Add Link')"
          class="action-add-link"
          @click.stop="add"
        ></v-btn>
      </v-card-title>
      <v-card-text>
        <v-expansion-panels
          v-model="expanded"
          variant="accordion"
          density="compact"
          rounded="6"
          tabindex="1"
          class="elevation-0"
        >
          <v-expansion-panel v-for="(link, index) in links" :key="link.UID" color="secondary" class="pa-0 elevation-0">
            <v-expansion-panel-title class="d-flex justify-start align-center ga-3 text-body-2 px-4">
              <v-icon icon="mdi-link"></v-icon>
              <div class="text-start not-selectable action-url d-inline-flex">
                /s/<strong v-if="link.Token" style="font-weight: 500">{{ link.getToken() }}</strong
                ><span v-else>…</span>
              </div>
            </v-expansion-panel-title>
            <v-expansion-panel-text>
              <v-card color="secondary-light">
                <v-card-text class="dense">
                  <v-row align="center" dense>
                    <v-col cols="12">
                      <v-text-field
                        :model-value="link.url()"
                        readonly
                        flat
                        hide-details
                        density="comfortable"
                        variant="solo"
                        autocorrect="off"
                        autocapitalize="none"
                        autocomplete="off"
                        append-inner-icon="mdi-content-copy"
                        class="input-url cursor-copy"
                        @click.stop="$util.copyText(link.url())"
                      >
                      </v-text-field>
                    </v-col>
                    <v-col cols="12" sm="6">
                      <v-select
                        v-model="link.Expires"
                        hide-details
                        density="comfortable"
                        variant="solo"
                        flat
                        :label="expires(link)"
                        browser-autocomplete="off"
                        :items="options.Expires()"
                        item-title="text"
                        item-value="value"
                        class="input-expires"
                      >
                      </v-select>
                    </v-col>
                    <v-col cols="12" sm="6">
                      <v-text-field
                        v-model="link.Token"
                        hide-details
                        density="comfortable"
                        variant="solo"
                        flat
                        autocomplete="off"
                        autocorrect="off"
                        autocapitalize="none"
                        :label="$gettext('Secret')"
                        :placeholder="$gettext('Token')"
                        class="input-secret"
                      ></v-text-field>
                    </v-col>
                    <!-- <v-col cols="12" sm="6" class="pa-2">
                      <v-text-field
                        v-model="link.Password"
                        hide-details
                        autocomplete="off"
                        :label="label.pass"
                        :placeholder="link.HasPassword ? '••••••••' : 'optional'"
                        color="surface-variant"
                        :append-inner-icon="showPassword ? 'mdi-eye-off' : 'mdi-eye'"
                        :type="showPassword ? 'text' : 'password'"
                        @click:append-inner="showPassword = !showPassword"
                      ></v-text-field>
                    </v-col> -->
                    <v-col cols="12" class="d-flex justify-space-between align-center ga-3">
                      <v-btn
                        variant="text"
                        color="remove"
                        density="comfortable"
                        icon="mdi-delete"
                        :title="$gettext('Delete')"
                        class="action-delete"
                        @click.stop.exact="remove(index)"
                      >
                      </v-btn>
                      <v-btn variant="flat" color="highlight" class="action-save" @click.stop.exact="update(link)">
                        {{ $gettext(`Save`) }}
                      </v-btn>
                    </v-col>
                  </v-row>
                </v-card-text>
              </v-card>
            </v-expansion-panel-text>
          </v-expansion-panel>
        </v-expansion-panels>

        <div class="pt-3 text-caption">
          {{
            $gettext(`People you share a link with will be able to view public contents.`, { name: model.modelName() })
          }}
          {{ $gettext(`A click will copy it to your clipboard.`) }}
          {{ $gettext(`Any private photos and videos remain private and won't be shared.`) }}
          {{ $gettext(`Alternatively, you can upload files directly to WebDAV servers like Nextcloud.`) }}
        </div>
      </v-card-text>
      <v-card-actions>
        <v-btn variant="flat" color="button" class="action-webdav" @click.stop="upload">
          {{ $gettext(`WebDAV Upload`) }}
        </v-btn>
        <v-btn variant="flat" color="button" class="action-close" @click.stop="confirm">
          {{ $gettext(`Close`) }}
        </v-btn>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>
<script>
import * as options from "options/options";

export default {
  name: "PShareDialog",
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
  data() {
    return {
      expanded: [0],
      host: window.location.host,
      showPassword: false,
      loading: false,
      search: null,
      links: [],
      options: options,
      label: {
        url: this.$gettext("Service URL"),
        user: this.$gettext("Username"),
        pass: this.$gettext("Password"),
        cancel: this.$gettext("Cancel"),
        confirm: this.$gettext("Done"),
      },
      rtl: this.$isRtl,
    };
  },
  watch: {
    visible: function (show) {
      if (show) {
        this.links = [];
        this.loading = true;
        this.expanded = [];
        this.model
          .links()
          .then((resp) => {
            if (resp.count === 0) {
              this.add();
            } else {
              this.links = resp.models;
              this.expanded = [0];
            }
          })
          .finally(() => (this.loading = false));
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
    expires(link) {
      let result = this.$gettext("Expires");

      if (link.Expires <= 0) {
        return result;
      }

      return `${result}: ${link.expires()}`;
    },
    add() {
      this.loading = true;

      this.model
        .createLink()
        .then((r) => {
          this.links.push(r);
          this.expanded = [this.links.length - 1];
        })
        .finally(() => (this.loading = false));
    },
    update(link) {
      if (!link) {
        this.$notify.error(this.$gettext("Failed updating link"));
        return;
      }

      this.loading = true;

      this.model
        .updateLink(link)
        .then(() => {
          this.$notify.success(this.$gettext("Changes successfully saved"));
        })
        .finally(() => {
          this.loading = false;
        });
    },
    remove(index) {
      const link = this.links[index];

      if (!link) {
        this.$notify.error(this.$gettext("Failed removing link"));
        return;
      }

      this.loading = true;

      this.model
        .removeLink(link)
        .then(() => {
          this.$notify.success(this.$gettext("Changes successfully saved"));
          this.links.splice(index, 1);
        })
        .finally(() => (this.loading = false));
    },
    upload() {
      this.$emit("upload");
    },
    close() {
      this.$emit("close");
    },
    confirm() {
      this.$emit("close");
    },
  },
};
</script>
