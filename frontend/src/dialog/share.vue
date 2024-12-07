<template>
  <v-dialog :model-value="show" persistent max-width="500" class="p-share-dialog" @keydown.esc="close">
    <v-card>
      <v-card-title class="pb-0">
        <v-row>
          <v-col cols="9">
            <h3 class="text-h5 mb-0">
              <translate :translate-params="{ name: model.modelName() }">Share %{name}</translate>
            </h3>
          </v-col>
          <v-col cols="3" :class="rtl ? 'text-left' : 'text-right'">
            <v-btn icon variant="text" color="surface-variant" class="ma-0 action-add-link" :title="$gettext('Add Link')" @click.stop="add">
              <v-icon>mdi-link-plus</v-icon>
            </v-btn>
          </v-col>
        </v-row>
      </v-card-title>
      <v-card-text>
        <v-expansion-panels class="pa-0 elevation-0">
          <v-expansion-panel v-for="(link, index) in links" :key="link.UID" class="pa-0 elevation-0 bg-secondary mb-1">
            <v-expansion-panel-title>
              <button :class="`text-${!rtl ? 'left' : 'right'} action-url ml-0 mt-0 mb-0 pa-0 mr-2`" style="user-select: none" @click.stop="copyUrl(link)">
                <v-icon size="16" class="pr-1">mdi-link</v-icon>
                /s/<strong v-if="link.Token" style="font-weight: 500"> {{ link.getToken() }} </strong><span v-else>…</span>
              </button>
            </v-expansion-panel-title>
            <v-expansion-panel-text>
              <v-card>
                <v-card-text class="secondary-light">
                  <v-container fluid class="pa-0">
                    <v-row>
                      <v-col cols="12" class="pa-2">
                        <v-text-field :model-value="link.url()" hide-details variant="solo" flat readonly :label="$gettext('URL')" autocorrect="off" autocapitalize="none" autocomplete="off" color="surface-variant" class="input-url" @click.stop="selectText($event)"> </v-text-field>
                      </v-col>
                      <v-col cols="12" sm="6" class="pa-2">
                        <v-select v-model="link.Expires" hide-details variant="solo" flat :label="expires(link)" browser-autocomplete="off" color="surface-variant" :items="options.Expires()" item-title="text" item-value="value" class="input-expires"> </v-select>
                      </v-col>
                      <v-col cols="12" sm="6" class="pa-2">
                        <v-text-field v-model="link.Token" hide-details variant="solo" flat required autocomplete="off" autocorrect="off" autocapitalize="none" :label="$gettext('Secret')" :placeholder="$gettext('Token')" color="surface-variant" class="input-secret"></v-text-field>
                      </v-col>
                      <!-- <v-col cols="12" sm="6" class="pa-2">
                        <v-text-field
                          v-model="link.Password"
                          hide-details
                          autocomplete="off"
                          :label="label.pass"
                          :placeholder="link.HasPassword ? '••••••••' : 'optional'"
                          color="surface-variant"
                          :append-icon="showPassword ? 'mdi-eye' : 'mdi-eye-off'"
                          :type="showPassword ? 'text' : 'password'"
                          @click:append="showPassword = !showPassword"
                        ></v-text-field>
                      </v-col> -->
                      <v-col cols="6" :class="!rtl ? 'text-left' : 'text-right'" class="pa-2">
                        <v-btn density="comfortable" icon variant="text" color="remove" class="ma-0 action-delete" :title="$gettext('Delete')" @click.stop.exact="remove(index)">
                          <v-icon>mdi-delete</v-icon>
                        </v-btn>
                      </v-col>
                      <v-col cols="6" :class="rtl ? 'text-left' : 'text-right'" class="pa-2">
                        <v-btn variant="flat" color="primary-button" class="ma-0 compact action-save" @click.stop.exact="update(link)">
                          <translate>Save</translate>
                        </v-btn>
                      </v-col>
                    </v-row>
                  </v-container>
                </v-card-text>
              </v-card>
            </v-expansion-panel-text>
          </v-expansion-panel>
        </v-expansion-panels>

        <v-container fluid :text-left="!rtl" :text-right="rtl" class="pb-0 pt-6 pr-0 pl-0 text-caption">
          <translate :translate-params="{ name: model.modelName() }">People you share a link with will be able to view public contents.</translate>
          <translate>A click will copy it to your clipboard.</translate>
          <translate>Any private photos and videos remain private and won't be shared.</translate>
          <translate>Alternatively, you can upload files directly to WebDAV servers like Nextcloud.</translate>
        </v-container>
      </v-card-text>
      <v-card-actions class="pt-0 px-6">
        <v-row class="pa-2">
          <v-col cols="6">
            <v-btn variant="flat" color="secondary-light" class="action-webdav" @click.stop="upload">
              <translate>WebDAV Upload</translate>
            </v-btn>
          </v-col>
          <v-col cols="6" :class="rtl ? 'text-left' : 'text-right'">
            <v-btn variant="flat" color="button" class="action-close" @click.stop="confirm">
              <translate>Close</translate>
            </v-btn>
          </v-col>
        </v-row>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>
<script>
import * as options from "options/options";
import Util from "common/util";

export default {
  name: "PShareDialog",
  props: {
    show: Boolean,
    model: {
      type: Object,
      default: () => {},
    },
  },
  data() {
    return {
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
      rtl: this.$rtl,
    };
  },
  watch: {
    show: function (show) {
      if (show) {
        this.links = [];
        this.loading = true;
        this.model
          .links()
          .then((resp) => {
            if (resp.count === 0) {
              this.add();
            } else {
              this.links = resp.models;
            }
          })
          .finally(() => (this.loading = false));
      }
    },
  },
  methods: {
    selectText(ev) {
      if (!ev || !ev.target) {
        return;
      }

      ev.target.select();
    },
    async copyUrl(link) {
      try {
        const url = link.url();
        await Util.copyToMachineClipboard(url);
        this.$notify.success(this.$gettext("Copied to clipboard"));
      } catch (error) {
        this.$notify.error(this.$gettext("Failed copying to clipboard"));
      }
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
