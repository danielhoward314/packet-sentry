#!/bin/sh

set -e

echo "Generating config.js from environment..."

cat <<EOF > /usr/share/nginx/html/config.js
window.__ENV__ = {
  API_BASE_URL: "${API_BASE_URL}"
};
EOF

exec "$@"
