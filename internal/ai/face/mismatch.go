package face

import "sync"

var (
	mismatchMu     sync.RWMutex
	mismatchReason string
)

// BlockEmbeddings records why the vectors a library holds cannot be compared with the model
// this instance is configured for, which pauses embedding, clustering and matching until a
// migration reconciles them.
//
// Filtering the incomparable rows out instead lets indexing keep writing vectors in a second
// space beside the first, which is the state that has no cheap way back.
func BlockEmbeddings(reason string) {
	mismatchMu.Lock()
	mismatchReason = reason
	mismatchMu.Unlock()
}

// UnblockEmbeddings clears the block, which a completed migration does because it has just
// rewritten every vector in the configured model's space.
func UnblockEmbeddings() {
	BlockEmbeddings("")
}

// EmbeddingsBlockedReason returns why embedding work is paused, or an empty string when it
// is not.
func EmbeddingsBlockedReason() string {
	mismatchMu.RLock()
	reason := mismatchReason
	mismatchMu.RUnlock()

	return reason
}

// EmbeddingsBlocked reports whether embedding work is paused because the library and the
// configured model disagree.
func EmbeddingsBlocked() bool {
	return EmbeddingsBlockedReason() != ""
}
