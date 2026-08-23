package face

import (
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/txt"
)

// LicenseAcceptanceVar is the environment variable through which an operator accepts the terms
// that apply to license-gated model weights. One vendor, one variable: the same terms cover the
// detector and the embedding models it publishes.
const LicenseAcceptanceVar = "INSIGHTFACE_ACCEPT_LICENSE"

// LicenseEligibleEditions lists the editions in which license-gated weights may be used.
//
// The builds a personal user runs are the Community Edition and Plus, and the organizational
// editions are outside the terms as we read them. The edition proves nothing about the use
// itself, which is why the notice an operator accepts carries that half.
var LicenseEligibleEditions = []string{"ce", "plus"}

// LicenseGated reports whether the model weights may only be used after their vendor's terms
// have been accepted.
func (m *EmbeddingModel) LicenseGated() bool {
	return m != nil && m.WeightLicense() == LicenseResearchOnly
}

// LicenseAccepted reports whether the operator accepted the terms of the gated weights.
func LicenseAccepted() bool {
	return txt.Bool(os.Getenv(LicenseAcceptanceVar))
}

// LicenseEligibleEdition reports whether gated weights may be used in the specified edition.
// An edition that is not listed is refused, so a build this predicate does not know about
// cannot enable them by default.
func LicenseEligibleEdition(edition string) bool {
	return slices.Contains(LicenseEligibleEditions, strings.ToLower(strings.TrimSpace(edition)))
}

// LicenseRefused returns why the specified model may not be used in the specified edition, or
// nil when it may. Models whose weights carry no gate are always permitted.
func LicenseRefused(name ModelName, edition string) error {
	model := FindEmbeddingModel(name)

	if !model.LicenseGated() {
		return nil
	}

	if !LicenseAccepted() {
		return fmt.Errorf("the %s weights are published for non-commercial use only, "+
			"set %s=1 to confirm that your use is covered by their terms", model.Name, LicenseAcceptanceVar)
	}

	if !LicenseEligibleEdition(edition) {
		return fmt.Errorf("the %s weights are published for non-commercial use only, "+
			"which the %s edition is outside of", model.Name, clean.Log(edition))
	}

	return nil
}
