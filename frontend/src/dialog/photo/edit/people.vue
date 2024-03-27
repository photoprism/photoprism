<template>
  <div class="p-tab p-tab-photo-people">
    <v-container grid-list-xs fluid class="pa-2 p-faces">
      <v-alert :value="markers.length === 0" color="secondary-dark" icon="lightbulb_outline" class="no-results ma-2 opacity-70" outline>
        <h3 class="body-2 ma-0 pa-0">
          <translate>No people found</translate>
        </h3>
        <p class="body-1 mt-2 mb-0 pa-0">
          <translate>You may rescan your library to find additional faces.</translate>
          <translate>Recognition starts after indexing has been completed.</translate>
        </p>
      </v-alert>
      <v-layout class="search-results face-results cards-view" row wrap fill-height>
        <v-flex v-for="(marker, index) in markers" :key="index" xs12 sm6 md3 xl2 d-flex>
          <v-card tile :data-id="marker.UID" style="user-select: none" :class="marker.classes()" class="result card">
            <div class="card-background card"></div>
            <v-img :src="marker.thumbnailUrl('tile_320')" :transition="false" aspect-ratio="1" class="card darken-1">
              <v-btn v-if="!marker.SubjUID && !marker.Invalid" :ripple="false" :depressed="false" class="input-reject" icon flat small absolute :title="$gettext('Remove')" @click.stop.prevent="onReject(marker)">
                <v-icon color="white" class="action-reject">clear</v-icon>
              </v-btn>
            </v-img>

            <v-card-actions class="card-details pa-0">
              <v-layout v-if="marker.Invalid" row wrap align-center>
                <v-flex xs12 class="text-xs-center pa-0">
                  <v-btn color="transparent" :disabled="busy" large depressed block :round="false" class="action-undo text-xs-center" :title="$gettext('Undo')" @click.stop="onApprove(marker)">
                    <v-icon dark>undo</v-icon>
                  </v-btn>
                </v-flex>
              </v-layout>
              <v-layout v-else-if="marker.SubjUID" row wrap align-center>
                <v-flex xs12 class="text-xs-left pa-0">
                  <v-text-field
                    v-model="marker.Name"
                    :rules="[textRule]"
                    :disabled="busy"
                    :readonly="true"
                    browser-autocomplete="off"
                    autocorrect="off"
                    class="input-name pa-0 ma-0"
                    hide-details
                    single-line
                    solo-inverted
                    clearable
                    clear-icon="eject"
                    @click:clear="onClearSubject(marker)"
                    @change="onRename(marker)"
                    @keyup.enter.native="onRename(marker)"
                  ></v-text-field>
                </v-flex>
              </v-layout>
              <v-layout v-else row wrap align-center>
                <v-flex xs12 class="text-xs-left pa-0">
                  <v-combobox
                    v-model="marker.Name"
                    style="z-index: 250"
                    :items="$config.values.people"
                    item-value="Name"
                    item-text="Name"
                    :disabled="busy"
                    :return-object="false"
                    :menu-props="menuProps"
                    :allow-overflow="false"
                    :hint="$gettext('Name')"
                    hide-details
                    single-line
                    solo-inverted
                    open-on-clear
                    hide-no-data
                    append-icon=""
                    prepend-inner-icon="person_add"
                    browser-autocomplete="off"
                    class="input-name pa-0 ma-0"
                    @change="onRename(marker)"
                    @keyup.enter.native="onRename(marker)"
                  >
                  </v-combobox>
                </v-flex>
              </v-layout>
            </v-card-actions>
          </v-card>
        </v-flex>
      </v-layout>
    </v-container>
  </div>
</template>

<script>
export default {
  name: "PTabPhotoPeople",
  props: {
    model: {
      type: Object,
      default: () => {},
    },
    uid: String,
  },
  data() {
    return {
      busy: false,
      markers: this.model.getMarkers(true),
      imageUrl: this.model.thumbnailUrl("fit_720"),
      disabled: !this.$config.feature("edit"),
      config: this.$config.values,
      readonly: this.$config.get("readonly"),
      menuProps: {
        closeOnClick: false,
        closeOnContentClick: true,
        openOnClick: false,
        maxHeight: 300,
      },
      textRule: (v) => {
        if (!v || !v.length) {
          return this.$gettext("Name");
        }

        return v.length <= this.$config.get("clip") || this.$gettext("Name too long");
      },
    };
  },
  methods: {
    refresh() {},
    onReject(marker) {
      if (this.busy || !marker) return;

      this.busy = true;
      this.$notify.blockUI();

      marker.reject().finally(() => {
        this.$notify.unblockUI();
        this.busy = false;
      });
    },
    onApprove(marker) {
      if (this.busy || !marker) return;

      this.busy = true;

      marker.approve().finally(() => (this.busy = false));
    },
    onClearSubject(marker) {
      if (this.busy || !marker) return;

      this.busy = true;
      this.$notify.blockUI();

      marker.clearSubject(marker).finally(() => {
        this.$notify.unblockUI();
        this.busy = false;
      });
    },
    onRename(marker) {
      if (this.busy || !marker) return;

      this.busy = true;
      this.$notify.blockUI();

      marker.rename().finally(() => {
        this.$notify.unblockUI();
        this.busy = false;
      });
    },
  },
};
</script>
