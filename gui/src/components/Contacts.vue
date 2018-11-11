<template>
  <div class="contacts-component row">

    <div class="col s12">
      <h4>Contacts</h4>
    </div>

    <div class="col s12">
      <div class="card horizontal">
        <div class="card-content">
          <div v-for="o in originsAndBadges" v-on:click="onContactClick(o.name)">
            <chip :name="o.name"></chip>
            <span class="new badge" v-if="o.numUnread > 0">{{toBadgeString(o.numUnread)}}</span>
          </div>
      </div>
      </div>
    </div>

  </div>
</template>

<script>
    import Chip from "./Chip";
    export default {
      name: "Contacts",
      components: {Chip},
      props: {
        originsAndBadges: Array,
        onContactClick: Function,
      },
      methods: {
        toBadgeString: function(numUnread) {
          if (numUnread > 9) {
            return '+';
          } else {
            return numUnread;
          }
        }
      }
    }
</script>

<style scoped>
.card-content{
  text-align: left;
}
</style>
