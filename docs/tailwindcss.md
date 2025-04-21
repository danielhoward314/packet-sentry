
## tailwindcss

The Vue SPA uses [tailwindcss](https://tailwindcss.com/docs/installation) for its styling. To install and configure it initially, I did the following:

```
cd packet-sentry-spa
npm install -D tailwindcss
npx tailwindcss init
touch ./src/tailwind-input.css ./src/assets/tailwind-output.css
```

I modified the `content` and `theme` properties within the `tailwind.config.js` file to point to the files that use CSS and to use a font family. I also added the following to the `tailwind-input.css` file:

```
@tailwind base;
@tailwind components;
@tailwind utilities;

@import url("https://fonts.googleapis.com/css2?family=Poppins:wght@400;500&display=swap");

@layer base {
    html {
      font-family: "Poppins", system-ui, sans-serif;
    }
}
```

tailwind uses utility classes and only ever generates the CSS for them when they are used. In order to generate them on the fly during development as you use more or fewer utility classes, run this command:

```
npx tailwindcss -i ./src/tailwind-input.css -o ./src/assets/tailwind-output.css --watch
```
