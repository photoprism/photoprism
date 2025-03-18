// Default settings for standard components.
import IconLivePhoto from "./icon/live-photo.vue";
import IconSponsor from "./icon/sponsor.vue";
import IconPrism from "./icon/prism.vue";

// Additional icons for use with Vuetify.
export default {
  live_photo: {
    component: IconLivePhoto,
    props: {
      name: "live_photo",
    },
  },
  sponsor: {
    component: IconSponsor,
    props: {
      name: "sponsor",
    },
  },
  prism: {
    component: IconPrism,
    props: {
      name: "prism",
    },
  },
};
