package face

import (
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/txt"
)

// LicenseAcceptanceVar is the environment variable an operator sets to enable the model weights
// that are gated. One variable covers every model from the same publisher.
const LicenseAcceptanceVar = "INSIGHTFACE_ACCEPT_LICENSE"

// LicenseEligibleEditions lists the editions in which gated model weights may be enabled.
// The installer states what an operator is confirming when they set the acceptance variable.
var LicenseEligibleEditions = []string{"ce", "plus"}

// LicenseGated reports whether the model weights have to be enabled explicitly before use.
func (m *EmbeddingModel) LicenseGated() bool {
	return m != nil && m.WeightLicense() == LicenseNonFree
}

// WeightLicense returns the license of the detector's pretrained weights.
func (d *Detector) WeightLicense() string {
	if d == nil || d.ONNX == nil {
		return ""
	}

	return d.ONNX.License
}

// LicenseGated reports whether the detector weights have to be enabled explicitly before use.
func (d *Detector) LicenseGated() bool {
	return d != nil && d.WeightLicense() == LicenseNonFree
}

// LicenseAccepted reports whether the operator enabled the gated weights.
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
// nil when it may. Models whose weights are not gated are always permitted.
func LicenseRefused(name ModelName, edition string) error {
	model := FindEmbeddingModel(name)

	if !model.LicenseGated() {
		return nil
	}

	if !LicenseAccepted() {
		return fmt.Errorf("the %s weights have to be enabled explicitly, "+
			"see the model documentation and set %s=1 to use them", model.Name, LicenseAcceptanceVar)
	}

	if !LicenseEligibleEdition(edition) {
		return fmt.Errorf("the %s weights are not available in the %s edition",
			model.Name, clean.Log(edition))
	}

	return nil
}
