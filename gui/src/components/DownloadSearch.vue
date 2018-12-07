<template>
  <div class="row">
    <div class="input-field col s12">
      <select v-model="fileIdx">
        <option disabled :value="-1">Select file...</option>
        <option v-for="(file, idx) in filesSearched" :value="idx">{{file.filename}}</option>
      </select>

      <input type="text" placeholder="filename... (required)" v-model="filename">

      <button class="waves-effect waves-light btn" @click="sendFileRequestAndReset">
        <i class="material-icons right">cloud_download</i>
        Download
      </button>
    </div>
  </div>
</template>

<script>
  export default {
    name: "DownloadSearch",

    components: {},

    data() {
      return {
        fileIdx: -1,
        filename: "",
      }
    },

    props: {
      filesSearched: {
        type: Array,
        default: [],
      },
      sendFileRequest: {
        type: Function
      },
    },

    methods: {
      sendFileRequestAndReset: function () {
        if (this.fileIdx >= 0 && this.filename !== '') {
          if (typeof this.sendFileRequest === "function") {
            const file = this.filesSearched[this.fileIdx];
            this.sendFileRequest("", this.filename, file['meta-hash']);
            this.fileIdx = -1;
            this.filename = '';
          } else {
            console.error("not action defined for this button")
          }
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
