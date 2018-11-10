<template>
  <div class="node-component">

    <div class="row">

      <div class="input-field col s12">
        <h4>{{title}}</h4>
      </div>

      <div class="col s12">
        <div class="card horizontal">
          <div class="card-content">
            <div v-for="e in list">
              {{ elemToString2(e) }}
            </div>
          </div>
        </div>
      </div>

      <div v-if="typeof this.buttonAction === 'function'">
        <div class="input-field col s7">
          <input type="text" v-model="inputText" @keyup.enter="actionAndReset">
        </div>

        <div class="input-field col s5">
          <button class="waves-effect waves-light btn" @click="actionAndReset">
            <i class="material-icons right">{{buttonIcon}}</i>
            {{buttonText}}
          </button>
        </div>
      </div>

    </div>

  </div>
</template>

<script>
    export default {
      name: "ListWithInput",

      data () {
        return {
          inputText: '',
        }
      },

      props: {
        title: String,
        buttonText: String,
        buttonIcon: String,
        list: Array,
        elemToString: Function,
        buttonAction: Function,
        onContactClick: Function
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
        elemToString2: function (e) {
          if (typeof this.elemToString === "function") {
            return this.elemToString(e);
          }
          return e;
        }
      },
    }
</script>

<style scoped>
.card-content {
  text-align: left;
}
</style>
