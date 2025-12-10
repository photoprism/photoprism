package config

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func resetExtensionsForTest(t *testing.T) func() {
	t.Helper()

	extMutex.Lock()
	originalInit, _ := extensions.Load().(Extensions)
	savedInit := append(Extensions(nil), originalInit...)
	extensions.Store(Extensions{})
	extMutex.Unlock()

	bootMutex.Lock()
	originalBoot, _ := bootExtensions.Load().(Extensions)
	savedBoot := append(Extensions(nil), originalBoot...)
	bootExtensions.Store(Extensions{})
	bootMutex.Unlock()

	extInit = sync.Once{}
	bootInit = sync.Once{}

	return func() {
		extMutex.Lock()
		extensions.Store(savedInit)
		extMutex.Unlock()

		bootMutex.Lock()
		bootExtensions.Store(savedBoot)
		bootMutex.Unlock()

		extInit = sync.Once{}
		bootInit = sync.Once{}
	}
}

func TestRegisterStageBootRunsOnce(t *testing.T) {
	cleanup := resetExtensionsForTest(t)
	defer cleanup()

	var calls []string

	Register(StageBoot, "first", func(*Config) error {
		calls = append(calls, "first")
		return nil
	}, nil)
	Register(StageBoot, "second", func(*Config) error {
		calls = append(calls, "second")
		return nil
	}, nil)

	boot := Ext(StageBoot)
	assert.Len(t, boot, 2)

	cfg := &Config{}
	boot.Boot(cfg)
	assert.Equal(t, []string{"first", "second"}, calls)

	boot.Boot(cfg)
	assert.Equal(t, []string{"first", "second"}, calls, "boot extensions should run only once")
}

func TestRegisterStageInitRunsOnce(t *testing.T) {
	cleanup := resetExtensionsForTest(t)
	defer cleanup()

	var calls []string

	Register(StageInit, "alpha", func(*Config) error {
		calls = append(calls, "alpha")
		return nil
	}, nil)

	ext := Ext(StageInit)
	assert.Len(t, ext, 1)

	cfg := &Config{}
	ext.Init(cfg)
	assert.Equal(t, []string{"alpha"}, calls)

	ext.Init(cfg)
	assert.Equal(t, []string{"alpha"}, calls, "init extensions should run only once")
}

func TestRegisterInvalidStageIgnored(t *testing.T) {
	cleanup := resetExtensionsForTest(t)
	defer cleanup()

	Register(Stage("invalid"), "ignored", func(*Config) error { return nil }, nil)

	assert.Len(t, Ext(StageBoot), 0)
	assert.Len(t, Ext(StageInit), 0)
}

func TestClientExtIncludesBootAndInitValues(t *testing.T) {
	cleanup := resetExtensionsForTest(t)
	defer cleanup()

	Register(StageBoot, "boot-ext", func(*Config) error { return nil }, func(*Config, ClientType) Values {
		return Values{"boot": true}
	})
	Register(StageInit, "init-ext", func(*Config) error { return nil }, func(*Config, ClientType) Values {
		return Values{"init": true}
	})

	values := ClientExt(&Config{}, ClientPublic)

	require.Contains(t, values, "boot-ext")
	require.Contains(t, values, "init-ext")
	assert.Equal(t, Values{"boot": true}, values["boot-ext"])
	assert.Equal(t, Values{"init": true}, values["init-ext"])
}

func TestClientExtSkipsNilClientValues(t *testing.T) {
	cleanup := resetExtensionsForTest(t)
	defer cleanup()

	Register(StageBoot, "boot-no-client", func(*Config) error { return nil }, nil)
	Register(StageInit, "init-no-client", func(*Config) error { return nil }, nil)

	values := ClientExt(&Config{}, ClientPublic)
	assert.Empty(t, values)
}
