<template>
  <div>
      <h3 v-if="userName">Welcome, {{ userName }}</h3>
      <button v-if="!loggedIn" v-on:click="login">Login with Deezer</button>
      <button v-if="loggedIn" v-on:click="logout">Log out of Deezer</button>
      <div id="dz-root"></div>
  </div>
</template>

<script>
/* global DZ */
export default {
  name: 'DeezerPlayer',
  props: {
    appId: String
  },
  data() {
    return {
      loggedIn: false,
      userName: ''
    }
  },
  methods: {
    updateLoginStatus() {
      DZ.getLoginStatus((res) => {
        if(res.authResponse) {
          this.loggedIn = true
          DZ.api('/user/me', (res) => {
            this.userName = res.name
          })
        } else {
          this.loggedIn = false
          this.userName = ''
        }
      })
    },
    login() {
      DZ.login(() => {
        this.updateLoginStatus()
      }, {
        perms: 'basic_access'
      })
    },
    logout() {
      DZ.logout(() => {
        this.updateLoginStatus()
      })
    },
    play() {
      DZ.player.play()
    },
    pause() {
      DZ.player.pause()
    },
    next() {
      DZ.player.next()
    },
    prev() {
      DZ.player.prev()
    },
    setVolume(volume) {
      DZ.player.setVolume(volume)
    },
    setPosition(position) {
      DZ.player.seek(position)
    },
    setQueue(tracks) {
      if(!Array.isArray(tracks)) {
        //The parameter is not an array, meaning we reset the playback to a single song
        DZ.player.playTracks([])
        const track = parseInt(tracks)
        DZ.player.addToQueue([track])
        return
      }
      const currentTracks = DZ.player.getTrackList()
      if(this.checkCorrectOrder(currentTracks, tracks)) {
        var tracksToAdd = tracks.slice(currentTracks.length - 1)
        DZ.player.addToQueue(tracksToAdd)
        return
      }
      //Incorrect order, reset the playback and set a new queue
      DZ.player.playTracks(tracks)
    },
    checkCorrectOrder(oldTracks, newTracks) {
      for(const track of oldTracks) {
        if(track.track.id != newTracks[track.index]) {
          return false
        }
      }
      return true
    }
  },
  mounted() {
    //Loads an invisible Deezer player inside of dz-root
    DZ.init({
      appId: this.appId,
      channelUrl: window.location.origin + '/channel.html',
      player: {
        onload: function() {}
      }
    })
  }
}
</script>


