package entity

import (
	"hash/fnv"
	"sync"

	"github.com/photoprism/photoprism/internal/ai/face"
)

// Face and marker fixture vectors are generated rather than stored, because a stored vector
// belongs to exactly one model: it has that model's width and no recorded provenance, so every
// fixture cluster and marker becomes ineligible for matching as soon as a run resolves to a
// different model, and the tests that look like they exercise matching exercise the early exit.

// faceFixtureSeeds names the person each fixture cluster stands for. Independent seeds are
// near-orthogonal, so no two identities fall within an accept distance of one another.
var faceFixtureSeeds = map[string]uint64{
	"john-doe":  101,
	"unknown":   102,
	"joe-biden": 103,
	"jane-doe":  104,
	"fa-gr":     105,
	"actress-1": 106,
	"actor-1":   107,
}

// markerFixtureVector places a face marker relative to a cluster.
type markerFixtureVector struct {
	// face names the fixture cluster the marker is generated near, or is empty when the
	// marker stands for a person of its own.
	face string
	// factor is the distance from that cluster's centroid as a fraction of the distance it
	// accepts, so a marker stays inside or outside its cluster whatever the model and
	// whatever the thresholds are recalibrated to. Values above 1 sit outside.
	factor float64
	// seed picks the direction the marker is offset in, and identifies a marker that has no
	// cluster of its own.
	seed uint64
}

// markerFixtureVectors describes the geometry of every face marker fixture. The factors are
// the ratios the recorded face_dist values express against the FaceNet thresholds the stored
// vectors were generated with, so a marker keeps the position its row claims.
var markerFixtureVectors = map[string]markerFixtureVector{
	// Unassigned, and deliberately just outside what Joe Biden's cluster accepts: raising the
	// cluster radius has to be what brings it in, which is what TestFace_Match checks.
	"1000003-4": {face: "joe-biden", factor: 1.05, seed: 201},
	// The unknown cluster holds this marker alone, so it sits exactly on the centroid.
	"1000003-5": {face: "unknown", factor: 0, seed: 202},
	"1000003-6": {face: "john-doe", factor: 0.2439, seed: 203},
	"ma-ba-1":   {face: "fa-gr", factor: 0.6816, seed: 204},
	"fa-gr-1":   {face: "fa-gr", factor: 0.8180, seed: 205},
	"fa-gr-2":   {face: "fa-gr", factor: 0.8180, seed: 206},
	"fa-gr-3":   {face: "fa-gr", factor: 0.8180, seed: 207},

	"actress-a-1": {face: "actress-1", factor: 0.3958, seed: 208},
	"actress-a-2": {face: "actress-1", factor: 0.6643, seed: 209},
	"actress-a-3": {face: "actress-1", factor: 0.7516, seed: 210},

	"actor-a-1": {face: "actor-1", factor: 0.7214, seed: 211},
	"actor-a-2": {face: "actor-1", factor: 0.7028, seed: 212},
	"actor-a-3": {face: "actor-1", factor: 0.4337, seed: 213},
	"actor-a-4": {face: "actor-1", factor: 0.4337, seed: 214},

	"ms6sg6b14ahkyd24": {face: "joe-biden", factor: 0.3829, seed: 215},
}

var (
	fixtureVectorMutex sync.Mutex
	fixtureVectorModel face.ModelName
	fixtureVectorDims  int
	fixtureVectorsSet  bool
)

// GenerateFaceFixtureVectors fills the face and marker fixtures with embeddings for the configured
// model, before either is written. It regenerates when the model or its width changed, so a process
// that configures a second model does not seed the second library from the first one's vectors.
func GenerateFaceFixtureVectors() {
	model := face.EmbeddingModelName()
	dims := face.ExpectedDims()

	fixtureVectorMutex.Lock()
	defer fixtureVectorMutex.Unlock()

	if fixtureVectorsSet && fixtureVectorModel == model && fixtureVectorDims == dims {
		return
	}

	centroids := make(map[string]face.Embedding, len(faceFixtureSeeds))

	for name, seed := range faceFixtureSeeds {
		f, ok := FaceFixtures[name]

		if !ok {
			continue
		}

		embedding := face.FixtureEmbedding(seed)
		centroids[name] = embedding

		f.EmbeddingJSON = embedding.JSON()
		f.EmbedModel = model
		FaceFixtures[name] = f
	}

	for name, m := range MarkerFixtures {
		if m.MarkerType != MarkerFace {
			continue
		}

		spec, ok := markerFixtureVectors[name]

		if !ok {
			// A face marker that names no cluster stands for a person of its own, so one
			// added without an entry above is still matchable rather than silently inert.
			spec = markerFixtureVector{seed: fixtureSeed(name)}
		}

		m.SetEmbeddings(face.Embeddings{faceFixtureMarkerEmbedding(spec, centroids)}, model)
		MarkerFixtures[name] = m
	}

	fixtureVectorModel = model
	fixtureVectorDims = dims
	fixtureVectorsSet = true
}

// faceFixtureMarkerEmbedding returns the vector a marker fixture is generated with: a point at the
// intended distance from its cluster, or a person of its own when it names none. The distance follows
// that cluster, because a fixture radius may be narrower or wider than the calibrated one.
func faceFixtureMarkerEmbedding(spec markerFixtureVector, centroids map[string]face.Embedding) face.Embedding {
	centroid, found := centroids[spec.face]

	if !found {
		return face.FixtureEmbedding(spec.seed)
	}

	return face.FixtureEmbeddingAt(centroid, spec.factor*face.AcceptDist(FaceFixtures[spec.face].SampleRadius), spec.seed)
}

// fixtureSeed derives a stable seed from a fixture name, so a marker that is not placed
// relative to a cluster still gets the same vector on every run.
func fixtureSeed(name string) uint64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(name))

	return h.Sum64()
}
