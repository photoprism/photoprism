<template>
  <div class="p-page p-page-discover">
    <v-tabs v-model="active" flat grow color="secondary" slider-color="secondary-dark" :height="$vuetify.breakpoint.smAndDown ? 48 : 64">
      <v-tab id="tab-discover-colors" ripple @click="changePath('/discover')">
        <translate>Colors</translate>
      </v-tab>

      <v-tab id="tab-discover-similar" ripple @click="changePath('/discover/similar')">
        <translate>Similar</translate>
      </v-tab>

      <v-tab id="tab-discover-season" ripple @click="changePath('/discover/season')">
        <translate>Season</translate>
      </v-tab>

      <v-tab id="tab-discover-random" ripple @click="changePath('/discover/random')">
        <translate>Random</translate>
      </v-tab>

      <v-tabs-items touchless>
        <v-tab-item>
          <p-tab-discover-colors></p-tab-discover-colors>
        </v-tab-item>

        <v-tab-item>
          <p-tab-discover-todo></p-tab-discover-todo>
        </v-tab-item>

        <v-tab-item>
          <p-tab-discover-todo></p-tab-discover-todo>
        </v-tab-item>

        <v-tab-item>
          <p-tab-discover-todo></p-tab-discover-todo>
        </v-tab-item>
      </v-tabs-items>
    </v-tabs>
  </div>
</template>

<script>
import tabColors from "page/discover/colors.vue";
import tabTodo from "page/discover/todo.vue";

export default {
  name: "PPageDiscover",
  components: {
    "p-tab-discover-colors": tabColors,
    "p-tab-discover-todo": tabTodo,
  },
  props: {
    tab: Number,
  },
  data() {
    return {
      readonly: this.$config.get("readonly"),
      active: this.tab,
    };
  },
  methods: {
    changePath: function (path) {
      if (this.$route.path !== path) {
        this.$router.replace(path);
      }
    },
  },
};
</script>
