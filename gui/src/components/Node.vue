<template>
  <div class="node-component">

    <div class="row">

      <div class="input-field col s12">
        <h4>Nodes</h4>
      </div>

      <div class="col s12">
        <div class="card horizontal">
          <div class="card-content">
            <div v-for="node in nodes">
              {{ node.address }}
            </div>
          </div>
        </div>
      </div>

      <div class="input-field col s8">
        <input type="text" v-model="nodeBox" @keyup.enter="sendNodeAndResetBox">
      </div>

      <div class="input-field col s4">
        <button class="waves-effect waves-light btn" @click="sendNodeAndResetBox">
          <i class="material-icons right">router</i>
          Add
        </button>
      </div>

    </div>

  </div>
</template>

<script>
    export default {
      name: "Node",

      data () {
        return {
          nodeBox: '',
        }
      },

      props: {
        nodes: Array,
        sendNode: Function,
        onContactClick: Function
      },

      methods: {
        sendNodeAndResetBox: function() {
          if (this.nodeBox !== '') {
            this.sendNode(this.nodeBox);
            this.nodeBox = '';
          }
        },
      },
    }
</script>

<style scoped>
.card-content {
  text-align: left;
}
</style>
