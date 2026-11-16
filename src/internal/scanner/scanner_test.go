package scanner

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"biblioteca-ssd/internal/db"
)

func TestClassifyExtension(t *testing.T) {
	cases := map[string]string{".JPG": "photo", ".mkv": "video", ".MP3": "audio", ".pdf": "pdf", ".EPUB": "ebook", ".zip": "other"}
	for ext, want := range cases {
		if got := ClassifyExtension(ext); got != want {
			t.Errorf("ClassifyExtension(%q)=%q, want %q", ext, got, want)
		}
	}
}

// Run with BIBLIOTECA_PERF_TEST=1 to exercise the Phase 1 50k-file acceptance target.
func TestScan50KFilesAndReuseCatalog(t *testing.T) {
	if os.Getenv("BIBLIOTECA_PERF_TEST") != "1" {
		t.Skip("set BIBLIOTECA_PERF_TEST=1 to run the 50k-file acceptance check")
	}
	root := t.TempDir()
	for i := 0; i < 50000; i++ {
		p := filepath.Join(root, "item-"+strconv.Itoa(i)+".txt")
		if err := os.WriteFile(p, []byte("fixture"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	catalog, err := db.Open(filepath.Join(t.TempDir(), "catalog.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer catalog.Close()
	started := time.Now()
	first, err := Run(context.Background(), catalog, Options{Root: root, Workers: 8}, nil)
	elapsed := time.Since(started)
	if err != nil {
		t.Fatal(err)
	}
	if first.Visited != 50000 || first.Indexed != 50000 || first.Errors != 0 {
		t.Fatalf("50k scan result = %+v", first)
	}
	if elapsed > 10*time.Minute {
		t.Fatalf("50k initial scan took %s, over the 10-minute acceptance limit", elapsed)
	}
	second, err := Run(context.Background(), catalog, Options{Root: root, Workers: 8}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if second.Visited != 50000 || second.Indexed != 0 || second.Errors != 0 {
		t.Fatalf("unchanged rescan result = %+v", second)
	}
	t.Logf("50k files indexed in %s; unchanged reopen visited %d and reindexed %d", elapsed, second.Visited, second.Indexed)
}

func TestPartialHashUsesFileContentAndSize(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "a.bin")
	b := filepath.Join(dir, "b.bin")
	if err := os.WriteFile(a, []byte("same length A"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(b, []byte("same length B"), 0600); err != nil {
		t.Fatal(err)
	}
	ha, err := PartialHash(a, 13)
	if err != nil {
		t.Fatal(err)
	}
	hb, err := PartialHash(b, 13)
	if err != nil {
		t.Fatal(err)
	}
	if ha == hb {
		t.Fatal("different file contents produced the same partial hash")
	}
}

func TestRunIncrementalExclusionsAndMissingFiles(t *testing.T) {
	root := t.TempDir()
	catalogPath := filepath.Join(t.TempDir(), "catalog.db")
	catalog, err := db.Open(catalogPath)
	if err != nil {
		t.Fatal(err)
	}
	defer catalog.Close()
	if err := os.MkdirAll(filepath.Join(root, "lesson"), 0700); err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(root, "lesson", "Aula 01.MP4")
	if err := os.WriteFile(file, []byte("video bytes"), 0600); err != nil {
		t.Fatal(err)
	}
	ignored := filepath.Join(root, "System Volume Information")
	if err := os.MkdirAll(ignored, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(ignored, "hidden.pdf"), []byte("ignore"), 0600); err != nil {
		t.Fatal(err)
	}
	options := Options{Root: root, Workers: 2}
	first, err := Run(context.Background(), catalog, options, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !first.Done || first.Indexed != 1 || first.Errors != 0 {
		t.Fatalf("first scan = %+v", first)
	}
	second, err := Run(context.Background(), catalog, options, nil)
	if err != nil {
		t.Fatal(err)
	}
	if second.Visited != 1 || second.Indexed != 0 {
		t.Fatalf("unchanged rescan = %+v; expected no reprocessing", second)
	}
	items, total, err := List(catalog, "video", "lesson", 1, 50)
	if err != nil {
		t.Fatal(err)
	}
	if total != 1 || len(items) != 1 || items[0].Type != "video" || items[0].Path != "lesson/Aula 01.MP4" {
		t.Fatalf("catalog listing = %#v, total %d", items, total)
	}
	if err := os.Remove(file); err != nil {
		t.Fatal(err)
	}
	if _, err := Run(context.Background(), catalog, options, nil); err != nil {
		t.Fatal(err)
	}
	var removed int
	if err := catalog.QueryRow(`SELECT removed FROM items WHERE path='lesson/Aula 01.MP4'`).Scan(&removed); err != nil {
		t.Fatal(err)
	}
	if removed != 1 {
		t.Fatalf("removed = %d, want 1", removed)
	}
	var excluded int
	if err := catalog.QueryRow(`SELECT count(*) FROM items WHERE name='hidden.pdf'`).Scan(&excluded); err != nil {
		t.Fatal(err)
	}
	if excluded != 0 {
		t.Fatalf("system files indexed: %d", excluded)
	}
}

func TestListRejectsTraversal(t *testing.T) {
	catalog, err := db.Open(filepath.Join(t.TempDir(), "catalog.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer catalog.Close()
	if _, _, err := List(catalog, "", "../outside", 1, 50); err == nil {
		t.Fatal("expected traversal folder to be rejected")
	}
}
