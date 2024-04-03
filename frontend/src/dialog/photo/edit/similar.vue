<template>
  <div class="p-tab p-tab-photo-similar">
    <v-layout class="pa-2-md-and-up" row wrap align-top fill-height>
      <v-flex class="pa-2" xs12 md2 xxl1 fill-height>
        <p-photo-preview :model="model"></p-photo-preview>
      </v-flex>
      <v-flex class="pa-2" xs12 md10 xxl11 fill-width fill-height>
        <p class="subheading">
          Similar photos
        </p>
        <p-photo-cards :photos="results" :filter="filter"></p-photo-cards>
      </v-flex>
    </v-layout>
  </div>
</template>

<script>
import Photo from "model/photo";

export default {
  name: "PTabPhotoSimilar",
  props: {
    model: {
      type: Object,
      default: () => {},
    },
    uid: {
      type: String,
      default: "",
    },
  },
  data() {
    return {
      loading: true,
      results: [],
      filter: {
        order: "similar",
      }
    };
  },
  computed: {},
  methods: {
  },
  created() {
    this.loading = true;
    Photo.searchSimilar(this.model.UID)
      .then((response) => {
        this.results = response.models;
      })
      .finally(() => {
          this.loading = false;
        });
  }
};
</script>
