package entity

import "github.com/photoprism/photoprism/pkg/react"

type ReactionMap map[string]Reaction

func (m ReactionMap) Get(name string) Reaction {
	if result, ok := m[name]; ok {
		return result
	}

	return Reaction{}
}

func (m ReactionMap) Pointer(name string) *Reaction {
	if result, ok := m[name]; ok {
		return &result
	}

	return &Reaction{}
}

var ReactionFixtures = ReactionMap{
	"SubjectJohnLike": Reaction{
		UID:       SubjectFixtures.Get("john-doe").SubjUID,
		UserUID:   UserFixtures.Get("alice").UserUID,
		Reaction:  react.Like.String(),
		Reacted:   1,
		ReactedAt: TimeStamp(),
	},
	"PhotoAliceLove": Reaction{
		UID:       PhotoFixtures.Get("Photo01").PhotoUID,
		UserUID:   UserFixtures.Pointer("alice").UserUID,
		Reaction:  react.Love.String(),
		Reacted:   3,
		ReactedAt: TimeStamp(),
	},
	"PhotoBobLove": Reaction{
		UID:       PhotoFixtures.Get("Photo01").PhotoUID,
		UserUID:   UserFixtures.Pointer("bob").UserUID,
		Reaction:  react.Love.String(),
		Reacted:   1,
		ReactedAt: TimeStamp(),
	},
}

// CreateReactionFixtures inserts known entities into the database for testing.
func CreateReactionFixtures() {
	for _, entity := range ReactionFixtures {
		Db().Create(&entity)
	}
}
