I followed the Shadcn [documentation](https://ui.shadcn.com/docs/installation/vite) for a Vite build.

```bash
pnpm create vite@latest # answers to prompts: packet-sentry-web-console, React, TypeScript
cd packet-sentry-web-console
```

Replace everything in `index.css` with this:

```css
@import "tailwindcss";
```

Edit `tsconfig.json` file (`compilerOptions` added):

```json
{
  "files": [],
  "references": [
    {
      "path": "./tsconfig.app.json"
    },
    {
      "path": "./tsconfig.node.json"
    }
  ],
  "compilerOptions": {
    "baseUrl": ".",
    "paths": {
      "@/*": ["./src/*"]
    }
  }
}
```

I also added `"strict": true` to the compiler options.

Edit `tsconfig.app.json` file, adding this:

```json
    "baseUrl": ".",
    "paths": {
      "@/*": ["./src/*"]
    },
```

Get the types package:

```bash
pnpm add -D @types/node
```

Edit `vite.config.ts` file:

```typescript
import path from "path"
import tailwindcss from "@tailwindcss/vite"
import react from "@vitejs/plugin-react"
import { defineConfig } from "vite"
 
// https://vite.dev/config/
export default defineConfig({
  plugins: [react(), tailwindcss()],
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "./src"),
    },
  },
})
```

Init the project:

```bash
pnpm dlx shadcn@latest init # answered `Neutral` to the prompt
```

For development outside of Docker container, installed dependencies:

```bash
pnpm install
pnpm run dev
```

I added:

- `nginx.conf`
- the `compose.yml` definition for this container
- the `Dockerfile`
- a `.dockerignore` file,

The entry in the hosts file for the gateway has to match the `proxy_pass` host.
The `.dockerignore` was needed because I'd built a previous version of the same container using Vite + Vue and the leftover node_modules conflicted with the ones this new build was adding.
