<template>
  <div id="app">
    <header>
      <nav>
        <div class="nav-wrapper">
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
        <div class="back-button" v-if="selectedDest !== ''">
          <button class="waves-effect waves-light btn" @click="backToRumors()">
            <i class="material-icons right"></i>
            Back to Rumors
          </button>
        </div>
        <contacts :origins="origins" :on-contact-click="onContactClick"></contacts>
        <node :nodes="nodes" :send-node="sendNode"></node>
      </div>
    </div>


  </div>
</template>

<script>
import Chat from "./components/Chat";
import Node from "./components/Node"
import Contacts from "./components/Contacts";
import Chip from "./components/Chip";
import swal from "sweetalert2"


export default {
  name: 'app',
  components: {Chip, Contacts, Chat, Node},
  data () {
    return {
      apiURL: '',
      ws: null, // websocket
      rumors: [],
      privates: new Map(),
      selectedDest: '',
      selectedDestChat: [],
      nodes: [],
      origins: [],
      myOrigin: '',
      privateMsgBuffer: [],
    }
  },

  created: function() {
    const self = this;
    this.apiURL = window.location.host;
    this.ws = new WebSocket('ws://' + this.apiURL + '/ws');
    this.ws.addEventListener('open', function () {
      self.ws.send(JSON.stringify({'get-id' : {}})); // getting name
      self.ws.send(JSON.stringify({'subscribe-message' : {'subscribe': true, 'with-previous': true}})); // subscribing to messages
      self.ws.send(JSON.stringify({'subscribe-node' : {'subscribe': true, 'with-previous': true}})); // subscribing to nodes
      self.ws.send(JSON.stringify({'subscribe-origin' : {'subscribe': true, 'with-previous': true}})); // subscribing to origins
    });
    this.ws.addEventListener('close', function () {
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

    this.ws.addEventListener('message', function (e) { // here we receive packets from the websocket

      const packet = JSON.parse(e.data);
      console.log(packet);

      if (packet['get-id']) {
        self.handleGetId(packet['get-id']);
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
        self.handleContact(packet.contact);
        return;
      }

      console.log("packet not handled: " + JSON.stringify(packet));
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

    onContactClick: function (contact) {
      this.selectedDest = contact;
      this.selectedDestChat = this.privates.get(contact);
    },

    backToRumors: function () {
      this.selectedDest = '';
      this.selectedDestChat = null;
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
      self.privateMsgBuffer.forEach(function (msg) {
        self.handlePrivateMsg(msg)
      });
    },

    handlePrivateMsg: function (msg) {
      const self = this;
      if (self.privates.has(msg.origin)){
        self.privates.get(msg.origin).push(msg);
      } else if (self.privates.has(msg.destination)) {
        self.privates.get(msg.destination).push(msg);
      } else if (msg.origin === self.myOrigin) {
        self.privates.set(msg.destination, [msg])
      } else if (msg.destination === self.myOrigin) {
        self.privates.set(msg.origin, [msg])
      } else if (self.myOrigin === '') {
        self.privateMsgBuffer.push(msg);
      } else {
        console.log("strange private message: " + JSON.stringify(msg));
      }
    },

    handlePeer: function (peer) {
      const self = this;
      self.nodes.push(peer);
    },

    handleContact: function (contact) {
      const self = this;
      if (contact.origin !== self.myOrigin) {
        self.origins.push(contact.origin);
        if (!self.privates.has(contact.origin)){
          self.privates.set(contact.origin, [])
        }
      }
    }
  }
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
.back-button {
  margin-top: 50px;
}
</style>
