<template>
    <div class="p-tab p-tab-photo-share-todo">
        <v-container fluid>
            <p class="subheading pb-3">
                This is work in progress. Feedback and contributions welcome.
            </p>
        </v-container>
    </div>
</template>

<script>
    import options from "resources/options.json";

    export default {
        name: 'p-tab-photo-edit-todo',
        props: {
            model: Object,
        },
        data() {
            return {
                config: this.$config.values,
                all: {
                    countries: [{code: "", name: this.$gettext("Unknown")}],
                    cameras: [{ID: 0, CameraModel: this.$gettext("Unknown")}],
                    lenses: [{ID: 0, LensModel: "Unknown"}],
                    colors: [{label: "Unknown", name: ""}],
                },
                readonly: this.$config.getValue("readonly"),
                options: options,
                labels: {
                    search: this.$gettext("Search"),
                    view: this.$gettext("View"),
                    country: this.$gettext("Country"),
                    camera: this.$gettext("Camera"),
                    lens: this.$gettext("Lens"),
                    year: this.$gettext("Year"),
                    color: this.$gettext("Color"),
                    category: this.$gettext("Category"),
                    sort: this.$gettext("Sort By"),
                    before: this.$gettext("Taken before"),
                    after: this.$gettext("Taken after"),
                    language: this.$gettext("Language"),
                    theme: this.$gettext("Theme"),
                },
                showDatepicker: false,
                showTimepicker: false,
                date: "",
                time: "",
            };
        },
        computed: {
        },
        methods: {
            openPhoto() {
                this.$viewer.show([this.model], 0)
            },
            refresh() {
            },
        },
    };
</script>
