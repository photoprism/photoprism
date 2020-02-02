<template>
    <div class="p-page p-page-labels" v-infinite-scroll="loadMore" :infinite-scroll-disabled="scrollDisabled"
         :infinite-scroll-distance="10" :infinite-scroll-listen-for-event="'scrollRefresh'">

        <v-form ref="form" class="p-labels-search" lazy-validation @submit.prevent="updateQuery" dense>
            <v-toolbar flat color="secondary">
                <v-text-field class="pt-3 pr-3"
                              single-line
                              :label="labels.search"
                              prepend-inner-icon="search"
                              browser-autocomplete="off"
                              clearable
                              color="secondary-dark"
                              @click:clear="clearQuery"
                              v-model="filter.q"
                              @keyup.enter.native="updateQuery"
                              id="search"
                ></v-text-field>

                <v-spacer></v-spacer>

                <v-btn icon @click.stop="refresh">
                    <v-icon>refresh</v-icon>
                </v-btn>

                <v-btn v-if="!filter.all" icon @click.stop="showAll">
                    <v-icon>visibility</v-icon>
                </v-btn>
                <v-btn v-else icon @click.stop="showImportant">
                    <v-icon>visibility_off</v-icon>
                </v-btn>
            </v-toolbar>
        </v-form>

        <v-container fluid class="pa-4" v-if="loading">
            <v-progress-linear color="secondary-dark" :indeterminate="true"></v-progress-linear>
        </v-container>
        <v-container fluid class="pa-0" v-else>
            <p-scroll-top></p-scroll-top>

            <v-container grid-list-xs fluid class="pa-2 p-labels p-labels-details">
                <v-card v-if="results.length === 0" class="p-labels-empty secondary-light lighten-1" flat>
                    <v-card-title primary-title>
                        <div>
                            <h3 class="title mb-3">
                                <translate>No labels matched your search</translate>
                            </h3>
                            <div>
                                <translate>Try again using a related or otherwise similar term.</translate>
                            </div>
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
                            <v-card tile class="elevation-0 ma-1 accent lighten-3"
                                    :to="{name: 'photos', query: {q: 'label:' + label.LabelSlug}}">
                                <v-img
                                        :src="label.getThumbnailUrl('tile_500')"
                                        aspect-ratio="1"
                                        class="accent lighten-2"
                                >
                                    <v-layout
                                            slot="placeholder"
                                            fill-height
                                            align-center
                                            justify-center
                                            ma-0
                                    >
                                        <v-progress-circular indeterminate
                                                             color="accent lighten-5"></v-progress-circular>
                                    </v-layout>
                                </v-img>

                                <v-card-actions @click.stop.prevent="">
                                    <v-edit-dialog
                                            :return-value.sync="label.LabelName"
                                            lazy
                                            @save="onSave(label)"
                                            class="p-inline-edit"
                                    >
                                        <span v-if="label.LabelName">
                                            {{ label.LabelName | capitalize }}
                                        </span>
                                        <span v-else>
                                            <v-icon>edit</v-icon>
                                        </span>
                                        <template v-slot:input>
                                            <v-text-field
                                                    v-model="label.LabelName"
                                                    :rules="[titleRule]"
                                                    :label="labels.name"
                                                    color="secondary-dark"
                                                    single-line
                                                    autofocus
                                            ></v-text-field>
                                        </template>
                                    </v-edit-dialog>
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

                this.filter.q = query['q'] ? query['q'] : '';
                this.filter.all = query['all'] ? query['all'] : '';
                this.lastFilter = {};
                this.routeName = this.$route.name;
                this.search();
            }
        },
        data() {
            const query = this.$route.query;
            const routeName = this.$route.name;
            const q = query['q'] ? query['q'] : '';
            const all = query['all'] ? query['all'] : '';
            const filter = {q: q, all: all};
            const settings = {};

            return {
                results: [],
                scrollDisabled: true,
                loading: true,
                pageSize: 24,
                offset: 0,
                selection: this.$clipboard.selection,
                settings: settings,
                filter: filter,
                lastFilter: {},
                routeName: routeName,
                labels: {
                    search: this.$gettext("Search"),
                    name: this.$gettext("Label Name"),
                },
                titleRule: v => v.length <= 25 || this.$gettext("Name too long"),
            };
        },
        methods: {
            onSave(label) {
                label.update();
            },
            showAll() {
                this.filter.all = "true";
                this.updateQuery();
            },
            showImportant() {
                this.filter.all = "";
                this.updateQuery();
            },
            clearQuery() {
                this.filter.q = '';
                this.updateQuery();
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
                        this.$notify.info(this.$gettext('All ') + this.results.length + this.$gettext(' labels loaded'));
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

                if (JSON.stringify(this.$route.query) === JSON.stringify(query)) {
                    return
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
            refresh() {
                this.lastFilter = {};
                const pageSize = this.pageSize;
                this.pageSize = this.offset + pageSize;
                this.search();
                this.offset = this.pageSize;
                this.pageSize = pageSize;
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
                this.loading = true;

                const params = this.searchParams();

                Label.search(params).then(response => {
                    this.loading = false;

                    this.results = response.models;

                    this.scrollDisabled = (response.models.length < this.pageSize);

                    if (this.scrollDisabled) {
                        this.$notify.info(this.results.length + this.$gettext(' labels found'));
                    } else {
                        this.$notify.info(this.$gettext('More than 20 labels found'));

                        this.$nextTick(() => this.$emit("scrollRefresh"));
                    }
                }).catch(() => this.loading = false);
            },
        },
        created() {
            this.search();
        },
    };
</script>
