<template>
  <div class="chat-component">

    <div class="row">

      <div class="col s12">
        <h5 v-if="title">{{title}}</h5>
        <h5 v-if="dest">Private chat with <chip :name="dest"></chip></h5>
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

      <div class="input-field col s9">
        <input type="text" v-model="chatBox" @keyup.enter="sendMsgAndResetChatBox">
      </div>

      <div class="input-field col s3">
        <button class="waves-effect waves-light btn" @click="sendMsgAndResetChatBox">
          <i class="material-icons right">chat</i>
          Send
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
        dest: String,
        myOrigin: String,
        chatMessages: Array,
        sendMsg: Function,
      },

      methods: {
        sendMsgAndResetChatBox: function () {
          if (this.chatBox !== '') {
            this.sendMsg(this.chatBox, this.dest);
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
