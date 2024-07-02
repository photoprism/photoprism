package get

import (
	"sync"

	"github.com/photoprism/photoprism/internal/entity/query"
)

var onceQuery sync.Once

func initQuery() {
	services.Query = query.New(Config().Db())
}

func Query() *query.Query {
	onceQuery.Do(initQuery)

	return services.Query
}
