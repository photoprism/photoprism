<template>
    <div class="page page-photos">
        <div class="photo-grid">
            <template v-for="photo in rows">
                <div class="photo">
                    <div class="info">{{ photo.CreatedAt | moment("DD.MM.YYYY hh:mm:ss") }}<span class="right">{{ photo.CameraModel }}</span></div>
                    <div class="actions">
                        <a class="action like" v-bind:class="{ liked: photo.Liked, notliked: !photo.Liked }" v-on:click="likePhoto(photo)">
                            <i v-if="!photo.Liked" class="far fa-heart"></i>
                            <i v-if="photo.Liked" class="fas fa-heart"></i>
                        </a>
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
        props: {},
        data() {
            const query = this.$route.query;
            const resultCount = query.hasOwnProperty('count') ? parseInt(query['count']) : 15;
            const resultPage = query.hasOwnProperty('page') ? parseInt(query['page']) : 1;
            const resultOffset = resultCount * (resultPage - 1);
            const order = query.hasOwnProperty('order') ? query['order'] : '';
            const dir = query.hasOwnProperty('dir') ? query['dir'] : '';
            const q = query.hasOwnProperty('q') ? query['q'] : '';

            return {
                'rows': [],
                'images': [],
                'page': resultPage,
                'order': order,
                'dir': dir,
                'q': q,
                'pageOptions': [15, 30, 50, 100],
                'resultCount': resultCount,
                'resultOffset': resultOffset,
                'resultTotal': 'Many',
                'lastQuery': {},
                'submitTimeout': false,
            };
        },
        methods: {
            likePhoto(photo) {
                photo.Liked = !photo.Liked;
            },

            deletePhoto(photo) {
                this.$alert.success('Photo deleted');
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
                    q: this.q
                };

                Object.assign(urlParams, this.query);

                this.$router.replace({query: urlParams});

                Photo.search(params).then(response => {
                    console.log(response);
                    this.resultTotal = parseInt(response.headers['x-result-total']);
                    this.resultCount = parseInt(response.headers['x-result-count']);
                    this.resultOffset = parseInt(response.headers['x-result-offset']);
                    this.rows = response.models;
                    this.$alert.info(this.rows.length + ' photos loaded');
                });
            }
        },
        created() {
            this.refreshList();
        },
    };
</script>

<style scoped>
    h1, h2 {
        font-weight: normal;
    }

    ul {
        list-style-type: none;
        padding: 0;
    }

    li {
        display: inline-block;
        margin: 0 10px;
    }
</style>
