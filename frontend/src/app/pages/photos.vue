<template>
    <div class="page page-photos">
        <div class="page-form">
            <b-form inline @submit="formChange">
                <b-form-select class="mb-2 mr-sm-2"
                               v-b-tooltip.hover title="Category"
                               v-model="query.category"
                               :options="{ 'junction': 'Junction', 'tourism': 'Tourism', 'historic': 'Historic' }"
                               id="inlineFormCustomSelectPref">
                    <option slot="first" :value="null"></option>
                </b-form-select>

                <b-form-select @change="formChange" class="mb-2 mr-sm-2"
                               v-model="query.country"
                               :options="{ '1': 'One', '2': 'Two', '3': 'Three' }"
                               id="inlineFormCustomSelectPref">
                    <option slot="first" :value="null">Country</option>
                </b-form-select>
                <b-form-select @change="formChange" class="mb-2 mr-sm-2"
                               :v-model="query.camera"
                               :options="{ '1': 'One', '2': 'Two', '3': 'Three' }"
                               id="inlineFormCustomSelectPref">
                    <option slot="first" :value="null">Camera Model</option>
                </b-form-select>
                <b-form-select @change="formChange" class="mb-2 mr-sm-2"
                               v-model="dir"
                               :options="{ 'asc': 'Ascending', 'desc': 'Descending' }"
                               id="inlineFormCustomSelectPref">
                    <option slot="first" :value="null">Sort Order</option>
                </b-form-select>

                <b-form-select @change="formChange" class="mb-2 mr-sm-2"
                               v-model="view"
                               :options="{ 'list': 'List View', 'tile': 'Tile View (small)', 'tile_large': 'Tile View (large)' }"
                               id="inlineFormCustomSelectPref">
                </b-form-select>

                <b-form-input class="mb-2 mr-sm-2" v-b-tooltip.hover title="After" type="date"/>
                <b-form-input class="mb-2 mr-sm-2" v-b-tooltip.hover title="Before" type="date"/>

                <b-form-checkbox class="mb-2 mr-sm-2 mb-sm-0">
                    Favorites only
                </b-form-checkbox>
            </b-form>
            <div class="clearfix"></div>
        </div>
        <div class="page-container photo-grid">
            <template v-for="photo in items">
                <div class="photo">
                    <div class="info">{{ photo.TakenAt | moment("DD.MM.YYYY hh:mm:ss") }}<span class="right">{{ photo.CameraModel }}</span></div>
                    <div class="actions">
                        <span class="left">
                            <a class="action like" v-bind:class="{ favorite: photo.Favorite }" v-on:click="likePhoto(photo)">
                                <i v-if="!photo.Favorite" class="far fa-heart"></i>
                                <i v-if="photo.Favorite" class="fas fa-heart"></i>
                            </a>
                        </span>
                        <span class="center" v-if="photo.Location">
                            <a class="location" target="_blank" :href="photo.getGoogleMapsLink()" v-b-tooltip.hover :title="photo.Location.DisplayName">{{ photo.Location.Country }}</a>
                        </span>
                        <span class="right">
                            <a class="action delete" v-on:click="deletePhoto(photo)">
                                <i class="fas fa-trash"></i>
                            </a>
                        </span>
                    </div>
                <template v-for="file in photo.Files">
                    <img v-if="file.FileType === 'jpg'" :src="'/api/v1/files/' + file.ID + '/square_thumbnail?size=250'">
                </template>
                </div>
            </template>
        </div>
    </div>
</template>

<script>
    import Photo from 'model/photo';
    import _ from 'lodash/lang';

    export default {
        name: 'photos',
        props: {
        },
        data() {
            const query = this.$route.query;
            const resultCount = query.hasOwnProperty('count') ? parseInt(query['count']) : 70;
            const resultPage = query.hasOwnProperty('page') ? parseInt(query['page']) : 1;
            const resultOffset = resultCount * (resultPage - 1);
            const order = query.hasOwnProperty('order') ? query['order'] : 'taken_at';
            const dir = query.hasOwnProperty('dir') ? query['dir'] : '';
            const q = query.hasOwnProperty('q') ? query['q'] : '';
            const view = query.hasOwnProperty('view') ? query['view'] : 'tile';

            return {
                'items': [],
                'query': {
                    category: '',
                    country: '',
                    camera: '',
                    after: '',
                    before: '',
                    favorites_only: '',
                    q: q,
                },
                'page': resultPage,
                'order': order,
                'dir': dir,
                'view': view,
                'resultCount': resultCount,
                'resultOffset': resultOffset,
                'resultTotal': 'Many',
                'lastQuery': {},
                'submitTimeout': false,
            };
        },
        methods: {
            likePhoto(photo) {
                photo.Favorite = !photo.Favorite;
            },
            deletePhoto(photo) {
                this.$alert.success('Photo deleted');
            },
            formChange(event) {
                this.$alert.success('Form change');
                this.refreshList();
            },
            refreshList() {
                // Compose query parameters
                const params = {
                    count: this.resultCount,
                    offset: this.resultCount * (this.page - 1),
                    order: this.order !== '' ? this.order + ' ' + this.dir : '',
                };

                Object.assign(params, this.query);

                // Don't query the same data more than once
                if (_.isEqual(this.lastQuery, params)) return;

                this.lastQuery = params;

                // Set URL hash
                const urlParams = {
                    count: this.resultCount,
                    page: this.page,
                    order: this.order,
                    dir: this.dir,
                };

                Object.assign(urlParams, this.query);

                this.$router.replace({query: urlParams});

                Photo.search(params).then(response => {
                    console.log(response);
                    this.resultTotal = parseInt(response.headers['x-result-total']);
                    this.resultCount = parseInt(response.headers['x-result-count']);
                    this.resultOffset = parseInt(response.headers['x-result-offset']);
                    this.items = response.models;
                    this.$alert.info(this.items.length + ' photos found');
                });
            }
        },
        created() {
            this.refreshList();
        },
    };
</script>

<style scoped>
</style>
