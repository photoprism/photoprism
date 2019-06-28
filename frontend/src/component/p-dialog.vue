<template>
    <v-layout row justify-center>
        <!--<v-dialog v-model="dialog" color="blue-grey lighten-5" persistent max-width="350" transition="dialog-right-transition">
            <v-card raised>
                <v-toolbar flat color="blue-grey lighten-4">
                    <v-icon>delete</v-icon>
                    <v-toolbar-title>Delete Photos</v-toolbar-title>
                    <v-spacer></v-spacer>
                </v-toolbar>
                <v-card-text>You want to delete the {{ count }} selected items? You can't undo this action.</v-card-text>
                <v-card-actions>
                    <v-layout row align-center justify-center fill-height>
                    <v-btn @click.stop="close" depressed color="teal accent-4">Keep</v-btn>
                    <v-btn color="delete" @click.stop="agree">Delete<v-icon right dark>delete</v-icon></v-btn>
                    <v-btn @click.stop="close" depressed color="teal accent-4"><v-icon left dark>clear</v-icon>Cancel</v-btn>
                    <v-btn color="delete" @click.stop="agree"><v-icon left dark>done</v-icon> Confirm</v-btn>
                    </v-layout>
                </v-card-actions>
            </v-card>-->
       <v-dialog v-model="dialog" persistent max-width="350" transition="dialog-right-transition" class="p-photo-dialog">
            <v-card raised elevation="24">
                <v-layout row align-center justify-center>
                    <v-icon size="60" color="grey lighten-1" class="ma-2 pt-3">error_outline</v-icon>
                </v-layout>
                <v-layout column align-center justify-center fill-height>
                    <div class="headline my-1 mx-3">Are you sure?</div>
                <span class="text-align-center p-photo-dialog-text my-3 mx-3">That you want to delete the {{ count }} selected items? You can't undo this action.</span>
                    <v-spacer></v-spacer>
                </v-layout>
                <v-card-actions>
                    <v-layout row align-center justify-center fill-height class="pa-2">
                        <v-btn @click.stop="close" depressed color="grey darken-1" class="p-photo-dialog-cancel white--text">Cancel</v-btn>
                        <v-btn color="delete" @click.stop="agree" class="p-photo-dialog-confirm white--text">Delete</v-btn>
                    </v-layout>
                </v-card-actions>
            </v-card>
    <!--<v-dialog v-model="dialog" persistent max-width="450" transition="dialog-right-transition">
        <v-card raised elevation="24">
            <v-layout row align-center justify-center fill-height>
                <v-icon size="80" color="grey lighten-1" class="ma-3">error_outline</v-icon>
            <v-layout column fill-height class="pa-3">
                <div class="headline">Are you sure?</div>
                <span>You want to delete the {{ count }} selected items? You can't undo this action.</span>
                <v-spacer></v-spacer>
            </v-layout>
            </v-layout>
            <v-card-actions>
                <v-layout row align-center justify-center class="pb-2 ma-2">
                <v-btn @click.stop="close" depressed color="grey lighten-1">Cancel</v-btn>
                <v-btn color="delete" @click.stop="agree">Delete</v-btn>
            </v-layout>
            </v-card-actions>
        </v-card>-->
        </v-dialog>
    </v-layout>
</template>
<script>
    export default {
        name: 'p-dialog',
        props: {
            dialog: Boolean,
            count: String,
        },
        data () {
            return {
            }
        },
        methods: {
            close() {
                this.$emit('close');
            },
            agree() {
                this.$emit('agree');
            },
        }
    }
</script>