<template>
    <v-dialog v-model="show" persistent max-width="350" class="p-photo-delete-dialog" @keydown.esc="cancel">
        <v-card raised elevation="24">
            <v-container fluid class="pb-2 pr-2 pl-2">
                <v-layout row wrap>
                    <v-flex xs3 text-xs-center>
                        <v-icon size="54" color="grey lighten-1">delete_outline</v-icon>
                    </v-flex>
                    <v-flex xs9 text-xs-left align-self-center>
                        <div class="subheading pr-1">Are you sure you want to delete these photos?</div>
                    </v-flex>
                    <v-flex xs12 text-xs-right class="pt-3">
                        <v-btn @click.stop="cancel" depressed color="grey lighten-3" class="p-photo-dialog-cancel">
                            Cancel
                        </v-btn>
                        <v-btn color="blue-grey lighten-2" depressed dark @click.stop="confirm"
                               class="p-photo-dialog-confirm">Delete
                        </v-btn>
                    </v-flex>
                </v-layout>
            </v-container>
        </v-card>
    </v-dialog>
</template>
<script>
    export default {
        name: 'p-photo-delete-dialog',
        props: {
            show: Boolean,
        },
        data() {
            return {}
        },
        methods: {
            cancel() {
                this.$emit('cancel');
            },
            confirm() {
                this.$emit('confirm');
            },
        }
    }
</script>
