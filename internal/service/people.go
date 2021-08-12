package service

import (
	"sync"

	"github.com/photoprism/photoprism/internal/photoprism"
)

var oncePeople sync.Once

func initPeople() {
	services.People = photoprism.NewPeople(Config())
}

func People() *photoprism.People {
	oncePeople.Do(initPeople)

	return services.People
}
