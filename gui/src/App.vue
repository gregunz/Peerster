<template>
  <div id="app">
    <header>
      <nav>
        <div class="nav-wrapper">
          <div class="app-title brand-logo left">
            <a href="/"><img id="logo" src="./static/logo_100x100_white.png"/> Peerster App</a>
          </div>
          <a v-if="myOrigin !== ''" href="/" class="brand-logo right">connected as <chip :name="myOrigin"></chip></a>
        </div>
      </nav>
    </header>

    <div class="row">
      <div class="col s8">
        <chat v-if="selectedDest === ''" title="Rumors Chat" :my-origin="myOrigin" :chat-messages="rumors" :send-msg="sendRumor"></chat>
        <chat v-else :dest="selectedDest" :my-origin="myOrigin" :chat-messages="selectedDestChat" :send-msg="sendPrivate"></chat>
      </div>
      <div class="col s4">
        <contacts :origins-and-badges="originsAndBadges" :on-contact-click="onContactClick"></contacts>
        <div class="back-button" v-if="selectedDest !== ''">
          <button class="waves-effect waves-light btn" @click="backToRumors()">
            <i class="material-icons right"></i>
            Back to Rumors
          </button>
        </div>
        <list-with-input
          :list="nodes"
          :elem-to-string="nodeToString"
          :button-action="sendNode"
          title="Nodes"
          button-text="Add"
          button-icon="router">
        </list-with-input>
        <list-with-input
          :list="files"
          :button-action="indexFile"
          title="Indexed Files"
          button-text="Index"
          button-icon="create_new_folder">
        </list-with-input>
      </div>
    </div>


  </div>
</template>

<script>
import Chat from "./components/Chat";
import ListWithInput from "./components/ListWithInput"
import Contacts from "./components/Contacts";
import Chip from "./components/Chip";
import swal from "sweetalert2"
import axios from "axios";

export default {
  name: 'app',
  components: {Chip, Contacts, Chat, ListWithInput},
  data () {
    return {
      apiURL: '',
      ws: null, // websocket
      rumors: [],
      privates: new Map(),
      selectedDest: '',
      selectedDestChat: [],
      nodes: [],
      files: [],
      originsAndBadges: [],
      myOrigin: '',
    }
  },

  created: function() {
    const self = this;
    this.apiURL = window.location.host;

    axios.get('http://' + this.apiURL + '/id')
      .then(function (response) {
        const packet = response.data;

        self.handleGetId(packet['get-id']);

        self.ws = new WebSocket('ws://' + self.apiURL + '/ws');

        self.ws.addEventListener('open', function () {
          self.ws.send(JSON.stringify({'subscribe-message' : {'subscribe': true, 'with-previous': true}})); // subscribing to messages
          self.ws.send(JSON.stringify({'subscribe-node' : {'subscribe': true, 'with-previous': true}})); // subscribing to nodes
          self.ws.send(JSON.stringify({'subscribe-origin' : {'subscribe': true, 'with-previous': true}})); // subscribing to origins
          self.ws.send(JSON.stringify({'subscribe-file' : {'subscribe': true, 'with-previous': true}})); // subscribing to files
        });

        self.ws.addEventListener('close', function () {
          swal({
            type: 'error',
            title: 'server unavailable, try again...',
            width: 600,
            padding: '3em',
            backdrop: `
          rgba(0,0,123,0.4)
          url("https://sweetalert2.github.io/images/nyan-cat.gif")
          center left
          no-repeat
        `,
            showConfirmButton: false,
            allowOutsideClick: false,
            allowEscapeKey: false,
          })
        });

        self.ws.addEventListener('message', function (e) { // here we receive packets from the websocket

          const packet = JSON.parse(e.data);
          console.log(packet);

          if (packet['get-id']) {
            //self.handleGetId(packet['get-id']);
            console.log('not expection get-id message');
            return;
          }

          if (packet.rumor) { // rumor message packet
            self.rumors.push(packet.rumor);
            return;
          }

          if (packet.private) {
            const msg = packet.private;
            self.handlePrivateMsg(msg);
            return;
          }

          if (packet.peer) { // node packet
            self.handlePeer(packet.peer);
            return;
          }

          if (packet.contact) {
            self.handleContact(packet.contact.origin);
            return;
          }

          if (packet.file) {
            self.handleFile(packet.file);
            return;
          }

          console.error("packet not handled: " + JSON.stringify(packet));
        });

      });
  },

  methods: {
    sendRumor: function (rumorText) {
      const msgPacket = {
        'post-message': {
          'message': rumorText,
        }
      };
      this.ws.send(JSON.stringify(msgPacket));
    },

    sendPrivate: function (privateText, dest) {
      const msgPacket = {
        'post-message': {
          'message': privateText,
          'destination': dest,
        }
      };
      this.ws.send(JSON.stringify(msgPacket));
    },
    backToRumors: function () {
      this.selectedDest = '';
      this.selectedDestChat = null;
    },

    nodeToString: function (node) {
      return node.address;
    },

    sendNode: function (nodeText) {
      const nodePacket = {
        'post-node': {
          'node': nodeText
        }
      };
      this.ws.send(JSON.stringify(nodePacket));
    },

    handleGetId: function (get_id) {
      const self = this;
      self.myOrigin = get_id.id;
    },

    handlePrivateMsg: function (msg) {
      const self = this;
      if (self.privates.has(msg.origin)){
        self.addPrivateMsg(msg.origin, msg);
      } else if (self.privates.has(msg.destination)) {
        self.addPrivateMsg(msg.destination, msg);
      } else if (msg.destination === self.myOrigin) {
        self.addPrivateMsg(msg.origin, msg);
      } else if (msg.origin === self.myOrigin) {
        self.addPrivateMsg(msg.destination, msg);
      } else {
        console.error("I am <"+ self.myOrigin +"> receiving unexpected message: " + JSON.stringify(msg));
      }
    },

    addPrivateMsg: function (contactName, msg) {
      const self = this;
      const contact = self.handleContact(contactName);
      self.privates.get(contactName).messages.push(msg);
      if (msg.origin !== self.myOrigin) {
        contact.numUnread += 1;
      } else {
        contact.numUnread = 0;
      }
    },

    newPrivateObject: function (msg) {
      let messagesArray = [];
      if (msg) {
        messagesArray.push(msg);
      }
      return {
        messages: messagesArray,
      };
    },

    handlePeer: function (peer) {
      const self = this;
      self.nodes.push(peer);
    },

    handleContact: function (contactName) {
      const self = this;
      if (!self.privates.has(contactName)){
        self.privates.set(contactName, self.newPrivateObject());
        const contact = self.newContactObject(contactName);
        self.originsAndBadges.push(contact);
        return contact;
      }
      // we have the contact already if we skipped if the if clause
      return self.getContact(contactName);
    },

    newContactObject: function (contactName) {
      return {
        name: contactName,
        numUnread: 0,
      }
    },

    onContactClick: function (contactName) {
      this.selectedDest = contactName;
      this.selectedDestChat = this.privates.get(contactName).messages;
      this.getContact(contactName).numUnread = 0;
    },

    getContact: function(contactName) {
      const self = this;
      return self.originsAndBadges.find(function (contact) {
        return contact.name === contactName
      });
    },

    handleFile: function (file) {
      this.files.push(file.filename)
    },

    indexFile: function (filename) {
      const indexFile = {
        'index-file': {
          'filename': filename
        }
      };
      this.ws.send(JSON.stringify(indexFile));
    },
  },
}
</script>

<style>
#app {
  font-family: 'Avenir', Helvetica, Arial, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  text-align: center;
  color: #2c3e50;
}

body {
  display: flex;
  min-height: 100vh;
  flex-direction: column;
}

h1, h2 {
  font-weight: normal;
}

ul {
  list-style-type: none;
  padding: 0;
}

li {
  display: inline-block;
  margin: 0 10px;
}

a {
  color: #42b983;
}

main {
  flex: 1 0 auto;
}

.app-title {
  left: 10px;
}
#logo {
  height: 2.1rem;
}

.chip {
  display: inline-flex;
}

.back-button {
  /*margin-top: 50px;*/
}
</style>
