<template>
  <v-dialog :value="show" lazy persistent max-width="500" class="p-share-dialog" @keydown.esc="close">
    <v-card raised elevation="24">
      <v-card-title primary-title class="pb-0">
        <v-layout row wrap>
          <v-flex xs9>
            <h3 class="headline mb-0">
              <translate :translate-params="{ name: model.modelName() }">Share %{name}</translate>
            </h3>
          </v-flex>
          <v-flex xs3 :text-xs-right="!rtl" :text-xs-left="rtl">
            <v-btn icon flat dark color="secondary-dark" class="ma-0 action-add-link" :title="$gettext('Add Link')" @click.stop="add">
              <v-icon>add_link</v-icon>
            </v-btn>
          </v-flex>
        </v-layout>
      </v-card-title>
      <v-card-text>
        <v-expansion-panel class="pa-0 elevation-0">
          <v-expansion-panel-content v-for="(link, index) in links" :key="link.UID" class="pa-0 elevation-0 secondary mb-1">
            <template #header>
              <button :class="`text-xs-${!rtl ? 'left' : 'right'} action-url ml-0 mt-0 mb-0 pa-0 mr-2`" style="user-select: none" @click.stop="copyUrl(link)">
                <v-icon size="16" class="pr-1">link</v-icon>
                /s/<strong v-if="link.Token" style="font-weight: 500"> {{ link.getToken() }} </strong><span v-else>…</span>
              </button>
            </template>
            <v-card>
              <v-card-text class="secondary-light">
                <v-container fluid class="pa-0">
                  <v-layout row wrap>
                    <v-flex xs12 class="pa-2">
                      <v-text-field :value="link.url()" hide-details box flat readonly :label="$gettext('URL')" autocorrect="off" autocapitalize="none" browser-autocomplete="off" color="secondary-dark" class="input-url" @click.stop="selectText($event)"> </v-text-field>
                    </v-flex>
                    <v-flex xs12 sm6 class="pa-2">
                      <v-select v-model="link.Expires" hide-details box flat :label="expires(link)" browser-autocomplete="off" color="secondary-dark" item-text="text" item-value="value" :items="options.Expires()" class="input-expires"> </v-select>
                    </v-flex>
                    <v-flex xs12 sm6 class="pa-2">
                      <v-text-field v-model="link.Token" hide-details box flat required browser-autocomplete="off" autocorrect="off" autocapitalize="none" :label="$gettext('Secret')" :placeholder="$gettext('Token')" color="secondary-dark" class="input-secret"></v-text-field>
                    </v-flex>
                    <!-- v-flex xs12 sm6 class="pa-2">
                        <v-text-field
                                hide-details
                                browser-autocomplete="off"
                                :label="label.pass"
                                :placeholder="link.HasPassword ? '••••••••' : 'optional'"
                                color="secondary-dark"
                                v-model="link.Password"
                                :append-icon="showPassword ? 'visibility' : 'visibility_off'"
                                :type="showPassword ? 'text' : 'password'"
                                @click:append="showPassword = !showPassword"
                        ></v-text-field>
                    </v-flex -->
                    <v-flex xs6 :text-xs-left="!rtl" :text-xs-right="rtl" class="pa-2">
                      <v-btn small icon flat color="remove" class="ma-0 action-delete" :title="$gettext('Delete')" @click.stop.exact="remove(index)">
                        <v-icon>delete</v-icon>
                      </v-btn>
                    </v-flex>
                    <v-flex xs6 :text-xs-right="!rtl" :text-xs-left="rtl" class="pa-2">
                      <v-btn depressed dark color="primary-button" class="ma-0 compact action-save" @click.stop.exact="update(link)">
                        <translate>Save</translate>
                      </v-btn>
                    </v-flex>
                  </v-layout>
                </v-container>
              </v-card-text>
            </v-card>
          </v-expansion-panel-content>
        </v-expansion-panel>

        <v-container fluid :text-xs-left="!rtl" :text-xs-right="rtl" class="pb-0 pt-3 pr-0 pl-0 caption">
          <translate :translate-params="{ name: model.modelName() }">People you share a link with will be able to view public contents.</translate>
          <translate>A click will copy it to your clipboard.</translate>
          <translate>Any private photos and videos remain private and won't be shared.</translate>
          <translate>Alternatively, you can upload files directly to WebDAV servers like Nextcloud.</translate>
        </v-container>
      </v-card-text>
      <v-card-actions class="pt-0 px-3">
        <v-layout row wrap class="pa-2">
          <v-flex xs6>
            <v-btn depressed color="secondary-light" class="action-webdav" @click.stop="upload">
              <translate>WebDAV Upload</translate>
            </v-btn>
          </v-flex>
          <v-flex xs6 :text-xs-right="!rtl" :text-xs-left="rtl">
            <v-btn depressed color="secondary-light" class="action-close" @click.stop="confirm">
              <translate>Close</translate>
            </v-btn>
          </v-flex>
        </v-layout>
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
