<template>
  <div id="app">
    <header>
      <nav>
        <div class="nav-wrapper">
          <a href="/" class="brand-logo right">Gossiper {{ name }}</a>
        </div>
      </nav>
    </header>
    <div class="row">
      <div class="col s12">
        <div class="card horizontal">
          <div id="chat-messages" class="card-content">
            <div v-for="message in chatMessages" :class="[(message.origin === name) ? 'chat-message-me' : 'chat-message-others']" >
              <div class="chip">
                <img :src="avatarURL(message.origin)"/>
                {{ message.origin }}
              </div>
              <span v-html="message.text"></span>
            </div>
          </div>
        </div>
      </div>
    </div>
    <div class="row">
      <div class="input-field col s8">
        <input type="text" v-model="chatBox" @keyup.enter="sendMsg">
      </div>
      <div class="input-field col s4">
        <button class="waves-effect waves-light btn" @click="sendMsg">
          <i class="material-icons right">chat</i>
          Send Message
        </button>
      </div>
    </div>
    <div class="row">
      <div class="input-field col s4">
        <input type="text" v-model="nodeBox" @keyup.enter="sendNode">
      </div>
      <div class="input-field col s4">
        <button class="waves-effect waves-light btn" @click="sendNode">
          <i class="material-icons right">perm_identity</i>
          Send Node
        </button>
      </div>
    </div>
    <div class="row">
      <div v-for="node in nodes">
        {{ node.address }}
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'app',
  data () {
    return {
      apiURL: '',
      ws: null, // Our websocket
      chatBox: '', // Holds new messages to be sent to the server
      chatMessages: [], // chat messages list
      nodeBox: '',
      nodes: [],
      name: '',
    }
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
        const rumorMsg = packet.rumor;
        self.saveMsg(rumorMsg.text, rumorMsg.origin);
        return;
      }

      if (packet.private) {
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
        const msgPacket = {
          'post-message': {
            'message': this.chatBox,
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

    avatarURL: function(name) {
      return 'https://api.adorable.io/avatars/100/' + name;
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

#chat-messages {
  min-height: 10vh;
  height: 60vh;
  width: 100%;
  overflow-y: scroll;
}

.chat-message-me {
  text-align: right;
}

</style>
