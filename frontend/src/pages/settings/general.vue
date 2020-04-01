<template>
    <div class="p-tab p-settings-general">
        <v-form lazy-validation dense
                ref="form" class="p-form-settings" accept-charset="UTF-8"
                @submit.prevent="save">
            <v-card flat tile class="px-1 application">
                <v-card-title primary-title class="pb-2">
                    <h3 class="body-2 mb-0">User Interface</h3>
                </v-card-title>

                <v-card-actions>
                    <v-layout wrap align-center>
                        <v-flex xs12 sm6 class="px-2 pb-2">
                            <v-select
                                    :items="options.themes"
                                    :label="labels.theme"
                                    color="secondary-dark"
                                    background-color="secondary-light"
                                    v-model="settings.theme"
                                    hide-details box
                            ></v-select>
                        </v-flex>

                        <v-flex xs12 sm6 class="px-2 pb-2">
                            <v-select
                                    :items="options.languages"
                                    :label="labels.language"
                                    color="secondary-dark"
                                    background-color="secondary-light"
                                    v-model="settings.language"
                                    hide-details box
                            ></v-select>
                        </v-flex>
                    </v-layout>
                </v-card-actions>
            </v-card>

            <v-card flat tile class="mt-0 px-1 application">
                <v-card-title primary-title class="pb-2">
                    <h3 class="body-2 mb-0">Places</h3>
                </v-card-title>

                <v-card-actions>
                    <v-layout wrap align-center>
                        <v-flex xs12 sm6 class="px-2 pb-2">
                            <v-select
                                    :items="options.mapsStyle"
                                    :label="labels.mapsStyle"
                                    color="secondary-dark"
                                    background-color="secondary-light"
                                    v-model="settings.maps.style"
                                    hide-details box
                            ></v-select>
                        </v-flex>

                        <v-flex xs12 sm6 class="px-2 pb-2">
                            <v-select
                                    :items="options.mapsAnimate"
                                    :label="labels.mapsAnimate"
                                    color="secondary-dark"
                                    background-color="secondary-light"
                                    v-model="settings.maps.animate"
                                    hide-details box
                            ></v-select>
                        </v-flex>
                    </v-layout>
                </v-card-actions>
            </v-card>

            <v-container fluid class="mt-1">
                <v-btn color="secondary-dark"
                       class="ml-1"
                       depressed dark
                       @click.stop="save">
                    <translate>Save</translate>
                    <v-icon right dark>save</v-icon>
                </v-btn>
            </v-container>
        </v-form>
    </div>
</template>

<script>
    import Settings from "model/settings";
    import options from "resources/options.json";

    export default {
        name: 'p-settings-general',
        data() {
            return {
                readonly: this.$config.getValue("readonly"),
                settings: new Settings(this.$config.values.settings),
                options: options,
                labels: {
                    language: this.$gettext("Language"),
                    theme: this.$gettext("Theme"),
                    mapsAnimate: this.$gettext("Animation"),
                    mapsStyle: this.$gettext("Style"),
                },
            };
        },
        methods: {
            load() {
                this.settings.load();
            },
            save() {
                const reload = this.settings.changed("language");

                this.settings.save().then((s) => {
                    this.$config.updateSettings(s.getValues(), this.$vuetify);

                    if (reload) {
                        this.$notify.info(this.$gettext("Reloading..."));
                        this.$notify.blockUI();
                        setTimeout(() => window.location.reload(), 100);
                    } else {
                        this.$notify.success(this.$gettext("Settings saved"));
                    }
                })
            },
        },
        created() {
            this.load();
        },
    };
</script>
