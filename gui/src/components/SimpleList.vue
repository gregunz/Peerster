<template>
    <div class="row">

      <div v-if="title" class="col s12">
        <h5>{{title}}</h5>
      </div>

      <div class="col s12">
        <div class="card horizontal">
          <div class="card-content">
            <div class="list-elem" v-for="e in list">
              {{ elemToString(e) }}
            </div>
          </div>
        </div>
      </div>
    </div>
</template>

<script>
    export default {
      name: "SimpleList",

      data () {
        return {
          inputText: '',
        }
      },

      props: {
        title: String,
        list: {
          type: Array,
          default: [],
        },
        elemToString: {
          type: Function,
          default: function (e) {
            return e;
          },
        },
      },
    }
</script>

<style scoped>
  .card-content {
     width: 100%;
     text-align: left;
   }
  .list-elem{
    overflow: hidden;
    text-overflow: ellipsis;
    white-space:nowrap;
  }
</style>
