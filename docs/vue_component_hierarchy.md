## Vue Component Hierarchy

I wanted a top navbar that was rendered across all authenticated routes and would not be rendered for unprotected routes, such the `/signup` and `/login` routes.

The `App.vue` file, the one that defines the main Vue component, has the following template:

```vue
<template>
  <component :is="layout">
    <RouterView />
  </component>
</template>
```

I use the `layout` property to conditionally render either a blank layout or a main layout. It is a computed property derived from a meta field I set in the vue-router:

```javascript
  setup() {
    const route = useRoute()
    const layout = computed(() => (route.meta.layout === 'blank' ? BlankLayout : MainLayout))

    return { layout }
  }
```

Both of these layout components have a `slot` for rendering the child components determined by the vue-router. The only difference is that the blank layout has nothing else in its template, whereas the main layout also has the top navbar.

```vue
<!-- BlankLayout.vue -->
<template>
  <main>
    <slot></slot>
  </main>
</template>
```

```vue
<!-- MainLayout.vue -->
<template>
  <div>
    <MainNavigation>
        <!-- abridged...more children -->
    </MainNavigation>
    <main>
      <slot></slot>
    </main>
  </div>
</template>
```

The vue-router sets the `layout` meta field only for the exception routes, the ones that need the blank layout: 

```javascript
    {
      path: '/signup',
      name: 'signup',
      component: SignupView,
      meta: { layout: 'blank' }
    },
    {
      path: '/login',
      name: 'login',
      component: LoginView,
      meta: { layout: 'blank' }
    },
```