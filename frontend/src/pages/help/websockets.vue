<template>
  <div class="p-help-websockets">
    <h3>PhotoPrism uses <a target="_blank" href="https://en.wikipedia.org/wiki/WebSocket">WebSockets</a>
      for two-way interactive communication between your browser and the server</h3>
    <p>If the connection fails and can't be reestablished, your browser won't receive photo counts, log messages, or metadata updates.</p>
    <p>To fix this issue, please make sure you didn't block WebSocket connections in your browser or firewall and check the
      configuration of any Web server that is in front of PhotoPrism:</p>
    <v-expansion-panel class="elevation-0">
      <v-expansion-panel-content class="secondary-light mb-1">
        <template v-slot:header>
          <div>How to reverse proxy WebSockets with Apache 2.4</div>
        </template>
        <v-card class="white">
          <v-card-text>
            <p>
              In this <a target="_blank" href="https://www.serverlab.ca/tutorials/linux/web-servers-linux/how-to-reverse-proxy-websockets-with-apache-2-4/">tutorial</a>, you will learn how to configure Apache Web Server 2.4 to reverse proxy WebSockets.
            </p>
            <p>
              Example configuration:
            </p>
            <pre>
RewriteEngine on
RewriteCond ${HTTP:Upgrade} websocket [NC]
RewriteCond ${HTTP:Connection} upgrade [NC]
RewriteRule .* "ws:/photoprism:2342/$1" [P,L]

ProxyPass / http://photoprism:2342/
ProxyPassReverse / http://photoprism:2342/
ProxyRequests off
            </pre>
          </v-card-text>
        </v-card>
      </v-expansion-panel-content>
      <v-expansion-panel-content class="secondary-light mb-1">
        <template v-slot:header>
          <div>How to configure NGINX to proxy WebSockets</div>
        </template>
        <v-card class="white">
          <v-card-text>
            <p>
              In this <a target="_blank" href="https://www.serverlab.ca/tutorials/linux/web-servers-linux/how-to-configure-nginx-for-websockets/">tutorial</a>, you will learn how to configure NGINX WebSocket connections between your client and backend services.
            </p>
            <p>
              Example configuration:
            </p>
            <pre>
http {
  server {
    server_mame example.com

    location / {
      proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
      proxy_set_header Host $host;

      proxy_pass http://photoprism:2342;

      proxy_http_version 1.1;
      proxy_set_header Upgrade $http_upgrade;
      proxy_set_header Connection "upgrade";
    }
  }
}
            </pre>
          </v-card-text>
        </v-card>
      </v-expansion-panel-content>
      <v-expansion-panel-content class="secondary-light mb-1">
        <template v-slot:header>
          <div>How to configure Caddy 1 to proxy WebSockets</div>
        </template>
        <v-card class="white">
          <v-card-text>
            <p>
              Example configuration:
            </p>
            <pre>
example.com {
    proxy / photoprism:2342 {
        websocket
        transparent
    }
}
            </pre>
            <p>
              See <a target="_blank" href="https://caddyserver.com/v1/docs/websocket">documentation</a> for details.
            </p>
          </v-card-text>
        </v-card>
      </v-expansion-panel-content>
      <v-expansion-panel-content class="secondary-light mb-1">
        <template v-slot:header>
          <div>How to configure Caddy 2 to proxy WebSockets</div>
        </template>
        <v-card class="white">
          <v-card-text>
            <p>
              WebSocket proxying &quot;just works&quot; in Caddy 2; there is no need to &quot;enable&quot; websockets like for Caddy 1.
            </p>
            <p>
              Example configuration:
            </p>
            <pre>
example.com {
    reverse_proxy photoprism:2342
}
            </pre>
            <p>
              See <a target="_blank" href="https://caddyserver.com/docs/v2-upgrade#proxy">documentation</a> for details.
            </p>
          </v-card-text>
        </v-card>
      </v-expansion-panel-content>
    </v-expansion-panel>
  </div>
</template>

<script>
    export default {
        name: 'p-help-websockets',
        data() {
            return {
            };
        },
        methods: {
        },
    };
</script>
