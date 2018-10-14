<template>
  <div>
    <h1>Gossiper: {{name}}</h1>
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
        :nodes="nodes"
        :onNewNodeSent="onNewNodeSent"
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
      name: '-name-',
      nodes: [],
      participants: [], // the list of all the participant of the conversation. `name` is the user name, `id` is used to establish the author of a message, `imageUrl` is supposed to be the user avatar.
      titleImageUrl: 'https://a.slack-edge.com/66f9/img/avatars-teams/ava_0001-34.png',
      messageList: [], // the list of the messages to show, can be paginated and adjusted dynamically
      newMessagesCount: 0,
      showTypingIndicator: '', // when set to a value matching the participant.id it shows the typing indicator for the specific user
      alwaysScrollToBottom: false // when set to true always scrolls the chat to the bottom when new events are in (new message, user starts typing...)
    }
  },
  methods: {
    onMessageWasSent (message) {
      // called when the user sends a message
      if (message.author === 'me') {
        axios.post('http://localhost:8080/message', {message: message.data.text})
          .then(response => {
            console.log('successfully send message, response:' + response)
            this.messageList = [...this.messageList, message]
          })
          .catch(e => {
            console.log('error when sending message:' + e)
          })
      } else {
        this.addNewParticipant(message.author)
        this.messageList = [...this.messageList, message]
      }
    },
    onNewNodeSent (node) {
      axios.post('http://localhost:8080/node', {peer: node})
        .then(response => {
          console.log('successfully send node, response:' + response)
          this.addNewNode(node)
        })
        .catch(e => {
          console.log('error when sending message:' + e)
        })
    },
    addNewParticipant (newPart) {
      if (newPart !== 'me' && newPart.length > 0) {
        let newParticipant = {
          id: newPart,
          name: newPart,
          imageUrl: 'https://api.adorable.io/avatars/60/' + newPart
        }
        if (!this.participants.find(function (e) {
          return e.id === newParticipant.id
        })) {
          this.participants.push(newParticipant)
          console.log('adding new participant:' + newPart)
        }
      }
    },
    addNewNode (node) {
      if (this.nodes.indexOf(node) === -1) {
        this.nodes.push(node)
        console.log('adding new node:' + node)
      }
    },
    getNewMessages () {
      axios.get(`http://localhost:8080/message`)
        .then(response => {
          for (let msg of response.data) {
            this.onMessageWasSent({author: msg.origin, type: 'text', data: {text: msg.text}})
          }
        })
        .catch(e => {
          console.log(e)
        })
    },
    getNewNodes () {
      axios.get(`http://localhost:8080/node`)
        .then(response => {
          for (let node of response.data) {
            this.addNewNode(node)
          }
        })
        .catch(e => {
          console.log(e)
        })
    }
  },
  created () {
    axios.get(`http://localhost:8080/id`)
      .then(response => {
        this.name = response.data
      })
      .catch(e => {
        console.log(e)
      })
    this.getNewMessages()
    this.getNewNodes()
    setInterval(() => {
      this.getNewMessages()
      this.getNewNodes()
    }, 1000)
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
