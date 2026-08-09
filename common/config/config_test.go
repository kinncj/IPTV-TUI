package config

import (
	"os"
	"path/filepath"
	"testing"
)

func write(t *testing.T, dir, name, content string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestLoadMerge(t *testing.T) {
	dir := t.TempDir()
	a := write(t, dir, "a.json", `{"theme":"nord","sources":[{"name":"x","url":"http://x"}]}`)
	b := write(t, dir, "b.json", `{"theme":"gruvbox","player":"mpv","sources":[{"name":"y","url":"http://y"}]}`)

	cfg, loaded, err := Load(filepath.Join(dir, "missing.json"), a, b)
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded) != 2 {
		t.Errorf("loaded %v, want 2 files", loaded)
	}
	// later file wins for scalars
	if cfg.Theme != "gruvbox" || cfg.Player != "mpv" {
		t.Errorf("scalar merge wrong: %+v", cfg)
	}
	// sources are appended
	if len(cfg.Sources) != 2 {
		t.Errorf("want 2 merged sources, got %d", len(cfg.Sources))
	}
}

func TestLoadMissingIsNotError(t *testing.T) {
	cfg, loaded, err := Load(filepath.Join(t.TempDir(), "nope.json"))
	if err != nil {
		t.Fatalf("missing config should not error: %v", err)
	}
	if len(loaded) != 0 || len(cfg.Sources) != 0 {
		t.Errorf("expected empty config")
	}
}

func TestLoadMalformed(t *testing.T) {
	dir := t.TempDir()
	bad := write(t, dir, "bad.json", `{not json`)
	if _, _, err := Load(bad); err == nil {
		t.Error("malformed config should error")
	}
}

func TestSourceWithoutURLSkipped(t *testing.T) {
	dir := t.TempDir()
	p := write(t, dir, "c.json", `{"sources":[{"name":"empty"},{"name":"ok","url":"http://ok"}]}`)
	cfg, _, _ := Load(p)
	if len(cfg.Sources) != 1 || cfg.Sources[0].Name != "ok" {
		t.Errorf("URL-less source should be skipped: %+v", cfg.Sources)
	}
}

func TestSaveLoadFileRoundtrip(t *testing.T) {
	p := filepath.Join(t.TempDir(), "iptv.local.json")
	in := Config{Theme: "nord", Sources: []Source{{Name: "a", URL: "http://a"}}}
	if err := Save(p, in); err != nil {
		t.Fatal(err)
	}
	out, err := LoadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if out.Theme != "nord" || len(out.Sources) != 1 || out.Sources[0].URL != "http://a" {
		t.Errorf("roundtrip mismatch: %+v", out)
	}
}

func TestAddRemoveSource(t *testing.T) {
	p := filepath.Join(t.TempDir(), "iptv.local.json")
	if err := AddSource(p, Source{Name: "x", URL: "http://x"}); err != nil {
		t.Fatal(err)
	}
	if err := AddSource(p, Source{Name: "x2", URL: "http://x"}); err != nil { // dup URL
		t.Fatal(err)
	}
	cfg, _ := LoadFile(p)
	if len(cfg.Sources) != 1 {
		t.Fatalf("dup URL should not add twice: %+v", cfg.Sources)
	}
	if err := RemoveSource(p, "http://x"); err != nil {
		t.Fatal(err)
	}
	cfg, _ = LoadFile(p)
	if len(cfg.Sources) != 0 {
		t.Errorf("remove failed: %+v", cfg.Sources)
	}
}

func TestLoadFileMissing(t *testing.T) {
	cfg, err := LoadFile(filepath.Join(t.TempDir(), "none.json"))
	if err != nil || len(cfg.Sources) != 0 {
		t.Errorf("missing file should be empty/no-error: %+v %v", cfg, err)
	}
}

func TestAPIMerge(t *testing.T) {
	dir := t.TempDir()
	a := write(t, dir, "a.json", `{"api":true}`)
	b := write(t, dir, "b.json", `{"theme":"nord"}`) // api not set here must not clear it
	cfg, _, err := Load(a, b)
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.API {
		t.Error("api=true should survive merge with a later file that omits it")
	}
}
