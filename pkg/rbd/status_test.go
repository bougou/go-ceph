package rbd

import (
	"testing"
	"time"
)

func TestParsePersistentCacheState(t *testing.T) {
	raw := `{
		"host":"sceph9",
		"path":"/mnt/nvme0/rbd-pwl.rbd.101e5824ad9a.pool",
		"size":1073741824,
		"mode":"ssd",
		"stats_timestamp":1649597192,
		"present":true,
		"empty":false,
		"clean":false,
		"allocated_bytes":533725184,
		"cached_bytes":525336576,
		"dirty_bytes":354334720,
		"free_bytes":540016640,
		"hits_full":1450,
		"hits_partial":0,
		"misses":924,
		"hit_bytes":201326592,
		"miss_bytes":101711872
	}`

	got := parsePersistentCacheState(raw)
	if got == nil {
		t.Fatal("expected cache state, got nil")
	}
	if got.Host != "sceph9" || got.Mode != "ssd" || got.Size != 1073741824 {
		t.Fatalf("unexpected basic fields: %+v", got)
	}
	if !got.Present || got.Empty || got.Clean {
		t.Fatalf("unexpected flags: present=%v empty=%v clean=%v", got.Present, got.Empty, got.Clean)
	}
	if got.HitsFullPercent() != 61 {
		t.Fatalf("hits_full percent: got %d want 61", got.HitsFullPercent())
	}
	if got.HitBytesPercent() != 66 {
		t.Fatalf("hit_bytes percent: got %d want 66", got.HitBytesPercent())
	}
	wantTS := time.Unix(1649597192, 0)
	if !got.StatsTimestamp.Equal(wantTS) {
		t.Fatalf("stats_timestamp: got %v want %v", got.StatsTimestamp, wantTS)
	}
}

func TestParsePersistentCacheStateInvalid(t *testing.T) {
	if parsePersistentCacheState("") != nil {
		t.Fatal("empty raw should yield nil")
	}
	if parsePersistentCacheState("{") != nil {
		t.Fatal("invalid json should yield nil")
	}
}

func TestPercentage(t *testing.T) {
	if percentage(0, 0) != 0 {
		t.Fatal("0/0 should be 0")
	}
	if percentage(1, 2) != 50 {
		t.Fatal("1/2 should be 50")
	}
}
