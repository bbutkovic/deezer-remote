<template>
  <div class="home">
    <div id="nav">
      <router-link to="/player">Open a player</router-link> | 
      <router-link to="/remote">Open a remote</router-link>
    </div>
    <img alt="Vue logo" src="../assets/logo.png">
  </div>
</template>

<script>
export default {
  name: 'home'
}
</script>
