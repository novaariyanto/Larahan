package main

import (
	"embed"
	"log"

	"github.com/larahan/larahan/backend/apache"
	"github.com/larahan/larahan/backend/config"
	"github.com/larahan/larahan/backend/installer"
	"github.com/larahan/larahan/backend/logger"
	"github.com/larahan/larahan/backend/mysql"
	"github.com/larahan/larahan/backend/php"
	"github.com/larahan/larahan/backend/phpmyadmin"
	"github.com/larahan/larahan/backend/services"
	"github.com/larahan/larahan/backend/storage"
	"github.com/wailsapp/wails/v3/pkg/application"
)

//go:embed all:frontend/dist
var assets embed.FS

func init() {
	installer.RegisterEvents()
}

func main() {
	installPath := config.DefaultInstallPath
	paths := config.ResolvePaths(installPath)

	if err := config.EnsureDirectories(paths); err != nil {
		log.Fatalf("bootstrap directories: %v", err)
	}

	appLogger, err := logger.New(paths.Logs)
	if err != nil {
		log.Fatalf("logger init: %v", err)
	}
	defer appLogger.Close()
	appLogger.Info("Larahan starting")

	settingsMgr := config.NewManager(installPath)
	settings, err := settingsMgr.Load()
	if err != nil {
		appLogger.Error("failed to load settings: %v", err)
		log.Fatalf("settings: %v", err)
	}
	appLogger.Info("settings loaded (first_run=%v, path=%s)", settings.FirstRun, settings.InstallPath)

	if settings.InstallPath != "" && settings.InstallPath != installPath {
		installPath = settings.InstallPath
		paths = config.ResolvePaths(installPath)
		if err := config.EnsureDirectories(paths); err != nil {
			log.Fatalf("bootstrap directories: %v", err)
		}
	}

	store, err := storage.Open(paths.Config)
	if err != nil {
		appLogger.Error("sqlite open failed: %v", err)
		log.Fatalf("storage: %v", err)
	}
	defer store.Close()

	if err := store.SyncSettings(settingsMgr.Get()); err != nil {
		appLogger.Error("sync settings to sqlite: %v", err)
	} else {
		appLogger.Info("sqlite ready at %s\\larahan.db", paths.Config)
	}

	pipeline := installer.NewPipeline(paths, store, installer.WailsEmitter{})

	apacheMgr := apache.NewManager(installPath, pipeline, store, settingsMgr)
	phpMgr := php.NewManager(installPath, pipeline, store, settingsMgr, settings.ActivePHP)
	phpMgr.SetApacheBridge(apacheMgr)
	apacheMgr.SetBeforeStart(func() error {
		return phpMgr.EnsureApacheHandler()
	})
	mysqlMgr := mysql.NewManager(installPath, pipeline, store, settingsMgr)
	pmaMgr := phpmyadmin.NewManager(installPath, pipeline, store, settingsMgr)
	pmaMgr.SetApacheBridge(apacheMgr)

	app := application.New(application.Options{
		Name:        "Larahan",
		Description: "PHP local development environment manager",
		Services: []application.Service{
			application.NewService(services.NewDashboardService(apacheMgr, phpMgr, mysqlMgr, pmaMgr)),
			application.NewService(services.NewApacheService(apacheMgr)),
			application.NewService(services.NewPHPService(phpMgr)),
			application.NewService(services.NewLibraryService(phpMgr)),
			application.NewService(services.NewMySQLService(mysqlMgr)),
			application.NewService(services.NewPhpMyAdminService(pmaMgr)),
			application.NewService(services.NewSettingsService(settingsMgr, store, apacheMgr, mysqlMgr, pmaMgr)),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
	})

	app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:            "Larahan " + config.AppVersion,
		Width:            1180,
		Height:           740,
		MinWidth:         960,
		MinHeight:        640,
		BackgroundColour: application.NewRGB(248, 250, 252),
		URL:              "/",
	})

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
