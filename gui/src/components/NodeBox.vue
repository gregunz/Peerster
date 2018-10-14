<template>
  <div>
    <b-container>
      <b-row>
        <b-col>
          <ul id="nodes">
            <li v-for="(n, idx) in nodes" :key="idx">
              <h2>{{ n.name }}</h2>
            </li>
          </ul>
        </b-col>
        <b-col>
          <input v-model="newParticipant" placeholder="Add a new node">
          <button v-on:click="onAddNewNode(newParticipant)">ADD</button></b-col>
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
    onAddNewNode: {
      type: Function,
      required: true
    }
  },
  data () {
    return {
      newParticipant: ''
    }
  },
  methods: {
  }
}
</script>

<style scoped>

</style>
