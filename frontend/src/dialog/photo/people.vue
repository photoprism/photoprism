<template>
  <div class="p-tab p-tab-photo-people">
    <v-container grid-list-xs fluid class="pa-2 p-faces">
      <v-card v-if="markers.length === 0" class="no-results secondary-light lighten-1 ma-1" flat>
        <v-card-title primary-title>
          <div>
            <h3 class="title ma-0 pa-0">
              <translate>Couldn't find any people</translate>
            </h3>
            <p class="mt-4 mb-0 pa-0">
              <translate>To automatically detect faces, please re-index your library and wait until indexing has been completed.</translate>
            </p>
          </div>
        </v-card-title>
      </v-card>
      <v-layout row wrap class="search-results photo-results cards-view">
        <v-flex
            v-for="(marker, index) in markers"
            :key="index"
            xs6 sm4 md3 lg2 xl1 d-flex
        >
          <v-card tile
                  :data-id="marker.ID"
                  style="user-select: none"
                  class="result accent lighten-3">
            <div class="card-background accent lighten-3"></div>
            <canvas :id="'face-' + marker.ID" :key="marker.ID" width="300" height="300" style="width: 100%" class="v-responsive v-image accent lighten-2"></canvas>

            <v-card-actions v-if="marker.Score < 30" class="card-details pa-0">
              <v-layout row wrap align-center>
                <v-flex xs6 class="text-xs-center pa-1">
                  <v-btn color="accent lighten-2"
                         small depressed dark block :round="false"
                         class="action-archive text-xs-center"
                         :title="$gettext('Reject')" @click.stop="reject(marker)">
                    <v-icon dark>clear</v-icon>
                  </v-btn>
                </v-flex>
                <v-flex xs6 class="text-xs-center pa-1">
                  <v-btn color="accent lighten-2"
                         small depressed dark block :round="false"
                         class="action-approve text-xs-center"
                         :title="$gettext('Approve')" @click.stop="confirm(marker)">
                    <v-icon dark>check</v-icon>
                  </v-btn>
                </v-flex>
              </v-layout>
            </v-card-actions>

            <!-- v-card-title primary-title class="pa-3 card-details" style="user-select: none;">
              <div>
                <h3 class="body-2 mb-2">
                  <button class="action-title-edit" :data-uid="marker.ID"
                          @click.exact="alert('Name')">
                    Jens Mander
                  </button>
                </h3>
              </div>
            </v-card-title -->
          </v-card>
        </v-flex>
      </v-layout>
    </v-container>
  </div>
</template>

<script>
export default {
  name: 'PTabPhotoPeople',
  props: {
    model: Object,
    uid: String,
  },
  data() {
    return {
      markers: this.model.getMarkers(),
      imageUrl: this.model.thumbnailUrl("fit_720"),
      disabled: !this.$config.feature("edit"),
      config: this.$config.values,
      readonly: this.$config.get("readonly"),
    };
  },
  mounted () {
    this.markers.forEach((m) => {
      const canvas = document.getElementById('face-' + m.ID);

      let ctx = canvas.getContext('2d');
      let img = new Image();

      img.onload = function() {
        const w = Math.round(m.W * img.width);
        const h = Math.round(m.H * img.height);
        const s = w > h ? w : h;
        const x = Math.round((m.X - (m.W / 2)) * img.width);
        const y = Math.round((m.Y - (m.H / 2)) * img.height);

        ctx.drawImage(img, x, y, s, s, 0, 0, 300, 300);
      };

      if (m.W < 0.07) {
        // TODO: Not all users have thumbs with this resolution.
        img.src = this.model.thumbnailUrl("fit_7680");
      } else if (m.W < 0.1) {
        // TODO: Not all users have thumbs with this resolution.
        img.src = this.model.thumbnailUrl("fit_2048");
      } else if (m.W < 0.15) {
        // TODO: Not all users have thumbs with this resolution.
        img.src = this.model.thumbnailUrl("fit_1280");
      } else {
        img.src = this.imageUrl;
      }
    });
  },
  methods: {
    refresh() {
    },
    reject(marker) {
      this.$notify.warn("Work in progress");
    },
    confirm(marker) {
      this.$notify.warn("Work in progress");
    },
  },
};
</script>
