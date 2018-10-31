<template>
  <div class="contacts-component row">

    <div class="col s12">
      <h4>Contacts</h4>
    </div>

    <div class="col s12">
      <div class="card horizontal">
        <div class="card-content">
          <div v-for="o of origins" v-on:click="onContactClick(o)">
            <chip :name="o"></chip>
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
        origins: Array,
        onContactClick: Function,
      },
    }
</script>

<style scoped>
.card-content{
  text-align: left;
}
</style>
