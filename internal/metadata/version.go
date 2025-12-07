package metadata

import "runtime"

var (
	Version   = "dev"
	CommitSHA = "none"
	BuildTime = "unknown"
)

type BuildInfo struct {
	Version   string `json:"version"`
	CommitSHA string `json:"commit_sha"`
	BuildTime string `json:"build_time"`
	GoVersion string `json:"go_version"`
}

func GetBuildInfo() BuildInfo {
	return BuildInfo{
		Version:   Version,
		CommitSHA: CommitSHA,
		BuildTime: BuildTime,
		GoVersion: runtime.Version(),
	}
}
