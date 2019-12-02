<template>
    <div class="p-tab p-tab-general">
        <v-container fluid>
            <v-form ref="form" class="p-form-settings" lazy-validation @submit.prevent="save" dense>
                <v-layout wrap align-center>
                    <v-flex xs12 sm6 class="pr-3">
                        <v-select
                                :items="options.languages"
                                label="Language"
                                color="secondary-dark"
                                v-model="settings.language"
                                flat
                        ></v-select>
                    </v-flex>

                    <v-flex xs12 sm6 class="pr-3">
                        <v-select
                                :items="options.themes"
                                label="Theme"
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
                    Save
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
            };
        },
        methods: {
            load() {
                this.settings.load();
            },
            save() {
                this.settings.save().then((s) => {
                    this.$config.updateSettings(s.getValues(), this.$vuetify);
                    this.$notify.info("Settings saved");
                })
            },
        },
        created() {
            this.load();
        },
    };
</script>
