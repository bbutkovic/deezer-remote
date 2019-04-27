<template>
  <div>
    <DeezerPlayer :appId="appId" ref="player"></DeezerPlayer> 
  </div>
</template>



<script>
import DeezerPlayer from '@/components/DeezerPlayer.vue'

export default {
  name: 'player',
  data() {
    return {
      appId: process.env.VUE_APP_DEEZER_APP_ID,
      socket: null
    }
  },
  methods: {

  },
  components: {
    DeezerPlayer
  }
}
</script>
