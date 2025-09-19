package provisioner

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base32"
	"fmt"
	"strings"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/pkg/clean"
	"github.com/photoprism/photoprism/pkg/rnd"
)

const (
	dbPrefix   = "pp_"
	userPrefix = "pp_"
	dbSuffix   = 8
	userSuffix = 6
	userMax    = 32
	dbMax      = 64
)

// GenerateCreds computes deterministic database name and user for a node under the given portal
// plus a random password. Naming is stable for a given (clusterUUID, nodeName) pair and changes
// if the cluster UUID changes. The returned password is random and independent.
func GenerateCreds(conf *config.Config, nodeName string) (dbName, dbUser, dbPass string) {
	clusterUUID := conf.ClusterUUID()
	slug := clean.TypeLowerDash(nodeName)

	// Compute base32 (no padding) HMAC suffixes scoped by cluster UUID.
	sName := hmacBase32("db-name:"+clusterUUID, slug)
	sUser := hmacBase32("db-user:"+clusterUUID, slug)

	// Budgets: user ≤32, db ≤64
	// Patterns: pp_<slug>_<suffix>
	// Compute max slug lengths to honor budgets.
	userSlugMax := userMax - len(userPrefix) - 1 - userSuffix // 32 - 3 - 1 - 6 = 22
	dbSlugMax := dbMax - len(dbPrefix) - 1 - dbSuffix         // 64 - 3 - 1 - 8 = 52

	slugUser := trimRunes(slug, userSlugMax)
	slugDb := trimRunes(slug, dbSlugMax)

	dbName = fmt.Sprintf("%s%s_%s", dbPrefix, slugDb, sName[:dbSuffix])
	dbUser = fmt.Sprintf("%s%s_%s", userPrefix, slugUser, sUser[:userSuffix])
	dbPass = rnd.Base62(32)
	return
}

// BuildDSN returns a MySQL/MariaDB DSN suitable for PhotoPrism nodes.
func BuildDSN(host string, port int, user, pass, name string) string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&collation=utf8mb4_unicode_ci&parseTime=true",
		user, pass, host, port, name,
	)
}

func hmacBase32(key, data string) string {
	mac := hmac.New(sha256.New, []byte(key))
	_, _ = mac.Write([]byte(data))
	sum := mac.Sum(nil)
	enc := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(sum)

	return strings.ToLower(enc)
}

func trimRunes(s string, max int) string {
	if max <= 0 || len(s) <= max {
		return s
	}

	// Trim by runes to avoid mid-rune cut, though s should be ASCII by cleaning.
	r := []rune(s)
	if len(r) <= max {
		return s
	}

	return string(r[:max])
}
