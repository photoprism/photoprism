<template>
  <div class="p-help p-help-websockets text-ltr text-selectable">
    <h3>
      Either the connection to the server is temporarily unavailable or there is a problem with the WebSocket
      configuration if you are using a reverse proxy in front of PhotoPrism.
    </h3>

    <p>
      WebSockets are required for interactive two-way communication between your browser and the server. If the
      connection fails and can't be re-established, your browser will not receive photo counts, log messages or metadata
      updates.
    </p>
    <p>
      To resolve this issue, please ensure that WebSocket connections are not blocked by your browser, firewall, or any
      <a target="_blank" href="https://docs.photoprism.app/getting-started/proxies/traefik/" class="text-link"
        >HTTP/HTTPS reverse proxy</a
      >
      you may be connecting through, for example:
    </p>

    <v-expansion-panels variant="accordion" density="comfortable" rounded="6" class="elevation-0">
      <v-expansion-panel color="secondary" title="NGINX">
        <v-expansion-panel-text>
          <v-card color="secondary-light">
            <v-card-text>
              <p>
                WebSockets can be enabled either through the NGINX Proxy Manager UI or in the server configuration
                files, as shown in this example:
              </p>
              <pre>
http {
  server {
    listen 80 ssl;
    listen [::]:80 ssl;
    server_name example.com;
    client_max_body_size 500M;

    # With SSL via Let's Encrypt
    ssl_certificate /etc/letsencrypt/live/example.com/fullchain.pem; # managed by Certbot
    ssl_certificate_key /etc/letsencrypt/live/example.com/privkey.pem; # managed by Certbot
    include /etc/letsencrypt/options-ssl-nginx.conf; # managed by Certbot
    ssl_dhparam /etc/letsencrypt/ssl-dhparams.pem; # managed by Certbot

    location / {
      proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
      proxy_set_header Host $host;

      proxy_pass http://photoprism:2342;

      proxy_buffering off;
      proxy_http_version 1.1;
      proxy_set_header Upgrade $http_upgrade;
      proxy_set_header Connection "upgrade";

      client_max_body_size 500M;
    }
  }
}
            </pre
              >
              <p>
                Please refer to the
                <a target="_blank" href="https://nginx.org/en/docs/" class="text-link">official documentation</a> for
                further details.
              </p>
            </v-card-text>
          </v-card>
        </v-expansion-panel-text>
      </v-expansion-panel>
      <v-expansion-panel color="secondary" title="Caddy 1">
        <v-expansion-panel-text>
          <v-card color="secondary-light">
            <v-card-text>
              <p>
                If you are using Caddy 1 as reverse proxy, you can allow WebSocket connections as shown in this example:
              </p>
              <pre>
example.com {
    proxy / photoprism:2342 {
        websocket
        transparent
    }
}
            </pre
              >
              <p>
                Please refer to the
                <a target="_blank" href="https://caddyserver.com/v1/docs/websocket" class="text-link"
                  >official documentation</a
                >
                for further details.
              </p>
            </v-card-text>
          </v-card>
        </v-expansion-panel-text>
      </v-expansion-panel>
      <v-expansion-panel color="secondary" title="Caddy 2">
        <v-expansion-panel-text class="bg-secondary-light">
          <v-card color="secondary-light">
            <v-card-text>
              <p>
                WebSocket proxying automatically works in Caddy 2. There is no need to enable this as necessary for
                Caddy 1, Apache, and NGINX. In addition, Caddy 2 may automatically create and update Let's Encrypt HTTPS
                certificates.
              </p>
              <p>Example configuration:</p>
              <pre>
example.com {
    reverse_proxy photoprism:2342
}
              </pre>
              <p>
                In addition, Caddy 2 may automatically create and update Let's Encrypt HTTPS certificates. Please refer
                to the
                <a target="_blank" href="https://caddyserver.com/docs/v2-upgrade#proxy" class="text-link"
                  >official documentation</a
                >
                for further details.
              </p>
            </v-card-text>
          </v-card>
        </v-expansion-panel-text>
      </v-expansion-panel>
      <v-expansion-panel color="secondary" title="Apache">
        <v-expansion-panel-text>
          <v-card color="secondary-light">
            <v-card-text>
              <p>
                If you are using Apache 2.4 as reverse proxy, you can allow WebSocket connections as shown in this
                example:
              </p>
              <pre>
RewriteEngine on
RewriteCond %{HTTP:Upgrade} websocket [NC]
RewriteCond %{HTTP:Connection} upgrade [NC]
RewriteRule ^/?(.*) "ws://photoprism:2342/$1" [P,L]

ProxyPass / http://photoprism:2342/
ProxyPassReverse / http://photoprism:2342/
ProxyRequests off
              </pre>
              <p>
                In addition, you may need to enable the <code>proxy_wstunnel</code> module using the following command:
              </p>
              <pre>
a2enmod proxy_wstunnel
              </pre>
              <p>
                The
                <a
                  target="_blank"
                  href="https://httpd.apache.org/docs/2.4/mod/mod_proxy_wstunnel.html"
                  class="text-link"
                  >official documentation</a
                >
                explains in detail, how to configure Apache Web Server 2.4 to reverse proxy WebSockets.
              </p>
            </v-card-text>
          </v-card>
        </v-expansion-panel-text>
      </v-expansion-panel>
    </v-expansion-panels>
  </div>
</template>

<script>
export default {
  name: "PHelpWebsockets",
  data() {
    return {};
  },
  methods: {},
};
</script>
