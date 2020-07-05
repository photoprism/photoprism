<template>
  <v-dialog lazy v-model="show" persistent max-width="500" class="p-share-dialog" @keydown.esc="close">
    <v-card raised elevation="24">
      <v-card-title primary-title class="pb-0">
        <v-layout row wrap>
          <v-flex xs9>
            <h3 class="headline mb-0"><translate :translate-params="{name: model.modelName()}">Share %{name}</translate></h3>
          </v-flex>
          <v-flex xs3 text-xs-right>
            <v-btn icon flat dark color="secondary-dark" class="ma-0 action-add-link" @click.stop="add" :title="$gettext('Add Link')">
              <v-icon>add_link</v-icon>
            </v-btn>
          </v-flex>
        </v-layout>
      </v-card-title>
      <v-card-text>
        <v-expansion-panel class="pa-0 elevation-0">
          <v-expansion-panel-content v-for="(link, index) in links" :key="index"
                                     class="pa-0 elevation-0 secondary-light mb-1">
            <template v-slot:header>
              <button class="text-xs-left action-url ml-0 mt-0 mb-0 pa-0 mr-2" @click.stop="copyUrl(link)" style="user-select: none;">
                <v-icon size="16" class="pr-1">link</v-icon>
                /s/<strong style="font-weight: 500;" v-if="link.Token">{{ link.getToken() }}</strong><span v-else>…</span>
              </button>
            </template>
            <v-card>
              <v-card-text class="grey lighten-4">
                <v-container fluid class="pa-0">
                  <v-layout row wrap>
                    <v-flex xs12 class="pa-2">
                      <v-text-field
                              :label="$gettext('URL')"
                              browser-autocomplete="off"
                              hide-details readonly
                              color="secondary-dark"
                              @click.stop="selectText($event)"
                              v-model="link.url()"
                              class="input-url">
                      </v-text-field>
                    </v-flex>
                    <v-flex xs12 sm6 class="pa-2">
                      <v-select
                              :label="expires(link)"
                              browser-autocomplete="off"
                              hide-details
                              color="secondary-dark"
                              item-text="text"
                              item-value="value"
                              v-model="link.Expires"
                              :items="items.expires"
                              class="input-expires"
                      >
                      </v-select>
                    </v-flex>
                    <v-flex xs12 sm6 class="pa-2">
                      <v-text-field
                              hide-details required
                              browser-autocomplete="off"
                              :label="$gettext('Secret')"
                              :placeholder="$gettext('Token')"
                              color="secondary-dark"
                              v-model="link.Token"
                              class="input-secret"
                      ></v-text-field>
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
                    <v-flex xs6 text-xs-left class="pa-2">
                      <v-btn small icon flat color="remove" class="ma-0 action-delete"
                             @click.stop.exact="remove(index)" :title="$gettext('Delete')">
                        <v-icon>delete</v-icon>
                      </v-btn>
                    </v-flex>
                    <v-flex xs6 text-xs-right class="pa-2">
                      <v-btn small depressed dark color="secondary-dark" class="ma-0 action-save"
                             @click.stop.exact="update(link)">
                        <translate>Save</translate>
                      </v-btn>
                    </v-flex>
                  </v-layout>
                </v-container>
              </v-card-text>
            </v-card>
          </v-expansion-panel-content>
        </v-expansion-panel>

        <v-container fluid text-xs-left class="pb-0 pt-3 pr-0 pl-0 caption">
          <translate :translate-params="{name: model.modelName()}">People you share a link with will be able to view public contents.</translate>
          <translate>A click will copy it to your clipboard.</translate>
          <translate>Any private photos and videos remain private and won't be shared.</translate>
          <translate>Alternatively, you can upload files directly to WebDAV servers like Nextcloud.</translate>
        </v-container>
      </v-card-text>
      <v-card-actions class="pt-0">
        <v-layout row wrap class="pa-2">
          <v-flex xs6>
            <v-btn @click.stop="upload" depressed color="secondary-light"
                   class="action-webdav">
              <translate>WebDAV Upload</translate>
            </v-btn>
          </v-flex>
          <v-flex xs6 text-xs-right>
            <v-btn depressed color="secondary-light" @click.stop="confirm"
                   class="action-close">
              <translate>Close</translate>
            </v-btn>
          </v-flex>
        </v-layout>
      </v-card-actions>
    </v-card>
  </v-dialog>
</template>
<script>

    export default {
        name: 'p-share-dialog',
        props: {
            show: Boolean,
            model: Object,
        },
        data() {
            return {
                host: window.location.host,
                showPassword: false,
                loading: false,
                search: null,
                links: [],
                items: {
                    expires: [
                        {"value": 0, "text": "Never"},
                        {"value": 86400, "text": "After 1 day"},
                        {"value": 86400 * 3, "text": "After 3 days"},
                        {"value": 86400 * 7, "text": "After 7 days"},
                        {"value": 86400 * 14, "text": "After two weeks"},
                        {"value": 86400 * 31, "text": "After one month"},
                        {"value": 86400 * 60, "text": "After two months"},
                        {"value": 86400 * 365, "text": "After one year"},
                    ],
                },
                label: {
                    url: this.$gettext("Service URL"),
                    user: this.$gettext("Username"),
                    pass: this.$gettext("Password"),
                    cancel: this.$gettext("Cancel"),
                    confirm: this.$gettext("Done"),
                }
            }
        },
        methods: {
            selectText(ev) {
                if(!ev || !ev.target) {
                    return;
                }

                ev.target.select();
            },
            copyUrl(link) {
                window.navigator.clipboard.writeText(link.url())
                    .then(() => this.$notify.success(this.$gettext("Copied to clipboard")), () => this.$notify.error(this.$gettext("Failed copying to clipboard")));
            },
            expires(link) {
                let result = this.$gettext('Expires');

                if (link.Expires <= 0) {
                    return result
                }

                return `${result}: ${link.expires()}`;
            },
            add() {
                this.loading = true;

                this.model.createLink().then((r) => {
                    this.links.push(r);
                }).finally(() => this.loading = false)
            },
            update(link) {
                if (!link) {
                    this.$notify.error(this.$gettext("Failed updating link"))
                    return;
                }

                this.loading = true;

                this.model.updateLink(link).finally(() => this.loading = false);
            },
            remove(index) {
                const link = this.links[index];

                if (!link) {
                    this.$notify.error(this.$gettext("Failed removing link"))
                    return;
                }

                this.loading = true;

                this.model.removeLink(link).then(() => {
                    this.links.splice(index, 1);
                }).finally(() => this.loading = false)
            },
            upload() {
                this.$emit('upload');
            },
            close() {
                this.$emit('close');
            },
            confirm() {
                this.$emit('close');
            },
        },
        watch: {
            show: function (show) {
                if (show) {
                    this.links = [];
                    this.loading = true;
                    this.model.links().then((resp) => {
                        if (resp.count === 0) {
                            this.add();
                        } else {
                            this.links = resp.models;
                        }
                    }).finally(() => this.loading = false);
                }
            }
        },
    }
</script>
