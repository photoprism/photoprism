<template>
  <div class="p-tab p-tab-discover-season">
    <div class="pa-2 text-center">
      <v-row class="p-season" justify="center">
        <v-col cols="12" sm="8" md="6" lg="4" class="pa-2">
          <v-card :to="{ name: 'browse', query: memoriesQuery }" class="clickable py-3" color="surface" variant="elevated">
            <v-card-text>
              <v-icon size="48" color="primary" class="mb-3">mdi-calendar-heart</v-icon>
              <div class="text-h6">{{ $gettext(`On This Day`) }}</div>
              <div class="text-body-2 mt-2">
                {{ $gettextInterpolate($gettext(`Memories from %{date}`), { date: todayLabel }) }}
              </div>
            </v-card-text>
          </v-card>
        </v-col>
      </v-row>
    </div>
  </div>
</template>

<script>
export default {
  name: "PTabDiscoverSeason",
  data() {
    const today = new Date();

    return {
      readonly: this.$config.get("readonly"),
      today,
    };
  },
  computed: {
    memoriesQuery() {
      return {
        month: this.today.getMonth() + 1,
        day: this.today.getDate(),
        order: "oldest",
      };
    },
    todayLabel() {
      return this.today.toLocaleDateString(undefined, {
        month: "long",
        day: "numeric",
      });
    },
  },
};
</script>
