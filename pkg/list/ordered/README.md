## PhotoPrism — Ordered List Package

**Last Updated:** November 17, 2025

### Overview

The `pkg/list/ordered` package provides an ordered associative container that
combines O(1) lookups with predictable iteration. It underpins features such as
batch photo edits where we need stable selections, repeatable JSON responses,
and fast lookups by UID while keeping the insertion order defined by the UI.

Use `Map` when you want deterministic ordering (for example, mirroring the
selection order that comes from the frontend) and `SyncMap` when multiple
goroutines need to mutate or read the same ordered state.

### Basic Usage Example

```go
package main

import (
	"fmt"

	ordered "github.com/photoprism/photoprism/pkg/list/ordered"
)

func main() {
	m := ordered.NewMap[string, int]()
	m.Set("pq1z9t3", 1)
	m.Set("px4y2k0", 2)
	m.ReplaceKey("px4y2k0", "px4y2k9")

	if v, ok := m.Get("px4y2k9"); ok {
		fmt.Println("latest selection index:", v)
	}

	fmt.Println("ordered iteration")
	for el := m.Front(); el != nil; el = el.Next() {
		fmt.Printf("%s => %d\n", el.Key, el.Value)
	}
}
```

### JSON Serialization Example

Because JSON arrays preserve order, iterating with `Front() … Next()` (or the
`Keys()` / `Values()` iterators) lets us produce deterministic responses for the
frontend REST models described in `internal/photoprism/batch/README.md`.

```go
package photorest

import (
	"encoding/json"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/list/ordered"
)

type photoDTO struct {
	UID   string `json:"UID"`
	Title string `json:"Title"`
}

func MarshalPhotosOrdered(m *ordered.Map[string, *entity.Photo]) ([]byte, error) {
	payload := make([]photoDTO, 0, m.Len())
	for el := m.Front(); el != nil; el = el.Next() {
		payload = append(payload, photoDTO{
			UID:   el.Key,
			Title: el.Value.PhotoTitle,
		})
	}
	return json.Marshal(payload)
}
```

Calling `MarshalPhotosOrdered` guarantees that the frontend receives the photos
exactly in the order users selected them, which keeps batch edit dialogs and the
REST `/api/v1/batch/photos/edit` response in sync.

### Batch Edit Integration Example

The following snippet sketches how a batch edit handler can combine
`ordered.Map`, the entity models in `internal/entity`, and the helpers from
`internal/photoprism/batch` to preload photos, build the response payload, and
still offer constant-time lookups by UID:

```go
package batchhandler

import (
	"context"
	"fmt"

	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/entity/query"
	"github.com/photoprism/photoprism/internal/entity/search"
	"github.com/photoprism/photoprism/internal/photoprism/batch"
	ordered "github.com/photoprism/photoprism/pkg/list/ordered"
)

func BuildBatchResponse(ctx context.Context, uids []string) (*batch.PhotosResponse, error) {
	photosByUID := ordered.NewMap[string, *entity.Photo]()
	preloaded := make(map[string]*entity.Photo, len(uids))
	results := make(search.PhotoResults, 0, len(uids))

	for _, uid := range uids {
		if uid == "" {
			continue
		}
		photo, err := query.PhotoPreloadByUID(uid)
		if err != nil || !photo.HasID() {
			return nil, fmt.Errorf("load photo %s: %w", uid, err)
		}

		p := photo // capture copy because query returns a value
		photosByUID.Set(uid, &p)
		preloaded[uid] = &p

		results = append(results, search.Photo{
			PhotoUID:     p.PhotoUID,
			PhotoTitle:   p.PhotoTitle,
			PhotoCaption: p.PhotoCaption,
			TakenAt:      p.TakenAt,
			TimeZone:     p.TimeZone,
		})
	}

	resp := &batch.PhotosResponse{
		Models: results,
		Values: batch.NewPhotosFormWithEntities(results, preloaded),
	}

	// Quick lookup later in the request lifecycle (album diffing, ACL checks, etc.).
	if el := photosByUID.GetElement(resp.Models[0].PhotoUID); el != nil {
		logTitle := el.Value.PhotoTitle
		_ = logTitle // use in audit/logging
	}

	return resp, nil
}
```

`photosByUID` keeps the submission order defined by the UI so `resp.Models`
matches the frontend expectations, while the embedded map lets us jump to a
specific `entity.Photo` instantly during album/label updates. Passing the
preloaded map into `batch.NewPhotosFormWithEntities` avoids re-querying the same
photos, which keeps `/api/v1/batch/photos/edit` fast even for large selections.

### Concurrency Helpers

For long-lived caches that multiple goroutines touch (for example, background
workers adding or removing photos while HTTP handlers read the same selection),
wrap the ordered map with `SyncMap`:

```go
cache := ordered.NewSyncMap[string, *entity.Photo]()
cache.Set(photo.PhotoUID, photo)
if _, ok := cache.Get(uid); ok {
	// safe concurrent read
}
```

`SyncMap` applies read/write locking around every operation, so callers do not
need to sprinkle additional mutex logic around shared ordered selections.
