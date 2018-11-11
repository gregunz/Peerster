<template>
  <div class="row">
    <div class="input-field col s12">

      <select v-model="destination">
        <option disabled value="">Select contact...</option>
        <option v-for="contact in contacts" :value="contact.name">{{contact.name}}</option>
      </select>

      <input type="text" placeholder="filename... (required)" v-model="filename">

      <input type="text" placeholder="hash... (required)" v-model="request">

      <button class="waves-effect waves-light btn" @click="sendFileRequestAndReset">
        <i class="material-icons right">cloud_download</i>
        Request
      </button>

    </div>
  </div>
</template>

<script>
    export default {

      name: "RequestFile",

      components: {},

      data () {
        return {
          destination: "",
          filename: "",
          request: "",
        }
      },

      props: {
        contacts: {
          type: Array,
          default: [],
        },
        sendFileRequest: {
          type: Function
        },
      },

      methods: {
        sendFileRequestAndReset: function() {
          if (this.destination !== "" && this.filename !== "" && this.request !== "") {
            if (typeof this.sendFileRequest === "function") {
              this.sendFileRequest(this.destination, this.filename, this.request);
            } else {
              console.error("not action defined for this button")
            }
            this.destination = "";
            this.filename = "";
            this.request = "";
          }
        },
      },
    }
</script>

<style scoped>
select {
  display: block;
}
</style>
