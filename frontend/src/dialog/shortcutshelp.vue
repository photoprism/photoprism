<template>
    <v-dialog hide-overlay v-model="show" class="p-shortcuthelp-dialog">
        <v-card color="application">
            <v-toolbar dark flat color="navigation" :dense="$vuetify.breakpoint.smAndDown">
                <v-btn icon dark @click.stop="close">
                    <v-icon>close</v-icon>
                </v-btn>
                <v-toolbar-title>
                    <translate key="Shortcuts help">Keyboard Shortcuts</translate>
                </v-toolbar-title>
            </v-toolbar>

            <v-container grid-list-xs fill-height fluid overflow>
               <v-layout row wrap>
                 <v-flex v-for="(section, section_index) in $shortcuts.actions" :key="'shortcuts-'+section_index" 
                   xs12 sm10 md6 lg4 xl4 d-flex
                   class="p-shortcut-help-section"
                 >
                   <v-card tile>
                    <v-card-title>
                        <h2>{{ section.section }}</h2>
                    </v-card-title>

                    <table class="v-table v-data-table theme--light">
                        <thead>
                            <tr>
                                <th class="text-xs-left">Keyboard Shortcut</th>
                                <th class="text-xs-left">Action</th>
                            </tr>
                        </thead>
                        <tbody>
                            <tr v-for="(action, action_index) in section.actions" :key="'action-'+action_index">
                                <td class="text-xs-left">{{ action.name }}</td>
                                <td class="text-xs-left">
                                    <span v-for="(displayKey, key_index) in action.displayKeys" :key="'key-'+key_index">
                                        <span v-if="key_index > 0">or</span>
                                            <span v-for="(keystroke, keystroke_index) in displayKey" :key="'keystroke-'+keystroke_index">
                                                <span v-if="keystroke_index > 0">then</span>
                                                <span v-for="modifier in keystroke.modifiers" :key="modifier"><span class="keyboard-shortcuts-key">{{ modifier }}</span>+</span>
                                                <span class="keyboard-shortcuts-key">{{ keystroke.key }}</span>
                                            </span>
                                    </span>
                                </td>
                            </tr>
                        </tbody>
                    </table>
                  </v-card>
                 </v-flex>
                </v-layout>
            </v-container>
       </v-card>
    </v-dialog>
</template>
<script>
export default {
    name: 'p-shortcuts-help',
    props: {
        show: Boolean,
    },
    data() {
        return {
            dialog: false
        }
    },
    mounted() {
        this.$shortcuts.activate(this, 'dialog')
    },
    unmounted() {
        this.$shortcuts.deactivate('dialog')
    },
    methods: {
        close() {
            this.$emit('close');
        }
    }
}
</script>