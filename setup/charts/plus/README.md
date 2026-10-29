## Introduction

This chart provides an easy way to run [PhotoPrism®](https://www.photoprism.app/editions/#compare) on a home lab or small business Kubernetes cluster while maintaining an approachable configuration.

## Highlights

- Runs the official `photoprism/photoprism` image with non-root defaults (UID/GID 1000) and a security context that meets the Kubernetes `restricted` Pod Security Standard.
- Ships with SQLite out of the box; switch to MariaDB/MySQL by filling in the database section.
- Creates lightweight PVCs for `/photoprism/storage` (5 GiB) and, optionally, `/photoprism/originals` (10 GiB) – the cluster’s default storage class is used unless you override it, and the volumes are kept when you uninstall the release.
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
- `persistence.keep` (default `true`) keeps both volume claims when the release is uninstalled or its Argo CD application is deleted, see [Uninstalling](#uninstalling).
- Backup options (`PHOTOPRISM_BACKUP_*`) are exposed so you can point scheduled backups to another path or tweak retention.

## Customization

Key values you might want to adjust:

- `config.PHOTOPRISM_SITE_TITLE`, `config.PHOTOPRISM_APP_NAME`, `config.PHOTOPRISM_DEFAULT_THEME`
- `config.PHOTOPRISM_AUTH_MODE`, `config.PHOTOPRISM_LOGIN_INFO`, `config.PHOTOPRISM_PASSWORD_LENGTH`
- `config.PHOTOPRISM_CDN_URL`, `config.PHOTOPRISM_CORS_ORIGIN`
- `config.PHOTOPRISM_FILES_QUOTA`, `config.PHOTOPRISM_UPLOAD_LIMIT`

See [`values.yaml`](https://github.com/photoprism/photoprism/blob/develop/setup/charts/plus/values.yaml) for the full list.

## Networking & TLS

The chart exposes PhotoPrism on TCP 2342 through a ClusterIP service. You can override the service type or enable an Ingress resource when you terminate TLS at the cluster edge:

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

If both are set, `database.passwordSecretName` takes precedence and the inline `database.password` is ignored (it is not
copied into the managed Secret). If neither is set, no password env var is injected (suitable for SQLite).

## Container Startup & Security Context

By default, the chart runs the PhotoPrism binary directly (`command: ["/opt/photoprism/bin/photoprism"]`, `args: ["start"]`) as a non-root user with `allowPrivilegeEscalation: false`, all capabilities dropped, and the `RuntimeDefault` seccomp profile. This is compatible with the Kubernetes Pod Security Standard `restricted` and the OpenShift `restricted-v2` SCC.

Because the image's init script does not run, the following are not available:

- `PHOTOPRISM_INIT` actions such as installing drivers or updates. Update by pulling a newer image tag instead.
- `PHOTOPRISM_UID`, `PHOTOPRISM_GID`, and `PHOTOPRISM_UMASK`. Use `securityContext.runAsUser`, `securityContext.runAsGroup`, and `podSecurityContext.fsGroup` instead.
- Automatic permission updates. Existing volumes, for example on NFS, must be readable and writable by the configured user and group.

No hardware devices are mounted by default. For hardware video transcoding, mount the device (for example with a device plugin) and make sure the pod is a member of the group that owns it on the node, for example:

```yaml
podSecurityContext:
  fsGroup: 1000
  supplementalGroups: [44]   # group that owns /dev/dri/renderD* on the node
```

On OpenShift, remove `runAsUser`, `runAsGroup`, and `fsGroup` from your values so that the SCC can assign IDs from the namespace range.

If your cluster allows it and you need the init script, you can use the image entrypoint instead. It needs to escalate privileges at startup, so the security context must be relaxed as well, which is not permitted under the `restricted` standard:

```yaml
command: []
args: []
securityContext:
  allowPrivilegeEscalation: true
  capabilities: null   # removes the default "drop: [ALL]"
```

The other `securityContext` defaults, such as `runAsUser`, are kept, since Helm merges your values with the chart defaults.

## Security Tips

- When `adminPassword` is left blank, a random password is generated and stored in `secret/<release>-photoprism-secrets`.
- Prefer MariaDB/MySQL for multi-user setups or large libraries, and back up both the database and storage PVCs regularly.
- If you expose PhotoPrism on the public internet, pair it with HTTPS termination (Ingress/TLS or an external proxy) and keep the container image up to date.

## Health Checks

The chart adds startup and readiness probes on PhotoPrism's status endpoint (`/api/v1/status`, below the path of `PHOTOPRISM_SITE_URL` if it has one), so the pod counts as ready only once PhotoPrism is listening and its database is available. The startup probe allows up to 10 minutes for a first start, for example for database migrations. A liveness probe is available with `probes.liveness.enabled: true`; it is off by default, since indexing large files can delay responses. Set `probes.enabled: false` to remove all probes.

## Uninstalling

```bash
helm uninstall photos --namespace photos
```

This removes the release, including its secret with the generated admin password, but keeps the volume claims and with them your originals, index and thumbnails (`persistence.keep: true`). The same applies when you delete the release's application in Argo CD. Installing the chart again with the same release name and namespace reuses the volumes. Once you no longer need the data, delete the claims yourself, for example with `kubectl delete pvc --namespace photos -l app.kubernetes.io/instance=photos`.

## Getting Support

Commercial support is available with our Starter, Business, and Enterprise team plans:

- https://www.photoprism.app/teams/#compare
- https://www.photoprism.app/kb/getting-support/

## PhotoPrism® Documentation

For more information on specific features, services and related resources, please refer to the other documentation available in our Knowledge Base and User Guide:

- [PhotoPrism® User Guide](https://docs.photoprism.app/user-guide/)
- [PhotoPrism® Knowledge Base](https://www.photoprism.app/kb/)
