<template>
  <div class="p-tab p-tab-photo-people">
    <v-container grid-list-xs fluid class="pa-2 p-faces">
      <v-alert v-if="markers.length === 0" color="surface-variant" icon="mdi-lightbulb-outline" class="no-results ma-2 opacity-70" variant="outlined">
        <h3 class="text-subtitle-2 ma-0 pa-0">
          <translate>No people found</translate>
        </h3>
        <p class="mt-2 mb-0 pa-0">
          <translate>You may rescan your library to find additional faces.</translate>
          <translate>Recognition starts after indexing has been completed.</translate>
        </p>
      </v-alert>
      <v-row class="search-results face-results cards-view d-flex align-stretch ma-0">
        <v-col v-for="(marker, index) in markers" :key="index" cols="12" sm="6" md="3" xl="2" class="d-flex">
          <v-card tile :data-id="marker.UID" style="user-select: none" :class="marker.classes()" class="result card flex-grow-1">
            <div class="card-background card"></div>
            <v-img :src="marker.thumbnailUrl('tile_320')" :transition="false" aspect-ratio="1" class="card">
              <v-btn v-if="!marker.SubjUID && !marker.Invalid" :ripple="false" class="input-reject" icon variant="text" size="small" position="absolute" :title="$gettext('Remove')" @click.stop.prevent="onReject(marker)">
                <v-icon color="white" class="action-reject">mdi-close</v-icon>
              </v-btn>
            </v-img>

            <v-card-actions class="card-details pa-0">
              <v-row v-if="marker.Invalid" align="center">
                <v-col cols="12" class="text-center pa-0">
                  <v-btn :disabled="busy" size="large" variant="flat" block :rounded="false" class="action-undo text-center" :title="$gettext('Undo')" @click.stop="onApprove(marker)">
                    <!-- TODO: change this icon -->
                    <v-icon>undo</v-icon>
                  </v-btn>
                </v-col>
              </v-row>
              <v-row v-else-if="marker.SubjUID" align="center">
                <v-col cols="12" class="text-left pa-0">
                  <v-text-field
                    v-model="marker.Name"
                    :rules="[textRule]"
                    :disabled="busy"
                    :readonly="true"
                    autocomplete="off"
                    autocorrect="off"
                    class="input-name pa-0 ma-0"
                    hide-details
                    single-line
                    variant="solo-inverted"
                    clearable
                    clear-icon="mdi-eject"
                    @click:clear="onClearSubject(marker)"
                    @change="onRename(marker)"
                    @keyup.enter="onRename(marker)"
                  ></v-text-field>
                </v-col>
              </v-row>
              <v-row v-else align="center">
                <v-col cols="12" class="text-left pa-0">
                  <!-- TODO: check property allow-overflow TEST -->
                  <v-combobox
                    v-model="marker.Name"
                    style="z-index: 250"
                    :items="$config.values.people"
                    item-title="Name"
                    item-value="Name"
                    :disabled="busy"
                    :return-object="false"
                    :menu-props="menuProps"
                    :hint="$gettext('Name')"
                    hide-details
                    single-line
                    variant="solo-inverted"
                    open-on-clear
                    hide-no-data
                    append-icon=""
                    prepend-inner-icon="mdi-account-plus"
                    autocomplete="off"
                    class="input-name pa-0 ma-0"
                    @update:model-value="onRename(marker)"
                    @keyup.enter.native="onRename(marker)"
                  >
                  </v-combobox>
                </v-col>
              </v-row>
            </v-card-actions>
          </v-card>
        </v-col>
      </v-row>
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
