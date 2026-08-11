package models

// InstallStage identifies a step in the install pipeline.
type InstallStage string

const (
	StageDownload  InstallStage = "download"
	StageVerify    InstallStage = "verify"
	StageExtract   InstallStage = "extract"
	StageConfigure InstallStage = "configure"
	StageDone      InstallStage = "done"
)

// DownloadProgress is emitted during package download.
type DownloadProgress struct {
	Type     string  `json:"type"`
	Version  string  `json:"version"`
	Percent  float64 `json:"percent"`
	Bytes    int64   `json:"bytes"`
	Total    int64   `json:"total"`
	Cached   bool    `json:"cached"`
	Filename string  `json:"filename"`
}

// InstallStageEvent is emitted when the install pipeline changes stage.
type InstallStageEvent struct {
	Type    string       `json:"type"`
	Version string       `json:"version"`
	Stage   InstallStage `json:"stage"`
	Message string       `json:"message"`
}

// InstallErrorEvent is emitted when the install pipeline fails.
type InstallErrorEvent struct {
	Type    string `json:"type"`
	Version string `json:"version"`
	Message string `json:"message"`
}
