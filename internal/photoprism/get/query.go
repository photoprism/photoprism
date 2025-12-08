package get

import (
	"sync"

	"github.com/photoprism/photoprism/internal/entity/query"
)

var onceQuery sync.Once

func initQuery() {
	services.Query = query.New(Config().Db())
}

// Query returns the singleton query helper instance.
func Query() *query.Query {
	onceQuery.Do(initQuery)

	return services.Query
}
