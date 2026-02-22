import { createRouter, createWebHashHistory } from 'vue-router'

const routes = [
  { path: '/', name: 'DramaList', component: () => import('./views/DramaList.vue') },
  { path: '/drama/:id', name: 'DramaDetail', component: () => import('./views/DramaDetail.vue') },
  { path: '/drama/:id/episode/:epId', name: 'Editor', component: () => import('./views/Editor.vue') },
]

export default createRouter({
  history: createWebHashHistory(),
  routes,
})
