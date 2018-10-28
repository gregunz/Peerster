new Vue({
    el: '#app',

    data: {
        apiURL: '',
        ws: null, // Our websocket
        chatBox: '', // Holds new messages to be sent to the server
        chatMessages: [], // chat messages list
        name: '',
    },

    created: function() {
        const self = this;
        this.apiURL = window.location.host;
        axios
            .get('http://' + this.apiURL + '/id')
            .then(response =>  {
                this.name = response.data;
            });
        this.ws = new WebSocket('ws://' + this.apiURL + '/ws');
        this.ws.onopen = () => {
            self.ws.send(JSON.stringify({'subscribe-message' : {'with-previous': true}})); // subscribing to messages
        };

        this.ws.addEventListener('message', function (e) { // here we receive packets from the websocket

            const packet = JSON.parse(e.data);
            if (packet.text && packet.origin) { // message packet
                self.saveMsg(packet.text, packet.origin);
            } else {
                console.log("packet not handled: " + packet);
            }

        });
    },

    methods: {
        send: function () {
            if (this.chatBox !== '') {
                const msg = this.stripOutHtml(this.chatBox);
                const msgPacket = {
                    'post-message': {
                        'message': msg
                    }
                };
                this.ws.send(JSON.stringify(msgPacket));
                this.chatBox = ''; // Reset chatBox
            }
        },

        saveMsg: function (text, name) {
            text = this.stripOutHtml(text);
            this.chatMessages.push({
                text: text,
                origin: name,
            });
            const self = this;
            setTimeout(function () {
                self.scrollToTop('chat-messages')
            }, 1);
        },

        scrollToTop: function (id) {
            const element = document.getElementById(id);
            element.scrollTop = element.scrollHeight; // Auto scroll to the bottom
        },

        stripOutHtml: function (text) {
            return $('<p>').html(text).text();
        },

        avatarURL: function(name) {
            return 'https://api.adorable.io/avatars/100/' + name;
        }
    }
});
