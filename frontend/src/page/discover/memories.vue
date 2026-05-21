<template>
  <div class="p-tab p-tab-discover-memories">
    <p-page-photos :static-filter="staticFilter" :embedded="true" />
  </div>
</template>

<script>
import { todayMemoriesFilter } from "common/memories";
import PPagePhotos from "page/photos.vue";

export default {
  name: "PTabDiscoverMemories",
  components: {
    PPagePhotos,
  },
  data() {
    return {
      staticFilter: todayMemoriesFilter(),
    };
  },
};
</script>
