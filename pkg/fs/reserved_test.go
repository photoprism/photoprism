package fs

import (
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHasReservedComponent checks complete components without expanding ordinary hidden paths.
func TestHasReservedComponent(t *testing.T) {
	for _, name := range ReservedPathNames() {
		assert.True(t, HasReservedComponent(name), name)
		assert.True(t, HasReservedComponent("nested/"+strings.ToUpper(name)+"/photo.jpg"), name)
	}

	for _, name := range []string{".Xauthority", ".my.cnf", ".mylogin.cnf", "nested/.CUSTOM.CNF", `nested\.MYSQL.CNF`, ".ppignore", "nested/.PPIGNORE", ".gitignore", ".rsyncignore", ".dockerignore", ".ignore", ".python_history", ".node_repl_history", ".bash_history", ".mysql_history", ".zsh_history", "nested/.CUSTOMIGNORE/photo.jpg", `nested\.CUSTOM_HISTORY\file`, ".ppignore/.git/config", ".FORGEJO/workflows/build.yml", `nested\.Forgejo\photo.jpg`, "photo.rclonelink", "nested/PHOTO.JPG.RCLONELINK", `nested\link.RcloneLink\photo.jpg`, ".rclonelink", ".bash_history-04218.tmp", "nested/.BASH_HISTORY-86113.TMP", ".env.", ".git", ".config/photo.jpg", "nested/.photoprism/photo.jpg", ".git/photo.jpg", "a/.svn/photo.jpg", ".HG/photo.jpg", "a/.ssh/._photo.jpg", "a/.GnUpG/photo.jpg", ".env", "a/.env.production", "a/.ENV.example/p.jpg", `a\.SSH\photo.jpg`} {
		assert.True(t, HasReservedComponent(name), name)
	}

	for _, name := range []string{"", ".", ".cache-photos/photo.jpg", ".profile.jpg", ".cargo-images/photo.jpg", "my.cnf", "settings.cnf", "folder/photo.cnf", "plainignore", "plain_history", ".custom_history.jpg", ".dockerignore.txt", ".bash_history_notes.txt", ".bash_history-04218.tmp.jpg", ".bash_history-04218", ".env-keys-photo.jpg", ".htaccess.jpg", ".my.cnf.jpg", ".gitkeep", "serial", "keys/photo.jpg", "config/photo.jpg", "signing.key.jpg", "signing.key-backup", "client_secret.txt", ".gitignore.txt", ".github-backup/photo.jpg", "forgejo/photo.jpg", ".forgejo-backup/photo.jpg", ".forgejo.jpg", "photo.rclonelink.jpg", "rclonelink", "photo.rclonelinks", "photo.rclone", "rclonelink/photo.jpg", ".locality/photo.jpg", ".claude-example/photo.jpg", "_netrc.txt", "a/git/photo.jpg", ".environment.jpeg", "._photo.jpg", ".DS_Store", ".hidden/photo.jpg", "%2egit/photo.jpg", "photo.git.jpg"} {
		assert.False(t, HasReservedComponent(name), name)
	}
}

// TestReservedPathNames checks the reusable set and protects it from caller changes.
func TestReservedPathNames(t *testing.T) {
	names := ReservedPathNames()
	assert.Equal(t, []string{".aws", ".azure", ".bash_logout", ".bashrc", ".cache", ".cargo", ".claude", ".claude.json", ".codex", ".composer", ".config", ".docker", ".emulator_console_auth_token", ".env", ".env-keys", ".forgejo", ".git", ".git-credentials", ".gitconfig", ".github", ".gnupg", ".hg", ".htaccess", ".httpie", ".kube", ".local", ".m2", ".mozilla", ".netrc", ".npmrc", ".photoprism", ".ppstorage", ".profile", ".pypirc", ".rsync-filter", ".secrets", ".ssh", ".svn", ".temp", ".thunderbird", ".var", ".xauthority", ".zshenv", ".zshrc", "_netrc", "client_secret", "join_token", "signing.key"}, names)
	names[0] = "changed"
	assert.True(t, HasReservedComponent(".config/photo.jpg"))
	assert.Equal(t, ".aws", ReservedPathNames()[0])
}

// TestReservedPathSuffixes checks the reusable set and protects it from caller changes.
func TestReservedPathSuffixes(t *testing.T) {
	suffixes := ReservedPathSuffixes()
	assert.Equal(t, []string{".rclonelink"}, suffixes)

	for _, suffix := range suffixes {
		assert.NotEmpty(t, suffix)
		assert.Equal(t, strings.ToLower(suffix), suffix)
		assert.True(t, HasReservedComponent("photo"+suffix))
	}

	suffixes[0] = "changed"
	assert.Equal(t, ".rclonelink", ReservedPathSuffixes()[0])
	assert.False(t, HasReservedComponent("photo.changed"))
}

// TestReservedPathPatterns checks pattern validity and isolation from caller changes.
func TestReservedPathPatterns(t *testing.T) {
	patterns := ReservedPathPatterns()
	assert.Equal(t, []string{".env.*", ".*ignore", ".*_history", ".bash_history-*.tmp", ".*.cnf"}, patterns)

	for _, pattern := range patterns {
		assert.True(t, strings.HasPrefix(pattern, "."), pattern)
		_, err := path.Match(pattern, "control")
		require.NoError(t, err)
	}

	patterns[0] = "changed"
	assert.Equal(t, ".env.*", ReservedPathPatterns()[0])
	assert.True(t, HasReservedComponent(".env.production"))
}

// TestHasReservedTarget checks logical roots, aliases, and absent destination suffixes.
func TestHasReservedTarget(t *testing.T) {
	root := filepath.Join(t.TempDir(), ".config", "root")
	require.NoError(t, MkdirAll(filepath.Join(root, ".ssh")))
	require.NoError(t, os.WriteFile(filepath.Join(root, ".ssh/key"), []byte("reserved-control"), ModeFile))
	outside := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(outside, "photo.jpg"), []byte("ordinary-control"), ModeFile))
	require.NoError(t, os.Symlink(filepath.Join(root, ".ssh"), filepath.Join(root, "alias")))
	require.NoError(t, os.Symlink(outside, filepath.Join(root, "ordinary")))
	require.NoError(t, os.Symlink(filepath.Join(root, "missing"), filepath.Join(root, "broken")))

	for _, name := range []string{".ssh/key", "alias/key", "alias/new.jpg"} {
		reserved, err := HasReservedTarget(root, name)
		require.NoError(t, err)
		assert.True(t, reserved, name)
	}

	for _, name := range []string{"photo.jpg", "ordinary/photo.jpg", "ordinary/new.jpg", "new/folder/photo.jpg", "/"} {
		reserved, err := HasReservedTarget(root, name)
		require.NoError(t, err)
		assert.False(t, reserved, name)
	}

	for _, name := range append(ReservedPathNames(), ".bash_history-04218.tmp", ".dockerignore", ".python_history") {
		dir := filepath.Join(root, name)
		require.NoError(t, MkdirAll(dir))
		alias := "alias-" + strings.TrimPrefix(name, ".")
		require.NoError(t, os.Symlink(dir, filepath.Join(root, alias)))
		reserved, err := HasReservedTarget(root, alias+"/new.jpg")
		require.NoError(t, err)
		assert.True(t, reserved, name)
	}

	parentAlias := filepath.Join(filepath.Dir(root), "parent-alias")
	require.NoError(t, os.Symlink(filepath.Join(root, ".ssh"), parentAlias))
	parentReserved, parentErr := HasReservedTarget(root, "../parent-alias/key")
	require.NoError(t, parentErr)
	assert.True(t, parentReserved)
	_, err := HasReservedTarget(root, "broken")
	assert.Error(t, err)
	reserved, err := HasReservedTarget(filepath.Join(root, "new-root"), "new-file")
	require.NoError(t, err)
	assert.False(t, reserved)
}

// TestIsIgnoreFileName checks complete basenames independently of path prefixes.
func TestIsIgnoreFileName(t *testing.T) {
	for _, name := range []string{".ppignore", ".gitignore", ".dockerignore", ".rsyncignore", "folder/.IGNORE", `folder\.CUSTOMIGNORE`} {
		assert.True(t, IsIgnoreFileName(name), name)
	}

	for _, name := range []string{"", "ignore", "plainignore", ".gitignore.txt", ".ignore/photo.jpg", ".python_history"} {
		assert.False(t, IsIgnoreFileName(name), name)
	}
}

// TestReservedPathPolicy checks readable ignore names without admitting other reserved targets.
func TestReservedPathPolicy(t *testing.T) {
	policy := ReservedPathPolicy{AllowIgnoreNames: true}

	for _, name := range []string{".ppignore", "nested/.GITIGNORE", ".dockerignore/photo.jpg"} {
		assert.True(t, HasReservedComponent(name), name)
		assert.False(t, policy.HasReservedComponent(name), name)
	}

	for _, name := range []string{".cache/.ppignore", ".env.ignore", ".gitconfig", ".python_history", ".gitignore/.my.cnf", ".ppignore/.ssh/key"} {
		assert.True(t, policy.HasReservedComponent(name), name)
	}

	root := t.TempDir()
	require.NoError(t, MkdirAll(filepath.Join(root, ".gitignore")))
	require.NoError(t, os.WriteFile(filepath.Join(root, ".gitignore", ".my.cnf"), []byte("control"), ModeFile))
	require.NoError(t, os.Symlink(filepath.Join(root, ".gitignore"), filepath.Join(root, "alias")))

	for _, name := range []string{".gitignore", "alias/new.txt"} {
		reserved, err := HasReservedTarget(root, name)
		require.NoError(t, err)
		assert.True(t, reserved, name)

		reserved, err = policy.HasReservedTarget(root, name)
		require.NoError(t, err)
		assert.False(t, reserved, name)
	}

	reserved, err := policy.HasReservedTarget(root, "alias/.my.cnf")
	require.NoError(t, err)
	assert.True(t, reserved)
}

// TestResolvePathAncestors checks missing-suffix order and unresolved-link errors.
func TestResolvePathAncestors(t *testing.T) {
	root := t.TempDir()
	canonical, err := filepath.EvalSymlinks(root)
	require.NoError(t, err)
	require.NoError(t, MkdirAll(filepath.Join(root, "actual")))
	require.NoError(t, os.Symlink(filepath.Join(root, "actual"), filepath.Join(root, "alias")))
	require.NoError(t, os.Symlink(filepath.Join(root, "absent"), filepath.Join(root, "dangling")))
	require.NoError(t, os.Symlink("cycle", filepath.Join(root, "cycle")))

	for _, tc := range []struct{ label, name, want string }{
		{"ExistingDirectory", "actual", "actual"},
		{"ExistingAlias", "alias", "actual"},
		{"MissingSuffix", "absent/deep/file.jpg", "absent/deep/file.jpg"},
		{"AliasedSuffix", "alias/absent/deep/file.jpg", "actual/absent/deep/file.jpg"},
	} {
		t.Run(tc.label, func(t *testing.T) {
			resolved, resolveErr := resolvePathAncestors(filepath.Join(root, tc.name))
			require.NoError(t, resolveErr)
			assert.Equal(t, filepath.Join(canonical, tc.want), resolved)
		})
	}

	for _, name := range []string{"dangling", "dangling/deep/file.jpg", "cycle", "cycle/deep/file.jpg"} {
		_, resolveErr := resolvePathAncestors(filepath.Join(root, name))
		assert.Error(t, resolveErr, name)
	}
}
