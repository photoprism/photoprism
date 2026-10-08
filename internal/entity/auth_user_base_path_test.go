package entity

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/photoprism/photoprism/internal/auth/acl"
	"github.com/photoprism/photoprism/internal/event"
	"github.com/photoprism/photoprism/internal/form"
	"github.com/photoprism/photoprism/pkg/fs"
)

// useOriginalsPath sets OriginalsPath to a new folder holding the given base paths until the test ends.
func useOriginalsPath(t *testing.T, basePaths ...string) {
	t.Helper()

	orig := OriginalsPath
	OriginalsPath = t.TempDir()
	t.Cleanup(func() { OriginalsPath = orig })

	for _, basePath := range basePaths {
		require.NoError(t, os.MkdirAll(filepath.Join(OriginalsPath, basePath), fs.ModeDir))
	}
}

// captureAuditLog records audit log entries until the test ends.
func captureAuditLog(t *testing.T) *test.Hook {
	t.Helper()

	logger, hook := test.NewNullLogger()
	logger.SetLevel(logrus.TraceLevel)
	orig := event.AuditLog
	event.AuditLog = logger
	t.Cleanup(func() { event.AuditLog = orig })

	return hook
}

// basePathWarnings returns the number of logged warnings about an existing default base path.
func basePathWarnings(hook *test.Hook) (n int) {
	for _, e := range hook.AllEntries() {
		if e.Level == logrus.WarnLevel && strings.Contains(e.Message, "default base path") {
			n++
		}
	}

	return n
}

func TestUser_WarnExistingBasePath(t *testing.T) {
	registerContributorRole(t)

	t.Run("Exists", func(t *testing.T) {
		useOriginalsPath(t, "users/bp-exists")
		hook := captureAuditLog(t)
		m := &User{UserName: "bp-exists", UserRole: acl.RoleContributor.String()}
		assert.True(t, m.WarnExistingBasePath())
		require.Equal(t, 1, basePathWarnings(hook))
		assert.Contains(t, hook.LastEntry().Message, "users/bp-exists")
		assert.Contains(t, hook.LastEntry().Message, "bp-exists")
	})
	t.Run("ExplicitDefault", func(t *testing.T) {
		useOriginalsPath(t, "users/bp-explicit")
		m := &User{UserName: "bp-explicit", UserRole: acl.RoleContributor.String(), BasePath: "users/bp-explicit"}
		assert.True(t, m.WarnExistingBasePath())
	})
	t.Run("Missing", func(t *testing.T) {
		useOriginalsPath(t, "users/other")
		m := &User{UserName: "bp-missing", UserRole: acl.RoleContributor.String()}
		assert.False(t, m.WarnExistingBasePath())
	})
	t.Run("OtherBasePath", func(t *testing.T) {
		useOriginalsPath(t, "users/bp-other", "shared")
		m := &User{UserName: "bp-other", UserRole: acl.RoleContributor.String(), BasePath: "shared"}
		assert.False(t, m.WarnExistingBasePath())
	})
	t.Run("NotContributor", func(t *testing.T) {
		useOriginalsPath(t, "users/bp-admin")
		m := &User{UserName: "bp-admin", UserRole: acl.RoleAdmin.String()}
		assert.False(t, m.WarnExistingBasePath())
	})
	t.Run("NoOriginalsPath", func(t *testing.T) {
		dir := t.TempDir()
		require.NoError(t, os.MkdirAll(filepath.Join(dir, "users", "bp-none"), fs.ModeDir))
		t.Chdir(dir)
		orig := OriginalsPath
		OriginalsPath = ""
		t.Cleanup(func() { OriginalsPath = orig })
		m := &User{UserName: "bp-none", UserRole: acl.RoleContributor.String()}
		assert.False(t, m.WarnExistingBasePath())
	})
	t.Run("NoDefault", func(t *testing.T) {
		useOriginalsPath(t)
		m := &User{UserName: ".bp-nohandle", UserRole: acl.RoleContributor.String()}
		require.Equal(t, "", m.DefaultBasePath())
		assert.False(t, m.WarnExistingBasePath())
	})
	t.Run("Nil", func(t *testing.T) {
		var m *User
		assert.False(t, m.WarnExistingBasePath())
	})
}

func TestUser_Create_ExistingBasePath(t *testing.T) {
	registerContributorRole(t)
	useOriginalsPath(t, "users/bp-create")
	hook := captureAuditLog(t)

	m := NewUser()
	m.UserName = "bp-create"
	m.UserRole = acl.RoleContributor.String()
	require.NoError(t, m.Create())
	t.Cleanup(func() { deleteTestUser(t, m) })

	assert.Equal(t, 1, basePathWarnings(hook))
	assert.Equal(t, "", m.BasePath)
}

func TestAddUser_ExistingBasePath(t *testing.T) {
	registerContributorRole(t)
	useOriginalsPath(t, "users/bp-add")
	hook := captureAuditLog(t)

	t.Cleanup(func() { deleteTestUser(t, FindUserByName("bp-add")) })
	require.NoError(t, AddUser(form.User{UserName: "bp-add", UserRole: acl.RoleContributor.String(), AuthProvider: "none"}))
	require.NotNil(t, FindUserByName("bp-add"))

	assert.Equal(t, 1, basePathWarnings(hook))
}
