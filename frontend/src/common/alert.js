import Vue from 'vue';
import Event from 'pubsub-js';

const Alert = {
    info: function (message) {
        Event.publish('alert.info', message);
    },
    warning: function (message) {
        Event.publish('alert.warning', message);
    },
    error: function (message) {
        Event.publish('alert.error', message);
    },
    success: function (message) {
        Event.publish('alert.success', message);
    },
};

new Vue({
    el: '#alerts',
    template: '<div id="alerts"><div v-for="message in messages" :class="message.type" class="alert">{{ message.msg }}</div></div>',
    data() {
        return {
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
            this.addMessage('warning', message, 4000);
        },

        addErrorMessage: function (message) {
            this.addMessage('error', message, 8000);
        },

        addSuccessMessage: function (message) {
            this.addMessage('success', message, 3000);
        },

        addInfoMessage: function (message) {
            this.addMessage('info', message, 3000);
        },

        addMessage: function (type, message, delay) {
            if (message === this.lastMessage) return;

            this.lastMessageId++;
            this.lastMessage = message;

            const alert = {'id': this.lastMessageId, 'type': type, 'msg': message};

            this.messages.push(alert);

            setTimeout(() => {
                this.messages.shift();
                this.lastMessage = '';
            }, delay);
        },

        closeAlert: function (index) {
            this.messages.splice(index, 1);
        },
    },
});

export default Alert;