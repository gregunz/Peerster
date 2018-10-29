new Vue({
    el: '#app',

    data: {
        apiURL: '',
        ws: null, // Our websocket
        chatBox: '', // Holds new messages to be sent to the server
        chatMessages: [], // chat messages list
        nodeBox: '',
        nodes: [],
        name: '',
    },

    created: function() {
        const self = this;
        this.apiURL = window.location.host;
        this.ws = new WebSocket('ws://' + this.apiURL + '/ws');
        this.ws.onopen = () => {
            self.ws.send(JSON.stringify({'get-id' : {}})); // getting name
            self.ws.send(JSON.stringify({'subscribe-message' : {'subscribe': true, 'with-previous': true}})); // subscribing to messages
            self.ws.send(JSON.stringify({'subscribe-node' : {'subscribe': true, 'with-previous': true}})); // subscribing to nodes
        };

        this.ws.addEventListener('message', function (e) { // here we receive packets from the websocket

            const packet = JSON.parse(e.data);

            if (packet['get-id']) {
                self.name = packet['get-id'].id;
                return;
            }

            if (packet.rumor) { // rumor message packet
                const rumorMsg = packet.rumor.message;
                self.saveMsg(rumorMsg.text, rumorMsg.origin);
                return;
            }

            if (packet.peer) { // node packet
                self.nodes.push(packet.peer);
                return;
            }

            console.log("packet not handled: " + JSON.stringify(packet));
        });
    },

    methods: {
        sendMsg: function () {
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

        sendNode: function () {
          if (this.nodeBox !== '') {
              const nodePacket = {
                  'post-node': {
                      'node': this.nodeBox
                  }
              };
              this.ws.send(JSON.stringify(nodePacket));
              this.nodeBox = ''; // Reset chatBox
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
