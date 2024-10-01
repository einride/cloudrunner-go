package cloudslog

import (
	"log/slog"
	"runtime/debug"
)

func newBuildInfoValue(bi *debug.BuildInfo) buildInfoValue {
	return buildInfoValue{BuildInfo: bi}
}

type buildInfoValue struct {
	*debug.BuildInfo
}

func (v buildInfoValue) LogValue() slog.Value {
	buildSettings := make([]any, 0, len(v.Settings))
	for _, setting := range v.BuildInfo.Settings {
		buildSettings = append(buildSettings, slog.String(setting.Key, setting.Value))
	}
	return slog.GroupValue(
		slog.String("mainPath", v.BuildInfo.Main.Path),
		slog.String("goVersion", v.BuildInfo.GoVersion),
		slog.Group("buildSettings", buildSettings...),
	)
}
