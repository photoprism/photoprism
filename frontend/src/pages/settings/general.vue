<template>
    <div class="p-tab p-tab-general">
        <v-container fluid>
            <v-form lazy-validation dense
                    ref="form" class="p-form-settings" accept-charset="UTF-8"
                    @submit.prevent="save">
                <v-layout wrap align-center>
                    <v-flex xs12 sm6 class="pr-3">
                        <v-select
                                :items="options.languages"
                                :label="labels.language"
                                color="secondary-dark"
                                v-model="settings.language"
                                flat
                        ></v-select>
                    </v-flex>

                    <v-flex xs12 sm6 class="pr-3">
                        <v-select
                                :items="options.themes"
                                :label="labels.theme"
                                color="secondary-dark"
                                v-model="settings.theme"
                                flat
                        ></v-select>
                    </v-flex>
                </v-layout>

                <v-btn color="secondary-dark"
                       class="white--text ml-0 mt-2"
                       depressed
                       @click.stop="save">
                    <translate>Save</translate>
                    <v-icon right dark>save</v-icon>
                </v-btn>
            </v-form>
        </v-container>
    </div>
</template>

<script>
    import Settings from "model/settings";
    import options from "resources/options.json";

    export default {
        name: 'p-tab-general',
        data() {
            return {
                readonly: this.$config.getValue("readonly"),
                settings: new Settings(this.$config.values.settings),
                options: options,
                labels: {
                    language: this.$gettext("Language"),
                    theme: this.$gettext("Theme"),
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

                    if(reload) {
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
