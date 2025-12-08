package cluster

// OptionsUpdate represents a set of configuration values that should be
// persisted to options.yml after a Portal cluster operation.
type OptionsUpdate struct {
	ClusterUUID      *string
	ClusterCIDR      *string
	NodeClientID     *string
	NodeClientSecret *string
	JWKSUrl          *string
	NodeUUID         *string
	DatabaseDriver   *string
	DatabaseDSN      *string
	DatabaseServer   *string
	DatabaseName     *string
	DatabaseUser     *string
	DatabasePassword *string
}

// IsZero reports whether no fields have been set on the update.
func (u OptionsUpdate) IsZero() bool {
	return u.ClusterUUID == nil &&
		u.ClusterCIDR == nil &&
		u.NodeClientID == nil &&
		u.NodeClientSecret == nil &&
		u.JWKSUrl == nil &&
		u.NodeUUID == nil &&
		u.DatabaseDriver == nil &&
		u.DatabaseDSN == nil &&
		u.DatabaseServer == nil &&
		u.DatabaseName == nil &&
		u.DatabaseUser == nil &&
		u.DatabasePassword == nil
}

// HasDatabaseUpdate reports whether the update changes any database-related fields.
func (u OptionsUpdate) HasDatabaseUpdate() bool {
	return u.DatabaseDriver != nil ||
		u.DatabaseDSN != nil ||
		u.DatabaseServer != nil ||
		u.DatabaseName != nil ||
		u.DatabaseUser != nil ||
		u.DatabasePassword != nil
}

// Setter helpers ----------------------------------------------------------------

// SetClusterUUID sets the cluster UUID value.
func (u *OptionsUpdate) SetClusterUUID(value string) {
	u.ClusterUUID = stringPtr(value)
}

// SetClusterCIDR sets the cluster CIDR value.
func (u *OptionsUpdate) SetClusterCIDR(value string) {
	u.ClusterCIDR = stringPtr(value)
}

// SetNodeClientID sets the node client ID.
func (u *OptionsUpdate) SetNodeClientID(value string) {
	u.NodeClientID = stringPtr(value)
}

// SetNodeClientSecret sets the node client secret.
func (u *OptionsUpdate) SetNodeClientSecret(value string) {
	u.NodeClientSecret = stringPtr(value)
}

// SetJWKSUrl sets the JWKS URL.
func (u *OptionsUpdate) SetJWKSUrl(value string) {
	u.JWKSUrl = stringPtr(value)
}

// SetNodeUUID sets the node UUID.
func (u *OptionsUpdate) SetNodeUUID(value string) {
	u.NodeUUID = stringPtr(value)
}

// SetDatabaseDriver sets the database driver name.
func (u *OptionsUpdate) SetDatabaseDriver(value string) {
	u.DatabaseDriver = stringPtr(value)
}

// SetDatabaseDSN sets the database DSN.
func (u *OptionsUpdate) SetDatabaseDSN(value string) {
	u.DatabaseDSN = stringPtr(value)
}

// SetDatabaseServer sets the database server address.
func (u *OptionsUpdate) SetDatabaseServer(value string) {
	u.DatabaseServer = stringPtr(value)
}

// SetDatabaseName sets the database name.
func (u *OptionsUpdate) SetDatabaseName(value string) {
	u.DatabaseName = stringPtr(value)
}

// SetDatabaseUser sets the database username.
func (u *OptionsUpdate) SetDatabaseUser(value string) {
	u.DatabaseUser = stringPtr(value)
}

// SetDatabasePassword sets the database password.
func (u *OptionsUpdate) SetDatabasePassword(value string) {
	u.DatabasePassword = stringPtr(value)
}

// Visit enumerates all set fields and invokes fn with the corresponding key/value pair.
func (u OptionsUpdate) Visit(fn func(string, any)) {
	if u.ClusterUUID != nil {
		fn("ClusterUUID", *u.ClusterUUID)
	}
	if u.ClusterCIDR != nil {
		fn("ClusterCIDR", *u.ClusterCIDR)
	}
	if u.NodeClientID != nil {
		fn("NodeClientID", *u.NodeClientID)
	}
	if u.NodeClientSecret != nil {
		fn("NodeClientSecret", *u.NodeClientSecret)
	}
	if u.JWKSUrl != nil {
		fn("JWKSUrl", *u.JWKSUrl)
	}
	if u.NodeUUID != nil {
		fn("NodeUUID", *u.NodeUUID)
	}
	if u.DatabaseDriver != nil {
		fn("DatabaseDriver", *u.DatabaseDriver)
	}
	if u.DatabaseDSN != nil {
		fn("DatabaseDSN", *u.DatabaseDSN)
	}
	if u.DatabaseServer != nil {
		fn("DatabaseServer", *u.DatabaseServer)
	}
	if u.DatabaseName != nil {
		fn("DatabaseName", *u.DatabaseName)
	}
	if u.DatabaseUser != nil {
		fn("DatabaseUser", *u.DatabaseUser)
	}
	if u.DatabasePassword != nil {
		fn("DatabasePassword", *u.DatabasePassword)
	}
}

func stringPtr(value string) *string {
	v := value
	return &v
}
