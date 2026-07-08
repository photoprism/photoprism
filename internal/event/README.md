## PhotoPrism — Event System

**Last Updated:** June 23, 2026

### Overview

`internal/event` provides a lightweight pub/sub hub for in-process notifications. It underpins logging hooks, UI notifications, and domain events (entities created/updated/deleted/archived/restored). The package aliases the `hub` library to keep a stable interface while exposing simple helpers for common topics.

### Usage

Publish a custom event:
```go
event.Publish("photos.updated", event.Data{"ids": []string{"p1", "p2"}})
```

Publish localized notifications:
```go
event.SuccessMsg(i18n.MsgIndexingCompletedIn, elapsed)
event.Warn("low disk space")
```

The `*Msg` helpers (`SuccessMsg`/`ErrorMsg`/`InfoMsg`/`WarnMsg`) publish a structured payload
`Data{"message", "messageId", "messageParams"}`: `message` is the server-rendered string in the
instance locale, `messageId` is the untranslated source string (`i18n.Source(id)`), and
`messageParams` are the substitution values. The Web UI renders the notification from `messageId` +
`messageParams` in each user's current UI language; `message` is a fallback. The plain
`Success`/`Error`/`Info`/`Warn` string forms publish only `message` and are **not** localized —
reserve them for already-translated or non-user-facing text.

Subscribe to topics:
```go
sub := event.Subscribe("photos.*")
defer event.Unsubscribe(sub)
for msg := range sub.Receiver {
    fmt.Printf("topic=%s payload=%v\n", msg.Name, msg.Fields)
}
```

Log hook (used by default logger):
```go
hook := event.NewHook(event.SharedHub())
log.AddHook(hook)
```

Entity events (content-channel payloads carry only identity strings — UIDs/slugs — never entity bodies):
```go
event.EntitiesUpdated("photos", []string{photo.PhotoUID})
event.EntitiesDeleted("labels", []string{label.LabelUID})
```

### Package Layout (Code Map)

- Hub aliases & helpers: `hub.go`, `format.go`, `time.go`
- Logging hook: `log.go`
- Publish helpers: `publish.go`, `publish_entities.go`
- Tests: package-level tests alongside sources

### Related Packages

- `internal/photoprism` — core indexing/import flows that emit events.
- `internal/server` — HTTP layer that may consume event notifications.
- `internal/ai/vision` & `internal/ffmpeg` — emit log events via the shared logger.
- External hub library: `github.com/leandro-lugaresi/hub`

### Testing

- Lint: `golangci-lint run ./internal/event...`
- Unit tests: `go test ./internal/event/...` (lightweight)

### Notes

- Use `SharedHub()` for process-wide subscriptions; `NewHub()` when isolating tests.
- Topic separator is `.`; message separator for rendering is ` › `.
- Keep notifications human-readable; payloads should be small to avoid blocking subscribers.
