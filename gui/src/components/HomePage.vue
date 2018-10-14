<template>
  <div>
    <h1>Welcome {{name}}</h1>
    <div class="chat-box-container">
      <chat-box
        :participants="participants"
        :titleImageUrl="titleImageUrl"
        :onMessageWasSent="onMessageWasSent"
        :messageList="messageList"
        :newMessagesCount="newMessagesCount"
        :showEmoji="false"
        :showFile="false"
        :showTypingIndicator="showTypingIndicator"
        :alwaysScrollToBottom="alwaysScrollToBottom"></chat-box>
    </div>
    <div class="node-box-container">
      <node-box
        :nodes="participants"
        :onAddNewNode="onAddNewNode"
      ></node-box>
    </div>
  </div>
</template>

<script>
import axios from 'axios'
import ChatBox from '@/components/ChatBox'
import NodeBox from '@/components/NodeBox'

export default {
  name: 'HomePage',
  components: {NodeBox, ChatBox},
  data () {
    return {
      name: 'default name',
      participants: [], // the list of all the participant of the conversation. `name` is the user name, `id` is used to establish the author of a message, `imageUrl` is supposed to be the user avatar.
      titleImageUrl: 'https://a.slack-edge.com/66f9/img/avatars-teams/ava_0001-34.png',
      messageList: [
        {author: 'me', type: 'text', data: {text: 'default message'}}
      ], // the list of the messages to show, can be paginated and adjusted dynamically
      newMessagesCount: 0,
      showTypingIndicator: '', // when set to a value matching the participant.id it shows the typing indicator for the specific user
      alwaysScrollToBottom: false // when set to true always scrolls the chat to the bottom when new events are in (new message, user starts typing...)
    }
  },
  methods: {
    sendMessage (text) {
      if (text.length > 0) {
        this.newMessagesCount += 1
        this.onMessageWasSent({ author: 'support', type: 'text', data: {text: text} })
      }
    },
    onMessageWasSent (message) {
      // called when the user sends a message
      this.messageList = [ ...this.messageList, message ]
    },
    onAddNewNode (node) {
      if (node.length > 0) {
        let newParticipant = {
          id: node,
          name: node,
          imageUrl: 'https://api.adorable.io/avatars/60/' + node
        }
        if (!this.participants.find(function (e) { return e.id === newParticipant.id })) {
          this.participants.push(newParticipant)
          console.log(node)
        }
      }
    },
    getNewMessages () {
      axios.get(`localhost:8080/message`)
        .then(response => {
          // JSON responses are automatically parsed.
          console.log(response)
        })
        .catch(e => {
          console.log(e)
        })
    }
  }
}
</script>

<style scoped>
  .chat-box-container{
    text-align: center;
    width: 100%;
    max-width: 540px;
    height: 540px;
    margin:auto;
  }
  .node-box-container{
    width: 100%;
    max-width: 540px;
    margin: 100px auto 0;
  }
</style>
