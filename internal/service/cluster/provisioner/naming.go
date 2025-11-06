package provisioner

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base32"
	"fmt"
	"strings"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/service/cluster"
	"github.com/photoprism/photoprism/pkg/dsn"
	"github.com/photoprism/photoprism/pkg/rnd"
)

const (
	// Name prefix for generated DB objects.
	// Final pattern without slugs (UUID-based):
	//   database: cluster_d<hmac11>
	//   username: cluster_u<hmac11>
	dbSuffix   = 11
	userSuffix = 11
	// Budgets: keep user conservative for MySQL compatibility; MariaDB allows more.
	userMax = 32
	dbMax   = 64
	// prefixMax ensures usernames remain within the MySQL identifier limit.
	prefixMax = cluster.DatabaseProvisionPrefixMaxLen
)

// DatabasePrefix stores the default identifier prefix for provisioned databases and users.
// Portal deployments override this value during initialization based on configuration.
var DatabasePrefix = cluster.DefaultDatabaseProvisionPrefix

// GenerateCredentials computes deterministic database name and user for a node under the given portal
// plus a random password. Naming is stable for a given (clusterUUID, nodeUUID) pair and changes
// if the cluster UUID or node UUID changes.
func GenerateCredentials(conf *config.Config, nodeUUID, nodeName string) (dbName, dbUser, dbPass string) {
	clusterUUID := conf.ClusterUUID()

	prefix := DatabasePrefix
	if conf != nil {
		if p := conf.DatabaseProvisionPrefix(); p != "" {
			prefix = p
		}
	}
	if prefix == "" {
		prefix = cluster.DefaultDatabaseProvisionPrefix
	}
	if len(prefix) > prefixMax {
		prefix = prefix[:prefixMax]
	}

	// Compute base32 (no padding) HMAC suffixes scoped by cluster UUID and node UUID.
	sName := hmacBase32("db-name:"+clusterUUID, nodeUUID)
	sUser := hmacBase32("db-user:"+clusterUUID, nodeUUID)

	// Budgets: user ≤32, db ≤64. Suffixes are fixed length and derived from UUID.
	dbName = fmt.Sprintf("%sd%s", prefix, sName[:dbSuffix])
	dbUser = fmt.Sprintf("%su%s", prefix, sUser[:userSuffix])
	dbPass = rnd.Base62(32)

	return
}

// BuildDSN returns a DSN suitable for PhotoPrism nodes given a database driver.
// Currently, "mysql"/"mariadb" are supported; other drivers log a warning and fall back to MySQL format.
func BuildDSN(driver, host string, port int, user, pass, name string) string {
	d := strings.ToLower(driver)
	switch d {
	case dsn.DriverMySQL, dsn.DriverMariaDB:
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?%s",
			user, pass, host, port, name, dsn.Params[dsn.DriverMySQL],
		)
	default:
		log.Warnf("provisioner: unsupported driver %q, falling back to mysql DSN format", driver)
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&collation=utf8mb4_unicode_ci&parseTime=true",
			user, pass, host, port, name,
		)
	}
}

// hmacBase32 returns a lowercase Base32 (no padding) encoded HMAC-SHA256 digest
// derived from the provided key and data. It is used to generate deterministic
// suffixes for database identifiers while keeping the resulting string URL/identifier safe.
func hmacBase32(key, data string) string {
	mac := hmac.New(sha256.New, []byte(key))
	_, _ = mac.Write([]byte(data))
	sum := mac.Sum(nil)
	enc := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(sum)

	return strings.ToLower(enc)
}
