<template>
  <div class="row">
    <div class="input-field col s7">
      <input type="text" :placeholder="placeholderText" v-model="inputText" @keyup.enter="actionAndReset">
    </div>

    <div class="input-field col s5">
      <button class="waves-effect waves-light btn" @click="actionAndReset">
        <i class="material-icons right">{{buttonIcon}}</i>
        {{buttonText}}
      </button>
    </div>
  </div>
</template>

<script>
    export default {
      name: "SimpleInput",
      data () {
        return {
          inputText: '',
        }
      },

      props: {
        placeholderText: {
          type: String,
          default: "Type here...",
        },
        buttonText: {
          type: String,
          default: "Button"
        },
        buttonIcon: {
          type: String
        },
        buttonAction: {
          type: Function,
        },
      },

      methods: {
        actionAndReset: function() {
          if (this.inputText !== '') {
            if (typeof this.buttonAction === "function") {
              this.buttonAction(this.inputText);
            } else {
              console.error("not action defined for this button")
            }
            this.inputText = '';
          }
        },
      },
    }
</script>

<style scoped>

</style>
