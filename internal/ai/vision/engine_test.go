package vision

import "testing"

func TestRegisterEngineAlias(t *testing.T) {
	const alias = "unit-test"
	engineMu.Lock()
	prev, had := engineAliasIndex[alias]
	if had {
		delete(engineAliasIndex, alias)
	}
	engineMu.Unlock()

	t.Cleanup(func() {
		engineMu.Lock()
		if had {
			engineAliasIndex[alias] = prev
		} else {
			delete(engineAliasIndex, alias)
		}
		engineMu.Unlock()
	})

	RegisterEngineAlias("  Unit-Test  ", EngineInfo{RequestFormat: ApiFormat("custom"), ResponseFormat: "", FileScheme: "data", DefaultResolution: 512})

	info, ok := EngineInfoFor(alias)
	if !ok {
		t.Fatalf("expected engine alias %q to be registered", alias)
	}

	if info.RequestFormat != ApiFormat("custom") {
		t.Errorf("unexpected request format: %s", info.RequestFormat)
	}

	if info.ResponseFormat != ApiFormat("custom") {
		t.Errorf("expected response format default to request, got %s", info.ResponseFormat)
	}

	if info.FileScheme != "data" {
		t.Errorf("unexpected file scheme: %s", info.FileScheme)
	}

	if info.DefaultResolution != 512 {
		t.Errorf("unexpected resolution: %d", info.DefaultResolution)
	}
}

func TestRegisterEngine(t *testing.T) {
	format := ApiFormat("unit-format")
	engine := Engine{}

	engineMu.Lock()
	prev, had := engineRegistry[format]
	if had {
		delete(engineRegistry, format)
	}
	engineMu.Unlock()

	t.Cleanup(func() {
		engineMu.Lock()
		if had {
			engineRegistry[format] = prev
		} else {
			delete(engineRegistry, format)
		}
		engineMu.Unlock()
	})

	RegisterEngine(format, engine)
	got, ok := EngineFor(format)
	if !ok {
		t.Fatalf("expected engine for %s", format)
	}

	if got != engine {
		t.Errorf("unexpected engine value: %#v", got)
	}
}
