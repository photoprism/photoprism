//go:build integration

package photoprism

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"os"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/pkg/rnd"
)

const matchLibraryLandmarks = `[{"name":"nose","x":0.0027777778,"y":0.011090573,"w":0.0027777778,"h":0.0036968577},{"name":"mouth_l","y":0.014787431,"w":0.0027777778,"h":0.0036968577},{"name":"mouth_r","x":0.0055555557,"y":0.018484289,"w":0.0027777778,"h":0.0036968577},{"name":"eye_l","w":0.0027777778,"h":0.0036968577},{"name":"eye_r","x":0.008333334,"y":0.0018484289,"w":0.0027777778,"h":0.0036968577}]`

// seedMatchLibrary creates a reproducible library with clustered and faceless markers.
func seedMatchLibrary(t *testing.T, clusters, perCluster uint64) (faces entity.Faces) {
	t.Helper()
	db := entity.Db()
	model := face.EmbeddingModelName()
	dims := face.ExpectedDims()
	t.Logf("model=%s dims=%d driver=%s", model, dims, Config().DatabaseDriver())
	require.NoError(t, db.Exec("DELETE FROM markers").Error)
	require.NoError(t, db.Exec("DELETE FROM faces").Error)
	r := rand.New(rand.NewPCG(7, 11)) //nolint:gosec
	past := time.Now().UTC().Add(-2 * time.Hour).Truncate(time.Second)
	var rows []string
	var args []any
	flush := func() {
		if len(rows) == 0 {
			return
		}
		sql := "INSERT INTO markers (marker_uid, file_uid, marker_type, marker_src, marker_name, marker_review, marker_invalid, subj_uid, subj_src, face_id, face_dist, embed_model, detect_model, embeddings_json, landmarks_json, x, y, w, h, size, thumb_size, embed_detail, score, thumb, matched_at, created_at, updated_at) VALUES " + strings.Join(rows, ",")
		if err := db.Exec(sql, args...).Error; err != nil {
			t.Fatal(err)
		}
		rows, args = nil, nil
	}
	for c := uint64(0); c < clusters; c++ {
		center := face.FixtureEmbedding(c + 1)
		if c == 1 {
			center = face.FixtureEmbeddingAt(faces[0].Embedding(), 0.6, 900001)
		}
		members := make(face.Embeddings, perCluster)
		for i := uint64(0); i < perCluster; i++ {
			distance := max(0.45, min(0.95, 0.77+0.095*r.NormFloat64()))
			if c == 1 {
				distance = 0.05
			}
			members[i] = face.FixtureEmbeddingAt(center, distance, c*perCluster+i+1000)
		}
		subj := ""
		if c == 1 {
			subj = faces[0].SubjUID
		} else if (c+1)%5 != 0 {
			subject := entity.NewSubject(fmt.Sprintf("Equivalence Person %d", c), entity.SubjPerson, entity.SrcAuto)
			require.NoError(t, db.Create(subject).Error)
			subj = subject.SubjUID
		}
		f := entity.NewFace(subj, entity.SrcAuto, members, model)
		f.MatchedAt = &past
		if err := db.Create(f).Error; err != nil {
			t.Fatal(err)
		}
		faces = append(faces, *f)
		for i := uint64(0); i < perCluster; i++ {
			ej, err := json.Marshal(face.Embeddings{members[i]})
			require.NoError(t, err)
			dist := face.Embeddings{members[i]}.Dist(f.Embedding())
			faceID, msubj, msrc := f.ID, subj, ""
			if subj != "" {
				msrc = "auto"
			}
			if i < 63 {
				// Faceless vectors are independent of every cluster center.
				faceID, dist, msubj, msrc = "", -1, "", ""
				ej, err = json.Marshal(face.Embeddings{face.FixtureEmbedding(100000 + c*perCluster + i)})
				require.NoError(t, err)
			}
			rows = append(rows, "(?, ?, 'face', 'image', '', 0, 0, ?, ?, ?, ?, ?, 'scrfd', ?, ?, 0.1, 0.1, 0.2, 0.2, 200, 720, 100, 50, ?, ?, ?, ?)")
			args = append(args, rnd.GenerateUID('m'), rnd.GenerateUID('f'), msubj, msrc, faceID, dist, string(model), ej, []byte(matchLibraryLandmarks), rnd.Base36(40), past, past, past)
			if len(rows) >= 200 {
				flush()
			}
		}
	}
	flush()
	return faces
}

// assertMatchLibraryStamps checks that every eligible marker received this run's stamp.
func assertMatchLibraryStamps(t *testing.T, start time.Time) {
	t.Helper()
	var count int
	require.NoError(t, entity.Db().Model(&entity.Marker{}).
		Where("matched_at IS NULL OR matched_at < ?", start.UTC().Truncate(time.Second)).Count(&count).Error)
	require.Zero(t, count, "every eligible marker must receive a fresh match stamp")
}

// matchLibraryAssignment records the persisted cluster and subject link of one marker.
type matchLibraryAssignment struct {
	MarkerUID string
	FaceID    string
	SubjUID   string
	SubjSrc   string
}

// matchLibraryAssignments reads every marker in stable order.
func matchLibraryAssignments(t *testing.T) []matchLibraryAssignment {
	t.Helper()
	var rows []matchLibraryAssignment
	require.NoError(t, entity.Db().Table("markers").
		Select("marker_uid, face_id, subj_uid, subj_src").Order("marker_uid").Scan(&rows).Error)
	require.Len(t, rows, 60000)
	return rows
}

// copyMatchLibrary preserves the exact input rows for independent matching runs.
func copyMatchLibrary(t *testing.T) func() {
	t.Helper()
	db := entity.Db()
	type schedule struct {
		MarkerUID string
		FaceDist  float64
		MatchedAt *time.Time
	}
	var input []schedule
	require.NoError(t, db.Table("markers").Select("marker_uid, face_dist, matched_at").
		Order("marker_uid").Scan(&input).Error)
	for _, table := range []string{"markers", "faces"} {
		backup := "test_match_library_" + table
		require.NoError(t, db.Exec("CREATE TABLE "+backup+" AS SELECT * FROM "+table).Error)
		t.Cleanup(func() { require.NoError(t, db.Exec("DROP TABLE "+backup).Error) })
	}
	return func() {
		for _, table := range []string{"markers", "faces"} {
			require.NoError(t, db.Exec("DELETE FROM "+table).Error)
			require.NoError(t, db.Exec("INSERT INTO "+table+" SELECT * FROM test_match_library_"+table).Error)
		}
		var restored []schedule
		require.NoError(t, db.Table("markers").Select("marker_uid, face_dist, matched_at").
			Order("marker_uid").Scan(&restored).Error)
		require.Equal(t, input, restored, "restored markers must keep their input distances and match stamps")
		entity.FlushCaches()
	}
}

// TestFaces_MatchLibraryEquivalence compares normal and forced runs on identical correction inputs.
func TestFaces_MatchLibraryEquivalence(t *testing.T) {
	if os.Getenv("PHOTOPRISM_TEST_FACE_LIBRARY") != "1" {
		t.Skip("set PHOTOPRISM_TEST_FACE_LIBRARY=1 for the 60000-marker fixture")
	}

	level := log.GetLevel()
	log.SetLevel(logrus.InfoLevel)
	t.Cleanup(func() { log.SetLevel(level) })
	useTestEmbedder(t, face.ModelSFace)
	updateFaces := entity.UpdateFaces.Load()
	t.Cleanup(func() { entity.UpdateFaces.Store(updateFaces) })
	t.Cleanup(entity.ResetTestFixtures)

	faces := seedMatchLibrary(t, 600, 100)
	before := matchLibraryAssignments(t)
	idx := buildFaceIndex(faces)
	cursor, faceless := "", 0
	for {
		var markers entity.Markers
		require.NoError(t, entity.Db().Where("marker_uid > ? AND face_id = ?", cursor, "").
			Order("marker_uid").Limit(500).Find(&markers).Error)
		if len(markers) == 0 {
			break
		}
		for _, marker := range markers {
			candidate, _, ambiguous := selectBestFace(marker.Embeddings(), idx, false)
			require.Nil(t, candidate, "faceless fixture marker must be outside every cluster")
			require.False(t, ambiguous)
			faceless++
		}
		cursor = markers[len(markers)-1].MarkerUID
	}
	require.Equal(t, 37800, faceless)

	narrowed := &faces[0]
	narrowed.CollisionRadius = 0.75
	narrowed.Collisions = 1
	narrowed.MatchedAt = nil
	require.NoError(t, entity.Db().Model(narrowed).UpdateColumns(entity.Values{
		"collision_radius": narrowed.CollisionRadius, "collisions": 1, "matched_at": nil,
	}).Error)
	var held entity.Marker
	require.NoError(t, entity.Db().Where("face_id = ?", faces[1].ID).Order("marker_uid").First(&held).Error)
	accepts, competingDistance := narrowed.Match(held.Embeddings(), held.EmbedModel)
	require.True(t, accepts, "narrowed cluster must accept the held sentinel")
	require.Less(t, held.FaceDist+face.MatchMargin, competingDistance)
	require.NotEmpty(t, held.SubjUID)
	released, err := narrowed.ReviseMatches()
	require.NoError(t, err)
	require.Greater(t, len(released), 1)
	// Vector order keeps the new centroid independent of generated marker UIDs.
	slices.SortFunc(released, func(a, b entity.Marker) int {
		return bytes.Compare(a.EmbeddingsJSON, b.EmbeddingsJSON)
	})
	t.Logf("fixture: markers=60000 clusters=600 named=480 faceless=%d released=%d", faceless, len(released))

	for _, withNew := range []bool{false, true} {
		name := "PureNarrowing"
		if withNew {
			name = "NewCluster"
		}
		t.Run(name, func(t *testing.T) {
			var added *entity.Face
			if withNew {
				var embeddings face.Embeddings
				for i := range released {
					if i%2 == 0 {
						embeddings = append(embeddings, released[i].Embeddings()...)
					}
				}
				added = entity.NewFace(narrowed.SubjUID, entity.SrcAuto, embeddings, face.ModelSFace)
				require.NoError(t, entity.Db().Create(added).Error)
			}
			input := matchLibraryAssignments(t)
			var inputFaces entity.Faces
			require.NoError(t, entity.Db().Order("id").Find(&inputFaces).Error)
			restore := copyMatchLibrary(t)
			defer restore()
			start := time.Now()
			result, err := NewFaces(Config()).Match(FacesOptions{})
			normalTime := time.Since(start)
			require.NoError(t, err)
			normal := matchLibraryAssignments(t)

			byUID := make(map[string]matchLibraryAssignment, len(normal))
			for _, row := range normal {
				byUID[row.MarkerUID] = row
			}
			for _, row := range before {
				if row.FaceID != "" && row.FaceID != narrowed.ID {
					require.Equal(t, row, byUID[row.MarkerUID], "held marker must keep its cluster and subject")
				}
			}
			joined := 0
			for _, marker := range released {
				row := byUID[marker.MarkerUID]
				require.NotEqual(t, narrowed.ID, row.FaceID)
				if added != nil && row.FaceID == added.ID {
					require.Equal(t, added.SubjUID, row.SubjUID)
					joined++
				}
			}
			if added != nil {
				require.Positive(t, joined, "released markers must join the new cluster in this run")
			}

			assertMatchLibraryStamps(t, start)
			restore()
			require.Equal(t, input, matchLibraryAssignments(t), "forced run must start from the same assignments")
			var restoredFaces entity.Faces
			require.NoError(t, entity.Db().Order("id").Find(&restoredFaces).Error)
			require.Equal(t, inputFaces, restoredFaces, "forced run must start from the same cluster state")
			start = time.Now()
			forcedResult, err := NewFaces(Config()).Match(FacesOptions{Force: true})
			forcedTime := time.Since(start)
			require.NoError(t, err)
			assertMatchLibraryStamps(t, start)
			forced := matchLibraryAssignments(t)
			differences := 0
			for i := range normal {
				if normal[i] != forced[i] {
					differences++
					if differences <= 5 {
						t.Errorf("assignment differs: normal=%+v forced=%+v", normal[i], forced[i])
					}
				}
			}
			require.Zero(t, differences)
			t.Logf("equivalence: normal=%s forced=%s markers=%d differences=%d joined=%d normal_result=%+v forced_result=%+v",
				normalTime, forcedTime, len(normal), differences, joined, result, forcedResult)
		})
	}
}
