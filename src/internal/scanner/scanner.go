// Package scanner indexes a local library without following links or leaving its root.
package scanner

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

type Item struct {
	ID          int64     `json:"id"`
	Path        string    `json:"path"`
	Name        string    `json:"name"`
	Extension   string    `json:"extension"`
	Type        string    `json:"type"`
	Size        int64     `json:"size"`
	ModifiedAt  time.Time `json:"modifiedAt"`
	PartialHash string    `json:"partialHash"`
	Removed     bool      `json:"removed"`
}

type Progress struct {
	Visited int    `json:"visited"`
	Indexed int    `json:"indexed"`
	Errors  int    `json:"errors"`
	Done    bool   `json:"done"`
	Message string `json:"message,omitempty"`
}
type Options struct {
	Root, AppDir string
	Excluded     []string
	Workers      int
}

var extensions = map[string]string{
	".jpg": "photo", ".jpeg": "photo", ".png": "photo", ".gif": "photo", ".webp": "photo", ".bmp": "photo", ".tif": "photo", ".tiff": "photo", ".heic": "photo", ".heif": "photo",
	".mp4": "video", ".mkv": "video", ".mov": "video", ".avi": "video", ".webm": "video", ".m4v": "video", ".wmv": "video", ".mpeg": "video", ".mpg": "video",
	".mp3": "audio", ".wav": "audio", ".flac": "audio", ".aac": "audio", ".m4a": "audio", ".ogg": "audio", ".opus": "audio", ".wma": "audio",
	".pdf": "pdf", ".epub": "ebook", ".mobi": "ebook", ".azw": "ebook", ".azw3": "ebook", ".fb2": "ebook",
}

func ClassifyExtension(extension string) string {
	if value, ok := extensions[strings.ToLower(extension)]; ok {
		return value
	}
	return "other"
}

// PartialHash hashes file size, its first 64 KiB and its last 64 KiB.
func PartialHash(path string, size int64) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	var sizeBytes [8]byte
	for i := uint(0); i < 8; i++ {
		sizeBytes[i] = byte(uint64(size) >> (i * 8))
	}
	_, _ = h.Write(sizeBytes[:])
	const sample = 64 * 1024
	first := make([]byte, sample)
	n, err := io.ReadFull(f, first)
	if err != nil && !errors.Is(err, io.EOF) && !errors.Is(err, io.ErrUnexpectedEOF) {
		return "", err
	}
	_, _ = h.Write(first[:n])
	if size > sample {
		offset := size - sample
		if offset < int64(n) {
			offset = int64(n)
		}
		if _, err := f.Seek(offset, io.SeekStart); err != nil {
			return "", err
		}
		last := make([]byte, sample)
		n, err = io.ReadFull(f, last)
		if err != nil && !errors.Is(err, io.EOF) && !errors.Is(err, io.ErrUnexpectedEOF) {
			return "", err
		}
		_, _ = h.Write(last[:n])
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// Run scans recursively with bounded workers. Missing rows are marked removed only
// after traversal and processing finish without errors or cancellation.
func Run(ctx context.Context, database *sql.DB, opts Options, report func(Progress)) (Progress, error) {
	root, err := filepath.Abs(opts.Root)
	if err != nil {
		return Progress{}, err
	}
	root = filepath.Clean(root)
	if opts.Workers <= 0 {
		opts.Workers = runtime.NumCPU()
		if opts.Workers > 8 {
			opts.Workers = 8
		}
	}
	ignored := make([]string, 0, len(opts.Excluded)+3)
	for _, p := range append(append([]string{}, opts.Excluded...), "$RECYCLE.BIN", "System Volume Information") {
		if p != "" {
			if filepath.IsAbs(p) {
				ignored = append(ignored, filepath.Clean(p))
			} else {
				ignored = append(ignored, filepath.Clean(filepath.Join(root, p)))
			}
		}
	}
	if opts.AppDir != "" {
		app, e := filepath.Abs(opts.AppDir)
		if e == nil && !samePath(app, root) {
			ignored = append(ignored, filepath.Clean(app))
		}
	}
	contains := func(candidate string) bool {
		candidate = filepath.Clean(candidate)
		for _, p := range ignored {
			rel, e := filepath.Rel(p, candidate)
			if e == nil && (rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)))) {
				return true
			}
		}
		return false
	}
	var runID int64
	if err := database.QueryRowContext(ctx, "INSERT INTO scan_runs DEFAULT VALUES RETURNING id").Scan(&runID); err != nil {
		return Progress{}, fmt.Errorf("start scan: %w", err)
	}
	if _, err := database.ExecContext(ctx, `DELETE FROM scan_seen WHERE run_id < ?`, runID); err != nil {
		return Progress{}, fmt.Errorf("clean previous scan state: %w", err)
	}
	tx, err := database.BeginTx(ctx, nil)
	if err != nil {
		return Progress{}, fmt.Errorf("begin catalog update: %w", err)
	}
	type file struct {
		path, rel string
		info      os.FileInfo
	}
	jobs := make(chan file, opts.Workers*2)
	var wg sync.WaitGroup
	var mu sync.Mutex
	state := Progress{}
	update := func(delta Progress) {
		mu.Lock()
		state.Visited += delta.Visited
		state.Indexed += delta.Indexed
		state.Errors += delta.Errors
		snapshot := state
		mu.Unlock()
		if report != nil {
			report(snapshot)
		}
	}
	for i := 0; i < opts.Workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for f := range jobs {
				if ctx.Err() != nil {
					update(Progress{Errors: 1})
					continue
				}
				_, e := tx.ExecContext(ctx, `INSERT OR IGNORE INTO scan_seen(run_id,path) VALUES(?,?)`, runID, f.rel)
				var oldSize, oldMtime int64
				var oldRemoved int
				lookupErr := tx.QueryRowContext(ctx, `SELECT size,mtime_ns,removed FROM items WHERE path=?`, f.rel).Scan(&oldSize, &oldMtime, &oldRemoved)
				unchanged := lookupErr == nil && oldSize == f.info.Size() && oldMtime == f.info.ModTime().UnixNano() && oldRemoved == 0
				var hash string
				if e == nil && !unchanged {
					hash, e = PartialHash(f.path, f.info.Size())
				}
				if e == nil && !unchanged {
					ext := strings.ToLower(filepath.Ext(f.path))
					_, e = tx.ExecContext(ctx, `INSERT INTO items(path,name,extension,type,size,mtime_ns,partial_hash,removed,indexed_at) VALUES(?,?,?,?,?,?,?,0,CURRENT_TIMESTAMP) ON CONFLICT(path) DO UPDATE SET name=excluded.name,extension=excluded.extension,type=excluded.type,size=excluded.size,mtime_ns=excluded.mtime_ns,partial_hash=excluded.partial_hash,removed=0,indexed_at=CURRENT_TIMESTAMP`, f.rel, filepath.Base(f.path), ext, ClassifyExtension(ext), f.info.Size(), f.info.ModTime().UnixNano(), hash)
				}
				if e != nil {
					update(Progress{Visited: 1, Errors: 1})
					continue
				}
				indexed := 0
				if !unchanged {
					indexed = 1
				}
				update(Progress{Visited: 1, Indexed: indexed})
			}
		}()
	}
	walkErr := filepath.WalkDir(root, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			update(Progress{Errors: 1})
			return nil
		}
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if path != root && contains(path) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if d.Type()&os.ModeSymlink != 0 {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		info, e := d.Info()
		if e != nil {
			update(Progress{Errors: 1})
			return nil
		}
		rel, e := filepath.Rel(root, path)
		if e != nil {
			update(Progress{Errors: 1})
			return nil
		}
		select {
		case jobs <- file{path, filepath.ToSlash(rel), info}:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	})
	close(jobs)
	wg.Wait()
	commitErr := tx.Commit()
	mu.Lock()
	final := state
	mu.Unlock()
	complete := walkErr == nil && ctx.Err() == nil && final.Errors == 0 && commitErr == nil
	if commitErr != nil {
		final.Errors++
	}
	if complete {
		if _, err := database.ExecContext(ctx, `UPDATE items SET removed=1 WHERE removed=0 AND NOT EXISTS (SELECT 1 FROM scan_seen s WHERE s.run_id=? AND s.path=items.path)`, runID); err != nil {
			complete = false
			final.Errors++
		}
	}
	_, finishErr := database.ExecContext(context.Background(), `UPDATE scan_runs SET finished_at=CURRENT_TIMESTAMP,visited=?,indexed=?,errors=?,complete=? WHERE id=?`, final.Visited, final.Indexed, final.Errors, boolInt(complete), runID)
	final.Done = true
	final.Message = "Varredura concluída"
	if !complete {
		final.Message = "Varredura incompleta; itens ausentes foram preservados"
	}
	if report != nil {
		report(final)
	}
	if finishErr != nil {
		return final, finishErr
	}
	if commitErr != nil {
		return final, commitErr
	}
	if walkErr != nil && !errors.Is(walkErr, context.Canceled) {
		return final, walkErr
	}
	if ctx.Err() != nil {
		return final, ctx.Err()
	}
	return final, nil
}
func boolInt(v bool) int {
	if v {
		return 1
	}
	return 0
}
func samePath(a, b string) bool {
	a, b = filepath.Clean(a), filepath.Clean(b)
	if runtime.GOOS == "windows" {
		return strings.EqualFold(a, b)
	}
	return a == b
}

func List(database *sql.DB, kind, folder string, page, pageSize int) ([]Item, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 200 {
		pageSize = 50
	}
	where := `removed=0`
	args := []any{}
	if kind != "" && kind != "all" {
		where += ` AND type=?`
		args = append(args, kind)
	}
	if folder != "" {
		clean := filepath.ToSlash(filepath.Clean(folder))
		if clean == "." {
			clean = ""
		}
		if filepath.IsAbs(folder) || strings.HasPrefix(clean, "../") || clean == ".." {
			return nil, 0, fmt.Errorf("invalid folder")
		}
		if clean != "" {
			where += ` AND (path LIKE ? ESCAPE '\')`
			args = append(args, escapeLike(clean)+`/%`)
		}
	}
	var total int
	if err := database.QueryRow(`SELECT count(*) FROM items WHERE `+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	args = append(args, pageSize, (page-1)*pageSize)
	rows, err := database.Query(`SELECT id,path,name,extension,type,size,mtime_ns,partial_hash,removed FROM items WHERE `+where+` ORDER BY path COLLATE NOCASE LIMIT ? OFFSET ?`, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	items := []Item{}
	for rows.Next() {
		var v Item
		var ns int64
		var removed int
		if err := rows.Scan(&v.ID, &v.Path, &v.Name, &v.Extension, &v.Type, &v.Size, &ns, &v.PartialHash, &removed); err != nil {
			return nil, 0, err
		}
		v.ModifiedAt = time.Unix(0, ns)
		v.Removed = removed != 0
		items = append(items, v)
	}
	return items, total, rows.Err()
}
func escapeLike(v string) string {
	return strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(v)
}
