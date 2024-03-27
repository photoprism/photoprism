<template>
  <div class="p-help-websockets">
    <h3>PhotoPrism uses <a target="_blank" href="https://en.wikipedia.org/wiki/WebSocket">WebSockets</a> for two-way interactive communication between your browser and the server</h3>
    <p>If the connection fails and can't be reestablished, your browser won't receive photo counts, log messages, or metadata updates.</p>
    <p>To fix this issue, please make sure you didn't block WebSocket connections in your browser or firewall and check the configuration of any Web server that is in front of PhotoPrism:</p>
    <v-expansion-panel class="elevation-0">
      <v-expansion-panel-content class="secondary mb-1">
        <template #header>
          <div>How to configure NGINX to proxy WebSockets</div>
        </template>
        <v-card class="secondary-light">
          <v-card-text>
            <p>This <a target="_blank" href="https://www.serverlab.ca/tutorials/linux/web-servers-linux/how-to-configure-nginx-for-websockets/" class="text-link">tutorial</a> explains, how to configure NGINX WebSocket connections between your client and backend services.</p>
            <p>You may enable WebSockets and transparent proxying as show in this example:</p>
            <pre>
http {
  server {
    server_name example.com
    client_max_body_size 500M;
    
    location / {
      proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
      proxy_set_header Host $host;

      proxy_pass http://photoprism:2342;

      proxy_buffering off;
      proxy_http_version 1.1;
      proxy_set_header Upgrade $http_upgrade;
      proxy_set_header Connection "upgrade";
    }
  }
}
            </pre>
            <p>Please refer to the <a target="_blank" href="https://nginx.org/en/docs/" class="text-link">official documentation</a> for more details.</p>
          </v-card-text>
        </v-card>
      </v-expansion-panel-content>
      <v-expansion-panel-content class="secondary mb-1">
        <template #header>
          <div>How to configure Caddy 1 to proxy WebSockets</div>
        </template>
        <v-card class="secondary-light">
          <v-card-text>
            <p>You may enable WebSockets and transparent proxying as show in this example:</p>
            <pre>
example.com {
    proxy / photoprism:2342 {
        websocket
        transparent
    }
}
            </pre>
            <p>Please refer to the <a target="_blank" href="https://caddyserver.com/v1/docs/websocket" class="text-link">official documentation</a> for more details.</p>
          </v-card-text>
        </v-card>
      </v-expansion-panel-content>
      <v-expansion-panel-content class="secondary mb-1">
        <template #header>
          <div>How to configure Caddy 2 to proxy WebSockets</div>
        </template>
        <v-card class="secondary-light">
          <v-card-text>
            <p>WebSocket proxying automatically works in Caddy 2. There is no need to enable this, as necessary for Caddy 1, Apache, and NGINX.</p>
            <p>Example configuration:</p>
            <pre>
example.com {
    reverse_proxy photoprism:2342
}
            </pre>
            <p>In addition, Caddy 2 may automatically create and update Let's Encrypt HTTPS certificates. Please refer to the <a target="_blank" href="https://caddyserver.com/docs/v2-upgrade#proxy" class="text-link">official documentation</a> for more details.</p>
          </v-card-text>
        </v-card>
      </v-expansion-panel-content>
      <v-expansion-panel-content class="secondary mb-1">
        <template #header>
          <div>How to reverse proxy WebSockets with Apache 2.4</div>
        </template>
        <v-card class="secondary-light">
          <v-card-text>
            <p>You may enable WebSockets and transparent proxying as show in this example:</p>
            <pre>
RewriteEngine on
RewriteCond %{HTTP:Upgrade} websocket [NC]
RewriteCond %{HTTP:Connection} upgrade [NC]
RewriteRule ^/?(.*) "ws://photoprism:2342/$1" [P,L]

ProxyPass / http://photoprism:2342/
ProxyPassReverse / http://photoprism:2342/
ProxyRequests off
            </pre>
            <p>You will have to enable the <code>proxy_wstunnel</code> module:</p>
            <pre>
              a2enmod proxy_wstunnel
            </pre>
            <p>Then you will have to reload Apache.</p>
            <p>The <a target="_blank" href="https://httpd.apache.org/docs/2.4/mod/mod_proxy_wstunnel.html" class="text-link">official documentation</a> explains in detail, how to configure Apache Web Server 2.4 to reverse proxy WebSockets.</p>
          </v-card-text>
        </v-card>
      </v-expansion-panel-content>
    </v-expansion-panel>
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
