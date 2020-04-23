<template>
    <v-container grid-list-xs fluid class="pa-2 p-photos p-photo-cards">
        <v-card v-if="photos.length === 0" class="p-photos-empty secondary-light lighten-1 ma-1" flat>
            <v-card-title primary-title>
                <div>
                    <h3 class="title mb-3">
                        <translate>No photos matched your search</translate>
                    </h3>
                    <div>
                        <translate>Try using other terms and search options such as category, country and camera.
                        </translate>
                    </div>
                </div>
            </v-card-title>
        </v-card>
        <v-layout row wrap class="p-results">
            <v-flex
                    v-for="(photo, index) in photos"
                    :key="index"
                    class="p-photo"
                    xs12 sm6 md4 lg3 d-flex
                    v-bind:class="{ 'is-selected': $clipboard.has(photo) }"
            >
                <p-photo-card :photo="photo" :selection="selection" :index="index" :open-photo="openPhoto"
                              :select-range="selectRange"
                              :edit-photo="editPhoto" :open-location="openLocation" :show-location="places">
                </p-photo-card>
            </v-flex>
        </v-layout>
    </v-container>
</template>
<script>
    export default {
        name: 'p-photo-cards',
        props: {
            photos: Array,
            selection: Array,
            openPhoto: Function,
            editPhoto: Function,
            openLocation: Function,
            album: Object,
        },
        data() {
            return {
                places: this.$config.settings().features.places,
            };
        },
        methods: {
            selectRange(index) {
                this.$clipboard.addRange(index, this.photos);
            }
        }
    };
</script>
