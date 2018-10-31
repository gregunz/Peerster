<template>
  <div class="chat-component">

    <div class="row">

      <div class="col s12">
        <h4>{{ title }} Chat</h4>
      </div>

      <div class="col s12">
        <div class="card horizontal">
          <div id="chat-messages" class="card-content">
            <div v-for="message in chatMessages" :class="[(message.origin === myOrigin) ? 'chat-message-me' : 'chat-message-others']" >
              <chip :name="message.origin"></chip>
              <span v-html="message.text"></span>
            </div>
          </div>
        </div>
      </div>

      <div class="input-field col s12">
        <input type="text" v-model="chatBox" @keyup.enter="sendMsgAndResetChatBox">
      </div>

      <div class="input-field col s12">
        <button class="waves-effect waves-light btn" @click="sendMsgAndResetChatBox">
          <i class="material-icons right">chat</i>
          Send Message
        </button>
      </div>

    </div>

  </div>
</template>

<script>
    import Chip from "./Chip";
    export default {
      name: "Chat",
      components: {Chip},
      data () {
        return {
          chatBox: "",
        }
      },

      props: {
        title: String,
        myOrigin: String,
        chatMessages: Array,
        sendMsg: Function,
      },

      methods: {
        sendMsgAndResetChatBox: function () {
          if (this.chatBox !== '') {
            this.sendMsg(this.chatBox);
            this.chatBox = ''; // Reset chatBox
          }
        },
        scrollToTop: function (id) {
          const element = document.getElementById(id);
          element.scrollTop = element.scrollHeight; // Auto scroll to the bottom
        }
      },

      watch: {
        chatMessages: function () {
          const self = this;
          setTimeout(function () {
            self.scrollToTop('chat-messages')
          }, 1);
        }
      },
    }
</script>

<style scoped>
  #chat-messages {
    min-height: 10vh;
    height: 60vh;
    width: 100%;
    overflow-y: scroll;
  }

  .chat-message-me {
    text-align: right;
  }
  .chat-message-others {
    text-align: left;
  }

</style>
