<template>
    <v-snackbar
            id="p-alert"
            v-model="visible"
            :color="color"
            :timeout="0"
            :class="textColor"
            top
            right
    >
        {{ text }}
        <v-btn
                :class="textColor + ' pr-0'"
                icon
                flat
                @click="close"
        >
            <v-icon>close</v-icon>
        </v-btn>
    </v-snackbar>
</template>
<script>
    import Event from 'pubsub-js';

    export default {
        name: 'p-alert',
        data() {
            return {
                text: '',
                color: 'primary',
                textColor: '',
                visible: false,
                messages: [],
                lastMessageId: 1,
                lastMessage: '',
                subscriptionId: '',
            };
        },
        created() {
            this.subscriptionId = Event.subscribe('alert', this.handleAlertEvent);
        },
        destroyed() {
            Event.unsubscribe(this.subscriptionId);
        },
        methods: {
            handleAlertEvent: function (ev, message) {
                const type = ev.split('.')[1];

                switch (type) {
                    case 'warning':
                        this.addWarningMessage(message);
                        break;
                    case 'error':
                        this.addErrorMessage(message);
                        break;
                    case 'success':
                        this.addSuccessMessage(message);
                        break;
                    case 'info':
                        this.addInfoMessage(message);
                        break;
                    default:
                        alert(message);
                }
            },

            addWarningMessage: function (message) {
                this.addMessage('warning', 'black--text', message, 4000);
            },

            addErrorMessage: function (message) {
                this.addMessage('error', 'white--text', message, 8000);
            },

            addSuccessMessage: function (message) {
                this.addMessage('success', 'white--text', message, 3000);
            },

            addInfoMessage: function (message) {
                this.addMessage('info', 'white--text', message, 3000);
            },

            addMessage: function (color, textColor, message, delay) {
                if (message === this.lastMessage) return;

                this.lastMessageId++;
                this.lastMessage = message;

                const alert = {'id': this.lastMessageId, 'color': color, 'textColor': textColor, 'delay': delay, 'msg': message};

                this.messages.push(alert);

                if(!this.visible) {
                    this.show();
                }
            },

            close: function () {
                this.visible = false;
                this.show();
            },
            show: function () {
                const message = this.messages.shift();

                if(message) {
                    this.text = message.msg;
                    this.color = message.color;
                    this.textColor = message.textColor;
                    this.visible = true;

                    setTimeout(() => {
                        this.lastMessage = '';
                        this.show();
                    },  message.delay);
                } else {
                    this.visible = false;
                    this.text = '';
                }
            },
        },
    };
</script>
