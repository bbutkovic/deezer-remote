import Vue from 'vue'
import Router from 'vue-router'
import Home from './views/Home.vue'

Vue.use(Router)

export default new Router({
  mode: 'history',
  base: process.env.BASE_URL,
  routes: [
    {
      path: '/',
      name: 'home',
      component: Home
    },
    {
      path: '/player',
      name: 'player',
      component: () => import(/* webpackChunkName: "player" */ './views/Player.vue')
    },
    {
      path: '/remote',
      name: 'remote',
      component: () => import(/* webpackChunkName: "remote" */ './views/Remote.vue')
    }
  ]
})
