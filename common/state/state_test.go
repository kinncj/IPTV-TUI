package state

import (
	"path/filepath"
	"testing"
)

func TestFavoritesToggleAndPersist(t *testing.T) {
	p := filepath.Join(t.TempDir(), "state.json")
	s := Load(p)
	e := Entry{Name: "A", URL: "u1", Group: "G"}

	if s.IsFavorite("u1") {
		t.Fatal("should start un-favorited")
	}
	if !s.ToggleFavorite(e) {
		t.Fatal("first toggle should favorite")
	}
	if !s.IsFavorite("u1") {
		t.Fatal("should be favorite after toggle")
	}
	if s.ToggleFavorite(e) {
		t.Fatal("second toggle should un-favorite")
	}

	// persist a favorite and reload
	s.ToggleFavorite(e)
	if err := s.Save(); err != nil {
		t.Fatal(err)
	}
	s2 := Load(p)
	if !s2.IsFavorite("u1") {
		t.Error("favorite did not persist across reload")
	}
}

func TestRecentDedupOrderAndCap(t *testing.T) {
	s := Load("")
	s.PushRecent(Entry{URL: "u1"})
	s.PushRecent(Entry{URL: "u2"})
	s.PushRecent(Entry{URL: "u1"}) // moves u1 to front, no dup

	_, rec := s.Snapshot()
	if len(rec) != 2 || rec[0].URL != "u1" {
		t.Fatalf("recent dedup/order wrong: %+v", rec)
	}

	for i := 0; i < maxRecent+10; i++ {
		s.PushRecent(Entry{URL: string(rune('a'+i%26)) + itoa(i)})
	}
	_, rec = s.Snapshot()
	if len(rec) != maxRecent {
		t.Errorf("recent not capped: got %d, want %d", len(rec), maxRecent)
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b []byte
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	return string(b)
}
