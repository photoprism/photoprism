package face

import "sync"

var (
	mismatchMu     sync.RWMutex
	mismatchReason string
)

// BlockEmbeddings records why a library's vectors cannot be compared with the configured model,
// which pauses embedding, clustering and matching until a migration reconciles them.
//
// Filtering the incomparable rows out instead lets indexing write a second vector space beside
// the first, which is the state with no cheap way back.
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
