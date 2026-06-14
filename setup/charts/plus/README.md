## Introduction

This chart provides an easy way to run [PhotoPrism®](https://www.photoprism.app/editions#compare) on a home lab or small business Kubernetes cluster while maintaining an approachable configuration.

## Highlights

- Runs the official `photoprism/photoprism` image with non-root defaults (UID/GID 1000).
- Ships with SQLite out of the box; switch to MariaDB/MySQL by filling in the database section.
- Creates lightweight PVCs for `/photoprism/storage` (5 GiB) and, optionally, `/photoprism/originals` (10 GiB) – the cluster’s default storage class is used unless you override it.
- Surfaces the most relevant customization settings (locale, themes, login footer, CORS/CDN, backup schedule) directly in Rancher’s UI or via `values.yaml`.

## Quick Start

```bash
helm repo add photoprism https://charts.photoprism.app/photoprism
helm repo update photoprism
helm upgrade --install photos photoprism/photoprism-plus \
  --namespace photos --create-namespace
```

This deploys PhotoPrism with SQLite storage. To use MariaDB (recommended for larger libraries):

```bash
helm upgrade --install photos photoprism/photoprism-plus \
  --namespace photos \
  --set database.driver=mysql \
  --set database.server=mariadb.default.svc.cluster.local:3306 \
  --set database.name=photoprism \
  --set database.user=photoprism \
  --set database.password=changeme
```

## Storage & Backups

- `persistence.storage` is always created and holds the application state.
- `persistence.originals` can be disabled or redirected to an NFS export when you manage originals elsewhere.
- Backup options (`PHOTOPRISM_BACKUP_*`) are exposed so you can point scheduled backups to another path or tweak retention.

## Customization

Key values you might want to adjust:

- `config.PHOTOPRISM_SITE_TITLE`, `config.PHOTOPRISM_APP_NAME`, `config.PHOTOPRISM_DEFAULT_THEME`
- `config.PHOTOPRISM_AUTH_MODE`, `config.PHOTOPRISM_LOGIN_INFO`, `config.PHOTOPRISM_PASSWORD_LENGTH`
- `config.PHOTOPRISM_CDN_URL`, `config.PHOTOPRISM_CORS_ORIGIN`
- `config.PHOTOPRISM_FILES_QUOTA`, `config.PHOTOPRISM_UPLOAD_LIMIT`

See [`values.yaml`](https://github.com/photoprism/photoprism/blob/develop/setup/charts/plus/values.yaml) for the full list.

## Networking & TLS

The chart exposes PhotoPrism on TCP 2342 through a ClusterIP service. You can override the service type or enable an Ingress resource when you terminate TLS  in the cluster edge:

```yaml
service:
  type: ClusterIP
  port: 2342

ingress:
  enabled: true
  className: traefik
  hosts:
    - host: photos.example.com
      paths:
        - path: /
          pathType: Prefix
  tls:
    - hosts:
        - photos.example.com
      secretName: photos-tls
```

Because TLS typically terminates at the ingress or proxy layer, the chart keeps `PHOTOPRISM_DISABLE_TLS` set to `true`. Only enable PhotoPrism’s internal TLS if your cluster design requires end-to-end encryption and you manage the certificates yourself.

## Database Password

You can provide the database password in two ways:

1. **External secret** – set `database.passwordSecretName` to reference a pre-existing Kubernetes Secret. By default the
   chart looks for a `PHOTOPRISM_DATABASE_PASSWORD` key; set `database.passwordSecretKey` to use a different key name:
   ```yaml
   database:
     passwordSecretName: my-db-credentials
     # passwordSecretKey: DB_PASSWORD  # optional, defaults to PHOTOPRISM_DATABASE_PASSWORD
   ```
2. **Inline** – set `database.password` directly (the chart stores it in its managed Secret and references it via
   `secretKeyRef`):
   ```yaml
   database:
     password: "changeme"
   ```

If both are set, `database.passwordSecretName` takes precedence. If neither is set, no password env var is injected (
suitable for SQLite).

## Security Tips

- When `adminPassword` is left blank, a random password is generated and stored in `secret/<release>-photoprism-secrets`.
- Prefer MariaDB/MySQL for multi-user setups or large libraries, and back up both the database and storage PVCs regularly.
- If you expose PhotoPrism on the public internet, pair it with HTTPS termination (Ingress/TLS or an external proxy) and keep the container image up to date.

## Getting Support

Commercial support is available with our Starter, Business, and Enterprise team plans:

- https://www.photoprism.app/teams#compare
- https://www.photoprism.app/kb/getting-support

## PhotoPrism® Documentation

For more information on specific features, services and related resources, please refer to the other documentation available in our Knowledge Base and User Guide:

- [PhotoPrism® User Guide](https://docs.photoprism.app/user-guide/)
- [PhotoPrism® Knowledge Base](https://www.photoprism.app/kb)
