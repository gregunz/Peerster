<template>
    <div class="chip">
      <img :src="avatarURL(name)"/>
      {{ name }}
    </div>
</template>

<script>
    export default {
      name: "Chip",

      props: {
        name: String,
      },

      methods: {
        avatarURL: function (name) {
          return 'https://api.adorable.io/avatars/100/' + name;
        },
      },
    }
</script>

<style scoped>

</style>
