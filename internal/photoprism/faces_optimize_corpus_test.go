//go:build facescorpus

package photoprism

import (
	"bufio"
	"os"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/ai/face"
	"github.com/photoprism/photoprism/internal/entity"
)

// TestMergeGroupsCorpus runs the production grouping over an export of a real library, and asserts
// the properties rather than the yield, which belongs to that library:
//
//	FACES_CORPUS=<export.tsv> go test ./internal/photoprism -tags facescorpus \
//	  -run MergeGroupsCorpus -count=1 -v
//
// FACES_CORPUS_THRESHOLDS names the exporting library's "clusterDist,clusterRadius,matchDist",
// since another model's thresholds change every verdict.
func TestMergeGroupsCorpus(t *testing.T) {
	path := os.Getenv("FACES_CORPUS")

	if path == "" {
		t.Skip("set FACES_CORPUS to a manual face export")
	}

	if s := os.Getenv("FACES_CORPUS_THRESHOLDS"); s != "" {
		defer func(d, r, m float64) { face.ClusterDist, face.ClusterRadius, face.MatchDist = d, r, m }(
			face.ClusterDist, face.ClusterRadius, face.MatchDist)

		cols := strings.Split(s, ",")
		require.Len(t, cols, 3, "expected clusterDist,clusterRadius,matchDist")

		for i, at := range []*float64{&face.ClusterDist, &face.ClusterRadius, &face.MatchDist} {
			v, err := strconv.ParseFloat(strings.TrimSpace(cols[i]), 64)
			require.NoError(t, err)
			*at = v
		}
	}

	faces := readCorpusFaces(t, path)
	require.NotEmpty(t, faces)

	groups := mergeGroups(faces)

	absorbed, pairs := 0, 0
	subjects := make(map[string]bool)

	for _, g := range groups {
		if len(g) < 2 {
			continue
		}

		absorbed += len(g) - 1
		subjects[g[0].SubjUID] = true

		for i := range g {
			assert.Equal(t, g[0].SubjUID, g[i].SubjUID, "a group must never span two subjects")

			for j := i + 1; j < len(g); j++ {
				if ok, _ := g[i].Mergeable(&g[j]); ok {
					pairs++
				}
			}
		}
	}

	t.Logf("cluster-dist %.2f: %d faces -> %d groups, %d absorbed, %d pairs across %d subjects",
		face.ClusterDist, len(faces), len(groups), absorbed, pairs, len(subjects))

	// A group has to rest on at least the links that span it, or the partition is
	// reporting a connection the criterion did not make.
	assert.GreaterOrEqual(t, pairs, absorbed)
	assert.Equal(t, len(faces)-absorbed, len(groups))

	t.Run("SameForEveryOrder", func(t *testing.T) {
		// Reversed rather than shuffled, so the cluster the export's ordering put first is
		// considered last - the arrangement an anchored predicate differs on.
		reversed := make(entity.Faces, len(faces))

		for i := range faces {
			reversed[len(faces)-1-i] = faces[i]
		}

		assert.Equal(t, groupSignature(groups), groupSignature(mergeGroups(reversed)))
	})
}

// groupSignature reduces a partition to a comparable form, so two orderings of the same clusters
// can be checked for having produced the same groups rather than the same slices.
func groupSignature(groups []entity.Faces) []string {
	result := make([]string, 0, len(groups))

	for _, g := range groups {
		ids := g.IDs()
		slices.Sort(ids)
		result = append(result, strings.Join(ids, ","))
	}

	slices.Sort(result)

	return result
}

// readCorpusFaces parses the tab-separated export: subject, id, samples, radius, collision, vector.
func readCorpusFaces(t *testing.T, path string) entity.Faces {
	t.Helper()

	f, err := os.Open(path)
	require.NoError(t, err)

	defer f.Close()

	var result entity.Faces

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	for scanner.Scan() {
		cols := strings.Split(scanner.Text(), "\t")

		if len(cols) < 6 {
			continue
		}

		samples, err := strconv.Atoi(cols[2])
		require.NoError(t, err)

		radius, err := strconv.ParseFloat(cols[3], 64)
		require.NoError(t, err)

		collision, err := strconv.ParseFloat(cols[4], 64)
		require.NoError(t, err)

		result = append(result, entity.Face{
			ID:              cols[1],
			SubjUID:         cols[0],
			Samples:         samples,
			SampleRadius:    radius,
			CollisionRadius: collision,
			EmbeddingJSON:   []byte(cols[5]),
			// One name for every row, so the criterion is decided by the distances alone.
			EmbedModel: face.ModelFaceNet,
		})
	}

	require.NoError(t, scanner.Err())

	return result
}
