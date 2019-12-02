<template>
    <div class="p-page p-page-labels" v-infinite-scroll="loadMore" :infinite-scroll-disabled="scrollDisabled"
         :infinite-scroll-distance="10" :infinite-scroll-listen-for-event="'scrollRefresh'">

        <v-form ref="form" class="p-labels-search" lazy-validation @submit.prevent="updateQuery" dense>
            <v-toolbar flat color="secondary">
                <v-text-field class="pt-3 pr-3"
                              single-line
                              label="Search"
                              prepend-inner-icon="search"
                              clearable
                              color="secondary-dark"
                              @click:clear="clearQuery"
                              v-model="filter.q"
                              @keyup.enter.native="updateQuery"
                              id="search"
                ></v-text-field>

                <v-spacer></v-spacer>
            </v-toolbar>
        </v-form>

        <v-container fluid class="pa-2">
            <p-scroll-top></p-scroll-top>

            <v-container grid-list-xs fluid class="pa-0 p-labels p-labels-details">
                <v-card v-if="results.length === 0" class="p-labels-empty" flat>
                    <v-card-title primary-title>
                        <div>
                            <h3 class="title mb-3">No labels matched your search</h3>
                            <div>Try again using a related or otherwise similar term.</div>
                        </div>
                    </v-card-title>
                </v-card>
                <v-layout row wrap>
                    <v-flex
                            v-for="(label, index) in results"
                            :key="index"
                            class="p-label"
                            xs6 sm4 md3 lg2 d-flex
                    >
                        <v-hover>
                            <v-card tile class="elevation-0 ma-1 accent lighten-3">
                                <v-img
                                        :src="label.getThumbnailUrl('tile_500')"
                                        aspect-ratio="1"
                                        style="cursor: pointer"
                                        class="accent lighten-2"
                                        @click.prevent="openLabel(index)"
                                >
                                    <v-layout
                                            slot="placeholder"
                                            fill-height
                                            align-center
                                            justify-center
                                            ma-0
                                    >
                                        <v-progress-circular indeterminate color="accent lighten-5"></v-progress-circular>
                                    </v-layout>
                                </v-img>

                                <v-card-actions>
                                    {{ label.LabelName | capitalize }}
                                    <v-spacer></v-spacer>
                                    <v-btn icon @click.stop.prevent="label.toggleLike()">
                                        <v-icon v-if="label.LabelFavorite" color="#FFD600">star
                                        </v-icon>
                                        <v-icon v-else color="accent lighten-2">star</v-icon>
                                    </v-btn>
                                </v-card-actions>
                            </v-card>
                        </v-hover>
                    </v-flex>
                </v-layout>
            </v-container>
        </v-container>
    </div>
</template>

<script>
    import Label from "model/label";

    export default {
        name: 'p-page-labels',
        props: {
            staticFilter: Object
        },
        watch: {
            '$route'() {
                const query = this.$route.query;

                this.filter.q = query['q'];
                this.lastFilter = {};
                this.routeName = this.$route.name;
                this.search();
            }
        },
        data() {
            const query = this.$route.query;
            const routeName = this.$route.name;
            const q = query['q'] ? query['q'] : '';
            const filter = {q: q};
            const settings = {};

            return {
                results: [],
                scrollDisabled: true,
                pageSize: 24,
                offset: 0,
                selection: this.$clipboard.selection,
                settings: settings,
                filter: filter,
                lastFilter: {},
                routeName: routeName,
            };
        },
        methods: {
            clearQuery() {
                this.filter.q = '';
                this.search();
            },
            openLabel(index) {
                const label = this.results[index];
                this.$router.push({name: "photos", query: {q: "label:" + label.LabelSlug}});
            },
            loadMore() {
                if (this.scrollDisabled) return;

                this.scrollDisabled = true;

                this.offset += this.pageSize;

                const params = {
                    count: this.pageSize,
                    offset: this.offset,
                };

                Object.assign(params, this.lastFilter);

                Label.search(params).then(response => {
                    this.results = this.results.concat(response.models);

                    this.scrollDisabled = (response.models.length < this.pageSize);

                    if (this.scrollDisabled) {
                        this.$notify.info('All ' + this.results.length + ' labels loaded');
                    }
                });
            },
            updateQuery() {
                const query = {
                    view: this.settings.view
                };

                Object.assign(query, this.filter);

                for (let key in query) {
                    if (query[key] === undefined || !query[key]) {
                        delete query[key];
                    }
                }

                this.$router.replace({query: query});
            },
            searchParams() {
                const params = {
                    count: this.pageSize,
                    offset: this.offset,
                };

                Object.assign(params, this.filter);

                if (this.staticFilter) {
                    Object.assign(params, this.staticFilter);
                }

                return params;
            },
            search() {
                this.scrollDisabled = true;

                // Don't query the same data more than once
                if (JSON.stringify(this.lastFilter) === JSON.stringify(this.filter)) {
                    this.$nextTick(() => this.$emit("scrollRefresh"));
                    return;
                }

                Object.assign(this.lastFilter, this.filter);

                this.offset = 0;

                const params = this.searchParams();

                Label.search(params).then(response => {
                    this.results = response.models;

                    this.scrollDisabled = (response.models.length < this.pageSize);

                    if (this.scrollDisabled) {
                        this.$notify.info(this.results.length + ' labels found');
                    } else {
                        this.$notify.info('More than 20 labels found');

                        this.$nextTick(() => this.$emit("scrollRefresh"));
                    }
                });
            },
        },
        created() {
            this.search();
        },
    };
</script>
