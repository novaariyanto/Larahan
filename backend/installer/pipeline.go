package installer

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/larahan/larahan/backend/downloader"
	"github.com/larahan/larahan/backend/models"
	"github.com/larahan/larahan/backend/storage"
)

// Emitter reports install pipeline progress to interested listeners (e.g. Wails events).
type Emitter interface {
	Progress(models.DownloadProgress)
	Stage(models.InstallStageEvent)
	Error(models.InstallErrorEvent)
}

// NopEmitter discards all events.
type NopEmitter struct{}

func (NopEmitter) Progress(models.DownloadProgress)     {}
func (NopEmitter) Stage(models.InstallStageEvent)       {}
func (NopEmitter) Error(models.InstallErrorEvent)       {}

// Pipeline downloads, verifies, and extracts component packages.
type Pipeline struct {
	client  *downloader.Client
	store   *storage.Store
	paths   models.Paths
	emitter Emitter
}

// NewPipeline constructs an install pipeline.
func NewPipeline(paths models.Paths, store *storage.Store, emitter Emitter) *Pipeline {
	if emitter == nil {
		emitter = NopEmitter{}
	}
	cache := downloader.NewCache(paths.Downloads)
	return &Pipeline{
		client:  downloader.NewClient(cache),
		store:   store,
		paths:   paths,
		emitter: emitter,
	}
}

// SetEmitter replaces the event emitter (e.g. after Wails app is ready).
func (p *Pipeline) SetEmitter(emitter Emitter) {
	if emitter == nil {
		p.emitter = NopEmitter{}
		return
	}
	p.emitter = emitter
}

// TargetDir resolves the extract destination for a component.
func (p *Pipeline) TargetDir(componentType models.ComponentType, version string) string {
	switch componentType {
	case models.TypePHP:
		return filepath.Join(p.paths.PHP, version)
	case models.TypeApache:
		return p.paths.Apache
	case models.TypeMySQL:
		return p.paths.MySQL
	case models.TypePhpMyAdmin:
		return p.paths.PhpMyAdmin
	default:
		return filepath.Join(p.paths.Temp, string(componentType), version)
	}
}

// ConfigureFunc runs after extract and before the final "done" stage.
type ConfigureFunc func(destDir string) error

// Install downloads (with cache), verifies, extracts, optionally configures, and records metadata.
func (p *Pipeline) Install(ctx context.Context, componentType models.ComponentType, version string, configure ...ConfigureFunc) error {
	spec, ok := downloader.Lookup(componentType, version)
	if !ok {
		err := fmt.Errorf("paket tidak dikenali: %s@%s", componentType, version)
		p.emitter.Error(models.InstallErrorEvent{Type: string(componentType), Version: version, Message: err.Error()})
		return err
	}

	localPath, err := p.fetch(ctx, spec)
	if err != nil {
		p.emitter.Error(models.InstallErrorEvent{Type: string(componentType), Version: version, Message: err.Error()})
		return err
	}

	p.emitter.Stage(models.InstallStageEvent{
		Type: string(componentType), Version: version, Stage: models.StageVerify, Message: "Memverifikasi paket",
	})
	if err := Verify(localPath, spec.SHA256, spec.SizeHint); err != nil {
		p.emitter.Error(models.InstallErrorEvent{Type: string(componentType), Version: version, Message: err.Error()})
		return err
	}

	dest := p.TargetDir(componentType, version)
	p.emitter.Stage(models.InstallStageEvent{
		Type: string(componentType), Version: version, Stage: models.StageExtract, Message: "Mengekstrak paket",
	})

	tmpExtract := filepath.Join(p.paths.Temp, fmt.Sprintf("extract-%s-%s", componentType, version))
	_ = os.RemoveAll(tmpExtract)
	if err := ExtractZip(localPath, tmpExtract, true); err != nil {
		p.emitter.Error(models.InstallErrorEvent{Type: string(componentType), Version: version, Message: err.Error()})
		return err
	}

	if err := os.RemoveAll(dest); err != nil && !os.IsNotExist(err) {
		p.emitter.Error(models.InstallErrorEvent{Type: string(componentType), Version: version, Message: err.Error()})
		return err
	}
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return err
	}
	if err := moveDir(tmpExtract, dest); err != nil {
		p.emitter.Error(models.InstallErrorEvent{Type: string(componentType), Version: version, Message: err.Error()})
		return fmt.Errorf("pindahkan hasil extract: %w", err)
	}

	for _, fn := range configure {
		if fn == nil {
			continue
		}
		p.emitter.Stage(models.InstallStageEvent{
			Type: string(componentType), Version: version, Stage: models.StageConfigure, Message: "Mengonfigurasi",
		})
		if err := fn(dest); err != nil {
			p.emitter.Error(models.InstallErrorEvent{Type: string(componentType), Version: version, Message: err.Error()})
			return err
		}
	}

	if p.store != nil {
		_ = p.store.AddDownloadHistory(storage.DownloadRecord{
			Type:      string(componentType),
			Version:   version,
			Filename:  spec.Filename,
			Checksum:  spec.SHA256,
			SourceURL: spec.URL,
			LocalPath: localPath,
			Status:    "completed",
		})
		if err := p.store.UpsertPackage(string(componentType), version, dest, "installed"); err != nil {
			p.emitter.Error(models.InstallErrorEvent{Type: string(componentType), Version: version, Message: err.Error()})
			return err
		}
	}

	p.emitter.Stage(models.InstallStageEvent{
		Type: string(componentType), Version: version, Stage: models.StageDone, Message: "Instalasi selesai",
	})
	return nil
}

func (p *Pipeline) fetch(ctx context.Context, spec downloader.PackageSpec) (string, error) {
	p.emitter.Stage(models.InstallStageEvent{
		Type: string(spec.Type), Version: spec.Version, Stage: models.StageDownload, Message: "Mengunduh paket",
	})

	result, err := p.client.Ensure(ctx, spec, func(written, total int64) {
		percent := 0.0
		if total > 0 {
			percent = (float64(written) / float64(total)) * 100
			if percent > 100 {
				percent = 100
			}
		}
		p.emitter.Progress(models.DownloadProgress{
			Type:     string(spec.Type),
			Version:  spec.Version,
			Percent:  percent,
			Bytes:    written,
			Total:    total,
			Cached:   false,
			Filename: spec.Filename,
		})
	})
	if err != nil {
		return "", err
	}

	p.emitter.Progress(models.DownloadProgress{
		Type:     string(spec.Type),
		Version:  spec.Version,
		Percent:  100,
		Bytes:    result.Bytes,
		Total:    result.Bytes,
		Cached:   result.Cached,
		Filename: spec.Filename,
	})
	if result.Cached {
		p.emitter.Stage(models.InstallStageEvent{
			Type: string(spec.Type), Version: spec.Version, Stage: models.StageDownload,
			Message: "Menggunakan cache lokal",
		})
	}
	return result.LocalPath, nil
}

// moveDir renames src to dest, falling back to recursive copy on Windows
// when rename fails (e.g. destination remnants or cross-device links).
func moveDir(src, dest string) error {
	if err := os.Rename(src, dest); err == nil {
		return nil
	}
	if err := copyDir(src, dest); err != nil {
		_ = os.RemoveAll(dest)
		return err
	}
	return os.RemoveAll(src)
}

func copyDir(src, dest string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dest, info.Mode()); err != nil {
		return err
	}
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, e := range entries {
		from := filepath.Join(src, e.Name())
		to := filepath.Join(dest, e.Name())
		if e.IsDir() {
			if err := copyDir(from, to); err != nil {
				return err
			}
			continue
		}
		if err := copyFile(from, to); err != nil {
			return err
		}
	}
	return nil
}

func copyFile(src, dest string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	info, err := in.Stat()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return err
	}
	out, err := os.OpenFile(dest, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode())
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}
