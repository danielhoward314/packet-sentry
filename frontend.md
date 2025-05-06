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


For env vars:

- `compose.yml`

```
    environment:
      - API_BASE_URL=https://api.example.com
```

- `Dockerfile`

```
+ COPY entrypoint.sh /entrypoint.sh

+ RUN chmod +x /entrypoint.sh

COPY --from=builder /app/dist /usr/share/nginx/html

EXPOSE 80

+ ENTRYPOINT ["/entrypoint.sh"]
```

- `nginx.conf`

```nginx
    location /config.js {
      add_header Cache-Control "no-store";
      alias /usr/share/nginx/html/config.js;
    }
```

- `entrypoint.sh`

```bash
#!/bin/sh

set -e

echo "Generating config.js from environment..."

cat <<EOF > /usr/share/nginx/html/config.js
window.__ENV__ = {
  API_BASE_URL: "${API_BASE_URL}"
};
EOF

exec "$@"
```

Because we want to reference a custom property (`__ENV__`) on the `window` object like so:

```typescript
const apiBaseUrl = window.__ENV__?.API_BASE_URL;
```

And we get errors because that property does not exist on the standard `window` type that TypeScript is aware of:

```
Property '__ENV__' does not exist on type 'Window & typeof globalThis'.
```

We extend the type in `./packet-sentry-web-console/src/types` and add that path in the `tsconfig.app.json`.

Making Chrome trust my local self-signed certs on Pop!_OS:

```
sudo cp certs/ca.cert.pem /usr/local/share/ca-certificates/packet-sentry-dev-ca.crt
sudo update-ca-certificates
```

The nginx container needed changes to it config for the protocol scheme http -> https and also to turn off TLS verification:

```
    location /v1/ {
~       proxy_pass https://gateway.packet-sentry.local:8080;
+       proxy_ssl_verify off; # nginx won't trust self-signed certs even if system trusts them
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header x-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
```