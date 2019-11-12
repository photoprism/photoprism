<template>
    <div class="p-tab p-tab-general">
        <v-container fluid>
            <v-form ref="form" class="p-form-settings" lazy-validation @submit.prevent="save" dense>
                <v-layout wrap align-center>
                    <v-flex xs12 sm6 class="pr-3">
                        <v-select
                                :items="languages"
                                label="Language"
                                color="blue-grey"
                                value="en"
                                flat
                        ></v-select>
                    </v-flex>

                    <v-flex xs12 sm6 class="pr-3">
                        <v-select
                                :items="themes"
                                label="Theme"
                                color="blue-grey"
                                value=""
                                flat
                        ></v-select>
                    </v-flex>
                </v-layout>

                <v-btn color="blue-grey"
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

    export default {
        name: 'p-tab-general',
        data() {
            return {
                readonly: this.$config.getValue("readonly"),
                active: 0,
                settings: new Settings(),
                list: {},
                themes: [{text: "Default", value: ""}],
                languages: [{text: "English", value: "en"}],
            };
        },
        methods: {
            load() {
                this.settings.load().then((r) => { this.list = r.getValues(); });
            },
            save() {
                this.settings.save().then(() => {
                    this.$alert.info("Settings saved");
                })
            },
        },
        created() {
            this.load();
        },
    };
</script>
