<template>
  <div>
    <h2>List of known peers</h2>
    <b-container>
      <b-row>
        <b-col>
          <ul id="nodes">
            <li v-for="(n, idx) in nodes" :key="idx">
              <h3>{{ n }}</h3>
            </li>
          </ul>
        </b-col>
        <b-col>
          <input v-model="node" placeholder="Type here...">
          <button v-on:click="onNewNodeSent(node)">ADD PEER</button></b-col>
      </b-row>
    </b-container>
  </div>
</template>

<script>
export default {
  name: 'NodeBox',
  props: {
    nodes: {
      type: Array,
      required: true
    },
    onNewNodeSent: {
      type: Function,
      required: true
    }
  },
  data () {
    return {
      node: ''
    }
  },
  methods: {
  }
}
</script>

<style scoped>

</style>
