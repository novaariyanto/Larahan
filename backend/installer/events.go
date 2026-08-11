package installer

import (
	"github.com/larahan/larahan/backend/models"
	"github.com/wailsapp/wails/v3/pkg/application"
)

const (
	EventDownloadProgress = "download:progress"
	EventInstallStage     = "install:stage"
	EventInstallError     = "install:error"
)

// WailsEmitter emits install events through the Wails application bus.
type WailsEmitter struct{}

// Progress implements Emitter.
func (WailsEmitter) Progress(p models.DownloadProgress) {
	if app := application.Get(); app != nil {
		app.Event.Emit(EventDownloadProgress, p)
	}
}

// Stage implements Emitter.
func (WailsEmitter) Stage(s models.InstallStageEvent) {
	if app := application.Get(); app != nil {
		app.Event.Emit(EventInstallStage, s)
	}
}

// Error implements Emitter.
func (WailsEmitter) Error(e models.InstallErrorEvent) {
	if app := application.Get(); app != nil {
		app.Event.Emit(EventInstallError, e)
	}
}

// RegisterEvents registers typed install events for binding generation.
func RegisterEvents() {
	application.RegisterEvent[models.DownloadProgress](EventDownloadProgress)
	application.RegisterEvent[models.InstallStageEvent](EventInstallStage)
	application.RegisterEvent[models.InstallErrorEvent](EventInstallError)
}
