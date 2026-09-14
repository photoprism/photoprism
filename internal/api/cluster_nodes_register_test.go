package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/gjson"

	"github.com/photoprism/photoprism/internal/config"
	"github.com/photoprism/photoprism/internal/entity"
	"github.com/photoprism/photoprism/internal/service/cluster"
	"github.com/photoprism/photoprism/internal/service/cluster/provisioner"
	reg "github.com/photoprism/photoprism/internal/service/cluster/registry"
	"github.com/photoprism/photoprism/pkg/authn"
	"github.com/photoprism/photoprism/pkg/fs"
	"github.com/photoprism/photoprism/pkg/http/header"
	"github.com/photoprism/photoprism/pkg/rnd"
)

func TestClusterNodesRegister(t *testing.T) {
	t.Run("FeatureDisabled", func(t *testing.T) {
		app, router, conf := NewApiTest()
		conf.Options().NodeRole = cluster.RoleInstance
		ClusterNodesRegister(router)

		r := PerformRequestWithBody(app, http.MethodPost, "/api/v1/cluster/nodes/register", `{"NodeName":"pp-node-01"}`)
		assert.Equal(t, http.StatusForbidden, r.Code)
	})
	t.Run("ExistingNodeMutationRequiresOAuthToken", func(t *testing.T) {
		app, router, conf := NewApiTest()
		enablePortalAPIs(t, conf)
		conf.Options().JoinToken = cluster.ExampleJoinToken
		ClusterNodesRegister(router)

		regy, err := reg.NewClientRegistryWithConfig(conf)
		assert.NoError(t, err)
		n := &reg.Node{Node: cluster.Node{UUID: rnd.UUIDv7(), Name: "pp-auth", Role: cluster.RoleInstance}}
		assert.NoError(t, regy.Put(n))
		nr, err := regy.RotateSecret(n.UUID)
		assert.NoError(t, err)

		// Join tokens must not mutate existing registrations.
		body := `{"NodeName":"pp-auth","Labels":{"env":"prod"}}`
		r := AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/cluster/nodes/register", body, cluster.ExampleJoinToken)
		assert.Equal(t, http.StatusConflict, r.Code)
		assert.Contains(t, r.Body.String(), "already registered")

		// Existing-node mutations with a valid node access token must succeed.
		token := oauthNodeAccessToken(t, app, router, conf, nr.ClientID, nr.ClientSecret)
		r = AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/cluster/nodes/register", body, token)
		assert.Equal(t, http.StatusOK, r.Code)
		cleanupRegisterProvisioning(t, conf, r)

		r = AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/cluster/nodes/register", `{"NodeName":"pp-auth","RotateSecret":true}`, token)
		assert.Equal(t, http.StatusOK, r.Code)
		cleanupRegisterProvisioning(t, conf, r)
	})
	t.Run("MissingToken", func(t *testing.T) {
		app, router, conf := NewApiTest()
		enablePortalAPIs(t, conf)
		ClusterNodesRegister(router)

		r := PerformRequestWithBody(app, http.MethodPost, "/api/v1/cluster/nodes/register", `{"NodeName":"pp-node-01"}`)
		assert.Equal(t, http.StatusUnauthorized, r.Code)
	})
	t.Run("RequestTooLarge", func(t *testing.T) {
		app, router, conf := NewApiTest()
		enablePortalAPIs(t, conf)
		conf.Options().JoinToken = cluster.ExampleJoinToken
		ClusterNodesRegister(router)

		body := `{"NodeName":"pp-node-big","Labels":{"env":"` + strings.Repeat("a", 300*1024) + `"}}`
		r := AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/cluster/nodes/register", body, cluster.ExampleJoinToken)
		assert.Equal(t, http.StatusRequestEntityTooLarge, r.Code)
	})
	t.Run("ClusterCIDRBlocksClientIPOutsideRange", func(t *testing.T) {
		app, router, conf := NewApiTest()
		enablePortalAPIs(t, conf)
		conf.Options().JoinToken = cluster.ExampleJoinToken
		prevClusterCIDR := conf.Options().ClusterCIDR
		t.Cleanup(func() {
			conf.Options().ClusterCIDR = prevClusterCIDR
		})
		conf.Options().ClusterCIDR = "192.0.2.0/24"
		ClusterNodesRegister(router)

		r := AuthenticatedRequestWithBodyAndIP(app, http.MethodPost, "/api/v1/cluster/nodes/register", `{"NodeName":"pp-node-cidr-blocked"}`, cluster.ExampleJoinToken, "198.51.100.9")
		assert.Equal(t, http.StatusUnauthorized, r.Code)
	})
	t.Run("ClusterCIDRAllowsClientIPInRange", func(t *testing.T) {
		app, router, conf := NewApiTest()
		enablePortalAPIs(t, conf)
		conf.Options().JoinToken = cluster.ExampleJoinToken
		prevClusterCIDR := conf.Options().ClusterCIDR
		t.Cleanup(func() {
			conf.Options().ClusterCIDR = prevClusterCIDR
		})
		conf.Options().ClusterCIDR = "192.0.2.0/24"
		ClusterNodesRegister(router)

		r := AuthenticatedRequestWithBodyAndIP(app, http.MethodPost, "/api/v1/cluster/nodes/register", `{"NodeName":"pp-node-cidr-allowed"}`, cluster.ExampleJoinToken, "192.0.2.42")
		assert.Equal(t, http.StatusCreated, r.Code)
		cleanupRegisterProvisioning(t, conf, r)
	})
	t.Run("ForbiddenFromCDN", func(t *testing.T) {
		app, router, conf := NewApiTest()
		enablePortalAPIs(t, conf)
		conf.Options().JoinToken = cluster.ExampleJoinToken
		ClusterNodesRegister(router)

		req, _ := http.NewRequest(http.MethodPost, "/api/v1/cluster/nodes/register", nil)
		req.Header.Set(header.CdnHost, "edge.example")
		req.Header.Set(header.Auth, header.AuthBearer+" "+cluster.ExampleJoinToken)
		req.Header.Set(header.Accept, "application/json")

		w := httptest.NewRecorder()
		app.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
	t.Run("CreateNodeWithoutRotateSkipsProvisioner", func(t *testing.T) {
		app, router, conf := NewApiTest()
		enablePortalAPIs(t, conf)
		conf.Options().JoinToken = cluster.ExampleJoinToken
		ClusterNodesRegister(router)

		// Provisioner is independent of the main DB; with MariaDB admin DSN configured
		// it should successfully provision and return 201.
		r := AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/cluster/nodes/register", `{"NodeName":"pp-node-01"}`, cluster.ExampleJoinToken)
		assert.Equal(t, http.StatusCreated, r.Code)
		body := r.Body.String()
		assert.Contains(t, body, "\"Database\"")
		assert.Contains(t, body, "\"Secrets\"")
		assert.Contains(t, body, "\"ClientSecret\"")
		assert.Equal(t, "", gjson.Get(body, "Database.Name").String())
		assert.False(t, gjson.Get(body, "AlreadyProvisioned").Bool())
		cleanupRegisterProvisioning(t, conf, r)
	})
	t.Run("CreateNodeRotateDatabaseProvisioned", func(t *testing.T) {
		app, router, conf := NewApiTest()
		enablePortalAPIs(t, conf)
		conf.Options().JoinToken = cluster.ExampleJoinToken
		ClusterNodesRegister(router)

		r := AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/cluster/nodes/register", `{"NodeName":"pp-node-rotate","RotateDatabase":true}`, cluster.ExampleJoinToken)
		assert.Equal(t, http.StatusCreated, r.Code)
		body := r.Body.String()
		assert.NotEqual(t, "", gjson.Get(body, "Database.Name").String())
		assert.NotEqual(t, "", gjson.Get(body, "Database.Password").String())
		assert.True(t, gjson.Get(body, "AlreadyProvisioned").Bool())
		cleanupRegisterProvisioning(t, conf, r)
	})
	t.Run("UUIDChangeWithJoinTokenReturnsConflict", func(t *testing.T) {
		app, router, conf := NewApiTest()
		enablePortalAPIs(t, conf)
		conf.Options().JoinToken = cluster.ExampleJoinToken
		ClusterNodesRegister(router)

		regy, err := reg.NewClientRegistryWithConfig(conf)
		assert.NoError(t, err)

		// Pre-create node with a UUID
		n := &reg.Node{Node: cluster.Node{UUID: rnd.UUIDv7(), Name: "pp-lock", Role: cluster.RoleInstance}}
		assert.NoError(t, regy.Put(n))

		// Attempt to change UUID via join token must fail with conflict.
		newUUID := rnd.UUIDv7()
		r := AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/cluster/nodes/register", `{"NodeName":"pp-lock","NodeUUID":"`+newUUID+`"}`, cluster.ExampleJoinToken)
		assert.Equal(t, http.StatusConflict, r.Code)
	})
	t.Run("JoinWithUnclaimedNodeUUID", func(t *testing.T) {
		app, router, conf := NewApiTest()
		enablePortalAPIs(t, conf)
		conf.Options().JoinToken = cluster.ExampleJoinToken
		ClusterNodesRegister(router)

		// Config.NodeUUID generates and persists one when unset, so every first join pins a UUID.
		pinned := rnd.UUIDv7()
		body := `{"NodeName":"pp-uuid-pinned","NodeRole":"instance","NodeUUID":"` + pinned + `"}`
		r := AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/cluster/nodes/register", body, cluster.ExampleJoinToken)
		assert.Equal(t, http.StatusCreated, r.Code)
		cleanupRegisterProvisioning(t, conf, r)
		assert.Equal(t, pinned, gjson.Get(r.Body.String(), "Node.UUID").String())

		regy, err := reg.NewClientRegistryWithConfig(conf)
		assert.NoError(t, err)
		got, err := regy.FindByNodeUUID(pinned)
		assert.NoError(t, err)
		if assert.NotNil(t, got) {
			assert.Equal(t, "pp-uuid-pinned", got.Name)
		}
	})
	t.Run("JoinWithMalformedNodeUUID", func(t *testing.T) {
		app, router, conf := NewApiTest()
		enablePortalAPIs(t, conf)
		conf.Options().JoinToken = cluster.ExampleJoinToken
		ClusterNodesRegister(router)

		// Not a canonical UUID: 36 hex characters with no separators.
		body := `{"NodeName":"pp-uuid-malformed","NodeUUID":"111111111111111111111111111111111111"}`
		r := AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/cluster/nodes/register", body, cluster.ExampleJoinToken)
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
	t.Run("JoinWithDisallowedNodeRole", func(t *testing.T) {
		app, router, conf := NewApiTest()
		enablePortalAPIs(t, conf)
		conf.Options().JoinToken = cluster.ExampleJoinToken
		ClusterNodesRegister(router)

		// A join creates an OAuth client, so it may only carry a node role.
		body := `{"NodeName":"pp-role-admin","NodeRole":"admin"}`
		r := AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/cluster/nodes/register", body, cluster.ExampleJoinToken)
		assert.Equal(t, http.StatusBadRequest, r.Code)

		regy, err := reg.NewClientRegistryWithConfig(conf)
		assert.NoError(t, err)
		_, err = regy.FindByName("pp-role-admin")
		assert.Error(t, err)
	})
	t.Run("JoinDefaultsToInstanceRole", func(t *testing.T) {
		app, router, conf := NewApiTest()
		enablePortalAPIs(t, conf)
		conf.Options().JoinToken = cluster.ExampleJoinToken
		ClusterNodesRegister(router)

		r := AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/cluster/nodes/register", `{"NodeName":"pp-role-default"}`, cluster.ExampleJoinToken)
		assert.Equal(t, http.StatusCreated, r.Code)
		cleanupRegisterProvisioning(t, conf, r)
		assert.Equal(t, cluster.RoleInstance, gjson.Get(r.Body.String(), "Node.Role").String())
	})
	t.Run("NonNodeClientRoleDenied", func(t *testing.T) {
		app, router, conf := NewApiTest()
		enablePortalAPIs(t, conf)
		conf.Options().JoinToken = cluster.ExampleJoinToken
		ClusterNodesRegister(router)

		// An ordinary OAuth client shares the name space with nodes and is not eligible.
		regy, err := reg.NewClientRegistryWithConfig(conf)
		assert.NoError(t, err)
		n := &reg.Node{Node: cluster.Node{UUID: rnd.UUIDv7(), Name: "pp-plain-client", Role: "client"}}
		assert.NoError(t, regy.Create(n))
		nr, err := regy.RotateSecret(n.UUID)
		assert.NoError(t, err)

		token := oauthNodeAccessTokenWithScope(t, app, router, conf, nr.ClientID, nr.ClientSecret, "cluster")
		r := AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/cluster/nodes/register", `{"NodeName":"pp-plain-client"}`, token)
		assert.Equal(t, http.StatusForbidden, r.Code)
	})
	t.Run("NodeUUIDClaimedByAnotherNode", func(t *testing.T) {
		app, router, conf := NewApiTest()
		enablePortalAPIs(t, conf)
		conf.Options().JoinToken = cluster.ExampleJoinToken
		ClusterNodesRegister(router)

		regy, err := reg.NewClientRegistryWithConfig(conf)
		assert.NoError(t, err)

		other := &reg.Node{Node: cluster.Node{UUID: rnd.UUIDv7(), Name: "pp-uuid-other", Role: cluster.RoleInstance}}
		assert.NoError(t, regy.Put(other))

		own := &reg.Node{Node: cluster.Node{UUID: rnd.UUIDv7(), Name: "pp-uuid-own", Role: cluster.RoleInstance}}
		assert.NoError(t, regy.Put(own))
		nr, err := regy.RotateSecret(own.UUID)
		assert.NoError(t, err)

		// A node may not name a UUID that belongs to another registration.
		token := oauthNodeAccessToken(t, app, router, conf, nr.ClientID, nr.ClientSecret)
		body := `{"NodeName":"pp-uuid-own","NodeUUID":"` + other.UUID + `","RotateSecret":true}`
		r := AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/cluster/nodes/register", body, token)
		assert.Equal(t, http.StatusForbidden, r.Code)
		assert.Empty(t, gjson.Get(r.Body.String(), "Secrets.ClientSecret").String())

		got, err := regy.FindByNodeUUID(other.UUID)
		assert.NoError(t, err)
		if assert.NotNil(t, got) {
			assert.Equal(t, "pp-uuid-other", got.Name)
			assert.Equal(t, other.ClientID, got.ClientID)
		}
	})
	t.Run("JoinWithClaimedNodeUUID", func(t *testing.T) {
		app, router, conf := NewApiTest()
		enablePortalAPIs(t, conf)
		conf.Options().JoinToken = cluster.ExampleJoinToken
		ClusterNodesRegister(router)

		regy, err := reg.NewClientRegistryWithConfig(conf)
		assert.NoError(t, err)
		n := &reg.Node{Node: cluster.Node{UUID: rnd.UUIDv7(), Name: "pp-uuid-joined", Role: cluster.RoleInstance}}
		assert.NoError(t, regy.Put(n))

		// A join token admits a new node, so a registered UUID is a conflict.
		body := `{"NodeName":"pp-uuid-unused","NodeUUID":"` + n.UUID + `"}`
		r := AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/cluster/nodes/register", body, cluster.ExampleJoinToken)
		assert.Equal(t, http.StatusConflict, r.Code)
		assert.Contains(t, r.Body.String(), "already registered")

		got, err := regy.FindByNodeUUID(n.UUID)
		assert.NoError(t, err)
		if assert.NotNil(t, got) {
			assert.Equal(t, "pp-uuid-joined", got.Name)
			assert.Equal(t, n.ClientID, got.ClientID)
		}
	})
	t.Run("OwnNodeUUIDAccepted", func(t *testing.T) {
		app, router, conf := NewApiTest()
		enablePortalAPIs(t, conf)
		conf.Options().JoinToken = cluster.ExampleJoinToken
		ClusterNodesRegister(router)

		regy, err := reg.NewClientRegistryWithConfig(conf)
		assert.NoError(t, err)
		n := &reg.Node{Node: cluster.Node{UUID: rnd.UUIDv7(), Name: "pp-uuid-self", Role: cluster.RoleInstance}}
		assert.NoError(t, regy.Put(n))
		nr, err := regy.RotateSecret(n.UUID)
		assert.NoError(t, err)

		// This is what a node reports on every start.
		token := oauthNodeAccessToken(t, app, router, conf, nr.ClientID, nr.ClientSecret)
		body := `{"NodeName":"pp-uuid-self","NodeUUID":"` + n.UUID + `"}`
		r := AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/cluster/nodes/register", body, token)
		assert.Equal(t, http.StatusOK, r.Code)
		cleanupRegisterProvisioning(t, conf, r)
		assert.Equal(t, n.UUID, gjson.Get(r.Body.String(), "Node.UUID").String())
		assert.Equal(t, nr.ClientID, gjson.Get(r.Body.String(), "Node.ClientID").String())
	})
	t.Run("UnclaimedNodeUUIDAccepted", func(t *testing.T) {
		app, router, conf := NewApiTest()
		enablePortalAPIs(t, conf)
		conf.Options().JoinToken = cluster.ExampleJoinToken
		ClusterNodesRegister(router)

		regy, err := reg.NewClientRegistryWithConfig(conf)
		assert.NoError(t, err)
		n := &reg.Node{Node: cluster.Node{UUID: rnd.UUIDv7(), Name: "pp-uuid-repin", Role: cluster.RoleInstance}}
		assert.NoError(t, regy.Put(n))
		nr, err := regy.RotateSecret(n.UUID)
		assert.NoError(t, err)

		// A node may move its own registration to a UUID no other node holds.
		token := oauthNodeAccessToken(t, app, router, conf, nr.ClientID, nr.ClientSecret)
		newUUID := rnd.UUIDv7()
		body := `{"NodeName":"pp-uuid-repin","NodeUUID":"` + newUUID + `"}`
		r := AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/cluster/nodes/register", body, token)
		assert.Equal(t, http.StatusOK, r.Code)
		cleanupRegisterProvisioning(t, conf, r)
		assert.Equal(t, newUUID, gjson.Get(r.Body.String(), "Node.UUID").String())

		got, err := regy.FindByNodeUUID(newUUID)
		assert.NoError(t, err)
		if assert.NotNil(t, got) {
			assert.Equal(t, n.ClientID, got.ClientID)
		}
	})
	t.Run("AdvertiseUrlHttpAllowed", func(t *testing.T) {
		app, router, conf := NewApiTest()
		enablePortalAPIs(t, conf)
		conf.Options().JoinToken = cluster.ExampleJoinToken
		ClusterNodesRegister(router)

		// http scheme is allowed for cluster-internal traffic, even on public hostnames.
		r := AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/cluster/nodes/register", `{"NodeName":"pp-node-03","AdvertiseUrl":"http://example.com"}`, cluster.ExampleJoinToken)
		assert.Equal(t, http.StatusCreated, r.Code)
		cleanupRegisterProvisioning(t, conf, r)
	})
	t.Run("GoodAdvertiseUrlAccepted", func(t *testing.T) {
		app, router, conf := NewApiTest()
		enablePortalAPIs(t, conf)
		conf.Options().JoinToken = cluster.ExampleJoinToken
		ClusterNodesRegister(router)

		// https is allowed for public host
		r := AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/cluster/nodes/register", `{"NodeName":"pp-node-04","AdvertiseUrl":"https://example.com"}`, cluster.ExampleJoinToken)
		assert.Equal(t, http.StatusCreated, r.Code)
		cleanupRegisterProvisioning(t, conf, r)

		// http is allowed for localhost
		r = AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/cluster/nodes/register", `{"NodeName":"pp-node-04b","AdvertiseUrl":"http://localhost:2342"}`, cluster.ExampleJoinToken)
		assert.Equal(t, http.StatusCreated, r.Code)
		cleanupRegisterProvisioning(t, conf, r)
	})
	t.Run("SiteUrlValidation", func(t *testing.T) {
		app, router, conf := NewApiTest()
		enablePortalAPIs(t, conf)
		conf.Options().JoinToken = cluster.ExampleJoinToken
		ClusterNodesRegister(router)

		// Reject http SiteUrl for public host
		r := AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/cluster/nodes/register", `{"NodeName":"pp-node-05","SiteUrl":"http://example.com"}`, cluster.ExampleJoinToken)
		assert.Equal(t, http.StatusBadRequest, r.Code)

		// Accept https SiteUrl
		r = AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/cluster/nodes/register", `{"NodeName":"pp-node-06","SiteUrl":"https://photos.example.com"}`, cluster.ExampleJoinToken)
		assert.Equal(t, http.StatusCreated, r.Code)
		cleanupRegisterProvisioning(t, conf, r)
	})
	t.Run("NormalizeName", func(t *testing.T) {
		app, router, conf := NewApiTest()
		enablePortalAPIs(t, conf)
		conf.Options().JoinToken = cluster.ExampleJoinToken
		ClusterNodesRegister(router)

		// Mixed separators and case should normalize to DNS label
		body := `{"NodeName":"My.Node/Name:Prod"}`
		r := AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/cluster/nodes/register", body, cluster.ExampleJoinToken)
		assert.Equal(t, http.StatusCreated, r.Code)
		cleanupRegisterProvisioning(t, conf, r)

		regy, err := reg.NewClientRegistryWithConfig(conf)
		assert.NoError(t, err)
		n, err := regy.FindByName("my-node-name-prod")
		assert.NoError(t, err)
		if assert.NotNil(t, n) {
			assert.Equal(t, "my-node-name-prod", n.Name)
		}
	})
	t.Run("BadName", func(t *testing.T) {
		app, router, conf := NewApiTest()
		enablePortalAPIs(t, conf)
		conf.Options().JoinToken = cluster.ExampleJoinToken
		ClusterNodesRegister(router)

		// Empty nodeName → 400
		r := AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/cluster/nodes/register", `{"NodeName":""}`, cluster.ExampleJoinToken)
		assert.Equal(t, http.StatusBadRequest, r.Code)
	})
	t.Run("RotateSecretPersistsAndRespondsOK", func(t *testing.T) {
		app, router, conf := NewApiTest()
		enablePortalAPIs(t, conf)
		conf.Options().JoinToken = cluster.ExampleJoinToken
		ClusterNodesRegister(router)

		// Pre-create node in registry so handler goes through existing-node path
		// and rotates the secret before attempting DB ensure. Don't reuse the
		// Monitoring fixture client ID to avoid changing its secret, which is
		// used by OAuth tests running in the same package.
		regy, err := reg.NewClientRegistryWithConfig(conf)
		assert.NoError(t, err)
		n := &reg.Node{Node: cluster.Node{UUID: rnd.UUIDv7(), Name: "pp-node-01", Role: cluster.RoleInstance}}
		assert.NoError(t, regy.Put(n))
		n, err = regy.RotateSecret(n.UUID)
		if !assert.NoError(t, err) || !assert.NotNil(t, n) {
			return
		}

		token := oauthNodeAccessToken(t, app, router, conf, n.ClientID, n.ClientSecret)
		body := `{"NodeName":"pp-node-01","RotateSecret":true}`
		r := AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/cluster/nodes/register", body, token)
		assert.Equal(t, http.StatusOK, r.Code)
		cleanupRegisterProvisioning(t, conf, r)

		// Secret should have rotated and been persisted even though DB ensure failed.
		// Fetch by name (most-recently-updated) to avoid flakiness if another test adds
		// a node with the same name and a different id.
		n2, err := regy.FindByName("pp-node-01")
		assert.NoError(t, err)
		// With client-backed registry, plaintext secret is not persisted; only rotation timestamp is updated.
		if assert.NotNil(t, n2) {
			assert.NotEmpty(t, n2.RotatedAt)
		}
	})
	t.Run("RotateSecretRequiresMatchingNodeClientCredentials", func(t *testing.T) {
		app, router, conf := NewApiTest()
		enablePortalAPIs(t, conf)
		conf.Options().JoinToken = cluster.ExampleJoinToken
		ClusterNodesRegister(router)

		regy, err := reg.NewClientRegistryWithConfig(conf)
		assert.NoError(t, err)

		victim := &reg.Node{Node: cluster.Node{UUID: rnd.UUIDv7(), Name: "pp-victim", Role: cluster.RoleInstance}}
		assert.NoError(t, regy.Put(victim))
		victim, err = regy.RotateSecret(victim.UUID)
		if !assert.NoError(t, err) || !assert.NotNil(t, victim) {
			return
		}

		attacker := &reg.Node{Node: cluster.Node{UUID: rnd.UUIDv7(), Name: "pp-attacker", Role: cluster.RoleInstance}}
		assert.NoError(t, regy.Put(attacker))
		attacker, err = regy.RotateSecret(attacker.UUID)
		if !assert.NoError(t, err) || !assert.NotNil(t, attacker) {
			return
		}

		// Using another node's valid access token must not authorize rotating the victim.
		token := oauthNodeAccessToken(t, app, router, conf, attacker.ClientID, attacker.ClientSecret)
		body := `{"NodeName":"pp-victim","RotateSecret":true}`
		r := AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/cluster/nodes/register", body, token)
		assert.Equal(t, http.StatusConflict, r.Code)
		assert.NotContains(t, r.Body.String(), "\"ClientSecret\"")
		assert.NotContains(t, r.Body.String(), "\"Password\"")
	})
	t.Run("ExistingNodeMutationRequiresWriteScope", func(t *testing.T) {
		app, router, conf := NewApiTest()
		enablePortalAPIs(t, conf)
		conf.Options().JoinToken = cluster.ExampleJoinToken
		ClusterNodesRegister(router)

		regy, err := reg.NewClientRegistryWithConfig(conf)
		assert.NoError(t, err)

		node := &reg.Node{Node: cluster.Node{UUID: rnd.UUIDv7(), Name: "pp-scope-check", Role: cluster.RoleInstance}}
		assert.NoError(t, regy.Put(node))
		node, err = regy.RotateSecret(node.UUID)
		if !assert.NoError(t, err) || !assert.NotNil(t, node) {
			return
		}

		client := entity.FindClientByUID(node.ClientID)
		if assert.NotNil(t, client) {
			client.SetScope("metrics")
			assert.NoError(t, client.Save())
		}

		// Tokens that do not include cluster permissions must be denied.
		token := oauthNodeAccessToken(t, app, router, conf, node.ClientID, node.ClientSecret)
		r := AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/cluster/nodes/register", `{"NodeName":"pp-scope-check","SiteUrl":"https://scope.example.com"}`, token)
		assert.Equal(t, http.StatusForbidden, r.Code)
	})
	t.Run("RotateDatabaseWithJoinTokenReturnsConflict", func(t *testing.T) {
		app, router, conf := NewApiTest()
		enablePortalAPIs(t, conf)
		conf.Options().JoinToken = cluster.ExampleJoinToken
		ClusterNodesRegister(router)

		regy, err := reg.NewClientRegistryWithConfig(conf)
		assert.NoError(t, err)
		n := &reg.Node{Node: cluster.Node{Name: "pp-node-db-rotate", Role: cluster.RoleInstance}}
		assert.NoError(t, regy.Put(n))

		r := AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/cluster/nodes/register", `{"NodeName":"pp-node-db-rotate","RotateDatabase":true}`, cluster.ExampleJoinToken)
		assert.Equal(t, http.StatusConflict, r.Code)
	})
	t.Run("ExistingNodeSiteUrlPersistsAndRespondsOK", func(t *testing.T) {
		app, router, conf := NewApiTest()
		enablePortalAPIs(t, conf)
		conf.Options().JoinToken = cluster.ExampleJoinToken
		ClusterNodesRegister(router)

		// Pre-create node in registry so handler goes through existing-node path.
		regy, err := reg.NewClientRegistryWithConfig(conf)
		assert.NoError(t, err)
		n := &reg.Node{Node: cluster.Node{UUID: rnd.UUIDv7(), Name: "pp-node-02", Role: cluster.RoleInstance}}
		assert.NoError(t, regy.Put(n))

		n, err = regy.RotateSecret(n.UUID)
		if !assert.NoError(t, err) || !assert.NotNil(t, n) {
			return
		}

		// Provisioner is independent; endpoint should respond 200 and persist metadata.
		token := oauthNodeAccessToken(t, app, router, conf, n.ClientID, n.ClientSecret)
		r := AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/cluster/nodes/register", `{"NodeName":"pp-node-02","SiteUrl":"https://Photos.Example.COM"}`, token)
		assert.Equal(t, http.StatusOK, r.Code)
		cleanupRegisterProvisioning(t, conf, r)

		// Ensure normalized/persisted SiteUrl.
		n2, err := regy.FindByName("pp-node-02")
		assert.NoError(t, err)
		assert.Equal(t, "https://photos.example.com", n2.SiteUrl)
	})
	t.Run("ExistingNodeWithoutRotateDoesNotProvisionDatabase", func(t *testing.T) {
		app, router, conf := NewApiTest()
		enablePortalAPIs(t, conf)
		conf.Options().JoinToken = cluster.ExampleJoinToken
		// If provisioning is called unexpectedly, this invalid DSN should cause a 409.
		conf.Options().DatabaseProvisionDSN = "invalid-dsn"
		ClusterNodesRegister(router)

		regy, err := reg.NewClientRegistryWithConfig(conf)
		assert.NoError(t, err)

		n := &reg.Node{Node: cluster.Node{UUID: rnd.UUIDv7(), Name: "pp-node-no-provision", Role: cluster.RoleInstance}}
		assert.NoError(t, regy.Put(n))
		n, err = regy.RotateSecret(n.UUID)
		if !assert.NoError(t, err) || !assert.NotNil(t, n) {
			return
		}

		token := oauthNodeAccessToken(t, app, router, conf, n.ClientID, n.ClientSecret)
		r := AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/cluster/nodes/register", `{"NodeName":"pp-node-no-provision","SiteUrl":"https://photos.example.com"}`, token)
		assert.Equal(t, http.StatusOK, r.Code)
		body := r.Body.String()
		assert.False(t, gjson.Get(body, "AlreadyProvisioned").Bool())
		assert.Equal(t, "", gjson.Get(body, "Database.Name").String())
		assert.Equal(t, "", gjson.Get(body, "Database.User").String())
	})
	t.Run("AssignNodeUUIDWhenMissing", func(t *testing.T) {
		app, router, conf := NewApiTest()
		enablePortalAPIs(t, conf)
		conf.Options().JoinToken = cluster.ExampleJoinToken
		ClusterNodesRegister(router)

		// Register without nodeUUID; server should assign one (UUID v7 preferred).
		r := AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/cluster/nodes/register", `{"NodeName":"pp-node-uuid"}`, cluster.ExampleJoinToken)
		assert.Equal(t, http.StatusCreated, r.Code)
		cleanupRegisterProvisioning(t, conf, r)

		// Response must include Node.UUID and the client identifier the node persists.
		body := r.Body.String()
		assert.NotEmpty(t, gjson.Get(body, "Node.UUID").String())
		assert.NotEmpty(t, gjson.Get(body, "Node.ClientID").String())

		// Verify it is persisted in the registry
		regy, err := reg.NewClientRegistryWithConfig(conf)
		assert.NoError(t, err)
		n, err := regy.FindByName("pp-node-uuid")
		assert.NoError(t, err)
		if assert.NotNil(t, n) {
			assert.NotEmpty(t, n.UUID)
		}
	})
	t.Run("ThemeHintProvided", func(t *testing.T) {
		app, router, conf := NewApiTest()
		enablePortalAPIs(t, conf)
		conf.Options().JoinToken = cluster.ExampleJoinToken
		ClusterNodesRegister(router)

		themeDir := conf.PortalThemePath()
		assert.NoError(t, os.MkdirAll(themeDir, fs.ModeDir))
		assert.NoError(t, os.WriteFile(filepath.Join(themeDir, fs.AppJsFile), []byte("// app\n"), fs.ModeFile))
		assert.NoError(t, os.WriteFile(filepath.Join(themeDir, fs.VersionTxtFile), []byte(" 2.0.0\n"), fs.ModeFile))
		t.Cleanup(func() { _ = os.RemoveAll(themeDir) })

		body := `{"NodeName":"pp-node-theme","Theme":"1.0.0"}`
		r := AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/cluster/nodes/register", body, cluster.ExampleJoinToken)
		assert.Equal(t, http.StatusCreated, r.Code)
		assert.Equal(t, "2.0.0", gjson.Get(r.Body.String(), "Theme").String())
		cleanupRegisterProvisioning(t, conf, r)

		regy, err := reg.NewClientRegistryWithConfig(conf)
		assert.NoError(t, err)
		node, err := regy.FindByName("pp-node-theme")
		assert.NoError(t, err)
		if assert.NotNil(t, node) {
			assert.Equal(t, "1.0.0", node.Theme)
		}

		node, err = regy.RotateSecret(node.UUID)
		if !assert.NoError(t, err) || !assert.NotNil(t, node) {
			return
		}
		token := oauthNodeAccessToken(t, app, router, conf, node.ClientID, node.ClientSecret)

		body = `{"NodeName":"pp-node-theme","Theme":"2.0.0"}`
		r2 := AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/cluster/nodes/register", body, token)
		assert.Equal(t, http.StatusOK, r2.Code)
		assert.Equal(t, "2.0.0", gjson.Get(r2.Body.String(), "Theme").String())
		cleanupRegisterProvisioning(t, conf, r2)
	})
}

func cleanupRegisterProvisioning(t *testing.T, conf *config.Config, r *httptest.ResponseRecorder) {
	t.Helper()

	if r.Code != http.StatusOK && r.Code != http.StatusCreated {
		return
	}

	var resp cluster.RegisterResponse
	if err := json.Unmarshal(r.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal register response: %v", err)
	}

	if !resp.AlreadyProvisioned {
		return
	}

	name := resp.Database.Name
	user := resp.Database.User

	if conf != nil && (name == "" || user == "") && resp.Node.Name != "" && resp.Node.UUID != "" {
		genName, genUser, _ := provisioner.GenerateCredentials(conf, resp.Node.UUID, resp.Node.Name)
		if name == "" {
			name = genName
		}
		if user == "" {
			user = genUser
		}
	}

	if name == "" && user == "" {
		return
	}

	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := provisioner.DropCredentials(ctx, name, user); err != nil {
			t.Fatalf("drop credentials for %s/%s: %v", name, user, err)
		}
	})
}

func AuthenticatedRequestWithBodyAndIP(r http.Handler, method, path, body string, authToken string, clientIP string) *httptest.ResponseRecorder {
	reader := strings.NewReader(body)
	req, _ := http.NewRequest(method, path, reader)
	req.RemoteAddr = clientIP + ":12345"
	header.SetAuthorization(req, authToken)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	return w
}

func oauthNodeAccessToken(t testing.TB, app http.Handler, router *gin.RouterGroup, conf *config.Config, clientID, clientSecret string) string {
	return oauthNodeAccessTokenWithScope(t, app, router, conf, clientID, clientSecret, "cluster")
}

func oauthNodeAccessTokenWithScope(t testing.TB, app http.Handler, router *gin.RouterGroup, conf *config.Config, clientID, clientSecret, scope string) string {
	t.Helper()

	prevAuthMode := conf.AuthMode()
	conf.SetAuthMode(config.AuthModePasswd)
	t.Cleanup(func() {
		conf.SetAuthMode(prevAuthMode)
	})

	OAuthToken(router)

	data := url.Values{
		"grant_type":    {authn.GrantClientCredentials.String()},
		"client_id":     {clientID},
		"client_secret": {clientSecret},
		"scope":         {scope},
	}

	req, _ := http.NewRequest(http.MethodPost, "/api/v1/oauth/token", strings.NewReader(data.Encode()))
	req.Header.Add(header.ContentType, header.ContentTypeForm)

	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code, "oauth token request failed: %s", w.Body.String())

	return gjson.Get(w.Body.String(), "access_token").String()
}

// TestBuildPortalLoginURL covers the browser-facing Portal login URL reported
// to nodes at registration: SiteUrl origin + login route, with an absolute
// custom LoginUri returned as-is.
func TestBuildPortalLoginURL(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		_, _, conf := NewApiTest()
		prevSite, prevLogin := conf.Options().SiteUrl, conf.Options().LoginUri
		t.Cleanup(func() { conf.Options().SiteUrl, conf.Options().LoginUri = prevSite, prevLogin })

		conf.Options().SiteUrl = "https://app.example.com/"
		conf.Options().LoginUri = ""
		assert.Equal(t, "https://app.example.com"+conf.LoginUri(), buildPortalLoginURL(conf))
	})
	t.Run("AbsoluteLoginUri", func(t *testing.T) {
		_, _, conf := NewApiTest()
		conf.SetAuthMode(config.AuthModePasswd)
		prevLogin := conf.Options().LoginUri
		t.Cleanup(func() {
			conf.Options().LoginUri = prevLogin
			conf.SetAuthMode(config.AuthModePublic)
		})

		conf.Options().LoginUri = "https://sso.example.com/login"
		assert.Equal(t, "https://sso.example.com/login", buildPortalLoginURL(conf))
	})
	t.Run("Nil", func(t *testing.T) {
		assert.Equal(t, "", buildPortalLoginURL(nil))
	})
}

// TestSanitizeAllowGroupRoles validates the lenient registration-time mapping
// sanitizer: malformed keys and non-federatable roles are dropped, not errors.
func TestSanitizeAllowGroupRoles(t *testing.T) {
	t.Run("AcceptsAllInstanceRoles", func(t *testing.T) {
		out := sanitizeAllowGroupRoles(map[string]string{
			"g-admin": "admin", "g-manager": "manager", "g-user": "user",
			"g-contributor": "contributor", "g-viewer": "viewer", "g-guest": "guest",
		})
		assert.Equal(t, map[string]string{
			"g-admin": "admin", "g-manager": "manager", "g-user": "user",
			"g-contributor": "contributor", "g-viewer": "viewer", "g-guest": "guest",
		}, out)
	})
	t.Run("DropsInvalidEntries", func(t *testing.T) {
		out := sanitizeAllowGroupRoles(map[string]string{"a": "cluster_admin", "b": "visitor", "c": "bogus", "   ": "admin"})
		assert.Nil(t, out)
	})
	t.Run("Empty", func(t *testing.T) {
		assert.Nil(t, sanitizeAllowGroupRoles(nil))
	})
}

func TestNodeUUIDClaimedBy(t *testing.T) {
	c := config.TestConfig()
	regy, err := reg.NewClientRegistryWithConfig(c)
	assert.NoError(t, err)

	n := &reg.Node{Node: cluster.Node{UUID: rnd.UUIDv7(), Name: "pp-claimed", Role: cluster.RoleInstance}}
	assert.NoError(t, regy.Create(n))

	t.Run("DifferentClient", func(t *testing.T) {
		assert.True(t, nodeUUIDClaimedBy(n.UUID, rnd.GenerateUID(entity.ClientUID)))
	})
	t.Run("OwnClient", func(t *testing.T) {
		assert.False(t, nodeUUIDClaimedBy(n.UUID, n.ClientID))
	})
	t.Run("Unregistered", func(t *testing.T) {
		assert.False(t, nodeUUIDClaimedBy(rnd.UUIDv7(), ""))
	})
	t.Run("Empty", func(t *testing.T) {
		assert.False(t, nodeUUIDClaimedBy("", ""))
	})
	t.Run("DuplicateRecords", func(t *testing.T) {
		// The column is not unique, so a second record sharing the UUID must still be seen
		// even when the newest one belongs to the caller.
		dup := entity.NewClient()
		dup.ClientName = "pp-claimed-dup"
		dup.NodeUUID = n.UUID
		assert.NoError(t, dup.Create())
		t.Cleanup(func() { assert.NoError(t, dup.Delete()) })

		assert.True(t, nodeUUIDClaimedBy(n.UUID, n.ClientID))
		assert.True(t, nodeUUIDClaimedBy(n.UUID, dup.ClientUID))
	})
}

func TestRegisterUUIDConflictError(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		uuid := rnd.UUIDv7()
		assert.Contains(t, registerUUIDConflictError(uuid), uuid)
		assert.Contains(t, registerUUIDConflictError(uuid), "already registered")
	})
	t.Run("Empty", func(t *testing.T) {
		assert.Contains(t, registerUUIDConflictError(""), "already registered")
	})
}

// TestClusterNodesRegister_GroupConfig covers the declarative per-node group
// config: a registration declares the policy, an admin PATCH pins it against
// re-registration, and clearing the admin override lets the declared config
// repopulate on the next registration.
func TestClusterNodesRegister_GroupConfig(t *testing.T) {
	app, router, conf := NewApiTest()
	enablePortalAPIs(t, conf)
	conf.Options().JoinToken = cluster.ExampleJoinToken
	ClusterNodesRegister(router)
	ClusterUpdateNode(router)

	groupData := func(uuid string) *entity.ClientData {
		client := entity.FindClientByNodeUUID(uuid)
		if client == nil {
			t.Fatalf("client for node %s not found", uuid)
		}
		return client.GetData()
	}

	// A new node declares its group config at join; invalid roles are dropped.
	body := `{"NodeName":"pp-groups","AllowGroups":["Media-Acme-Viewer"],"AllowGroupRoles":{"Media-Acme-Admin":"admin","Bad":"cluster_admin"},"GroupsFullView":true}`
	r := AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/cluster/nodes/register", body, cluster.ExampleJoinToken)
	assert.Equal(t, http.StatusCreated, r.Code, "body=%s", r.Body.String())
	cleanupRegisterProvisioning(t, conf, r)

	regy, err := reg.NewClientRegistryWithConfig(conf)
	assert.NoError(t, err)
	n, err := regy.FindByName("pp-groups")
	assert.NoError(t, err)

	t.Run("DeclaredAtJoin", func(t *testing.T) {
		data := groupData(n.UUID)
		assert.Equal(t, []string{"media-acme-viewer"}, data.AllowGroups)
		assert.Equal(t, map[string]string{"media-acme-admin": "admin"}, data.AllowGroupRoles, "non-federatable roles must be dropped")
		assert.True(t, data.GroupsFullView)
		assert.Equal(t, entity.ClientGroupsSrcNode, data.GroupsSrc)
	})

	// Subsequent registrations need the node's own OAuth credentials, and the
	// admin PATCH calls need a session since the token helper switches the
	// test config to password-auth mode.
	nr, err := regy.RotateSecret(n.UUID)
	assert.NoError(t, err)
	token := oauthNodeAccessToken(t, app, router, conf, nr.ClientID, nr.ClientSecret)
	adminToken := AuthenticateAdmin(app, router)

	t.Run("RedeclaredOnReRegister", func(t *testing.T) {
		body := `{"NodeName":"pp-groups","AllowGroups":["Media-Acme-User"]}`
		r := AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/cluster/nodes/register", body, token)
		assert.Equal(t, http.StatusOK, r.Code, "body=%s", r.Body.String())
		cleanupRegisterProvisioning(t, conf, r)

		data := groupData(n.UUID)
		assert.Equal(t, []string{"media-acme-user"}, data.AllowGroups)
		assert.Empty(t, data.AllowGroupRoles, "a declaration replaces the config wholesale")
		assert.False(t, data.GroupsFullView)
	})
	t.Run("AdminOverrideSurvivesReRegister", func(t *testing.T) {
		r := AuthenticatedRequestWithBody(app, http.MethodPatch, "/api/v1/cluster/nodes/"+n.UUID,
			`{"AllowGroups":["media-acme-pinned"],"AllowGroupRoles":{"media-acme-pinned":"guest"}}`, adminToken)
		assert.Equal(t, http.StatusOK, r.Code, "body=%s", r.Body.String())

		body := `{"NodeName":"pp-groups","AllowGroups":["Media-Acme-User"],"GroupsFullView":true}`
		r = AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/cluster/nodes/register", body, token)
		assert.Equal(t, http.StatusOK, r.Code, "body=%s", r.Body.String())
		cleanupRegisterProvisioning(t, conf, r)

		data := groupData(n.UUID)
		assert.Equal(t, []string{"media-acme-pinned"}, data.AllowGroups, "re-registration must not clobber an admin override")
		assert.Equal(t, map[string]string{"media-acme-pinned": "guest"}, data.AllowGroupRoles)
		assert.False(t, data.GroupsFullView)
		assert.Equal(t, entity.ClientGroupsSrcManual, data.GroupsSrc)
	})
	t.Run("ClearedOverrideRepopulates", func(t *testing.T) {
		r := AuthenticatedRequestWithBody(app, http.MethodPatch, "/api/v1/cluster/nodes/"+n.UUID,
			`{"AllowGroups":[],"AllowGroupRoles":{},"GroupsFullView":false}`, adminToken)
		assert.Equal(t, http.StatusOK, r.Code, "body=%s", r.Body.String())
		assert.Equal(t, "", groupData(n.UUID).GroupsSrc, "a fully cleared manual config must un-pin")

		body := `{"NodeName":"pp-groups","AllowGroups":["Media-Acme-User"]}`
		r = AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/cluster/nodes/register", body, token)
		assert.Equal(t, http.StatusOK, r.Code, "body=%s", r.Body.String())
		cleanupRegisterProvisioning(t, conf, r)

		data := groupData(n.UUID)
		assert.Equal(t, []string{"media-acme-user"}, data.AllowGroups)
		assert.Equal(t, entity.ClientGroupsSrcNode, data.GroupsSrc)
	})
	t.Run("EmptyDeclarationClears", func(t *testing.T) {
		body := `{"NodeName":"pp-groups"}`
		r := AuthenticatedRequestWithBody(app, http.MethodPost, "/api/v1/cluster/nodes/register", body, token)
		assert.Equal(t, http.StatusOK, r.Code, "body=%s", r.Body.String())
		cleanupRegisterProvisioning(t, conf, r)

		data := groupData(n.UUID)
		assert.Empty(t, data.AllowGroups, "an instance that stops declaring a config clears it")
		assert.Equal(t, "", data.GroupsSrc)
	})
}

// TestValidateAdvertiseURL ensures the validator accepts HTTP and HTTPS for advertise URLs.
func TestValidateAdvertiseURL(t *testing.T) {
	cases := []struct {
		u  string
		ok bool
	}{
		{"https://example.com", true},
		{"http://example.com", true},
		{"http://localhost:2342", true},
		{"http://photoprism.default.svc", true},
		{"http://photoprism.default.svc.cluster.local", true},
		{"http://photoprism.internal", true},
		{"https://127.0.0.1", true},
		{"ftp://example.com", false},
		{"https://", false},
		{"", false},
	}
	for _, c := range cases {
		if got := validateAdvertiseURL(c.u); got != c.ok {
			t.Fatalf("validateAdvertiseURL(%q) = %v, want %v", c.u, got, c.ok)
		}
	}
}

// TestValidateSiteURL enforces HTTPS for non-local site URLs.
func TestValidateSiteURL(t *testing.T) {
	cases := []struct {
		u  string
		ok bool
	}{
		{"https://photos.example.com", true},
		{"http://photos.example.com", false},
		{"http://127.0.0.1:2342", true},
		{"mailto:me@example.com", false},
		{"://bad", false},
	}
	for _, c := range cases {
		if got := validateSiteURL(c.u); got != c.ok {
			t.Fatalf("validateSiteURL(%q) = %v, want %v", c.u, got, c.ok)
		}
	}
}
