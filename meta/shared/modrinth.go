package shared

import (
	"github.com/Masterminds/semver/v3"
	"slices"
)

type ModrinthVersionList []ModrinthVersion

func (m ModrinthVersionList) FilterGameVersions(gameVersion string) ModrinthVersionList {
	a := make(ModrinthVersionList, 0)
	for _, i := range m {
		if slices.Contains(i.GameVersions, gameVersion) {
			a = append(a, i)
		}
	}
	return a
}

func (m ModrinthVersionList) GetLatest() string {
	return slices.MaxFunc(m, func(a, b ModrinthVersion) int {
		return a.VersionNumber.Compare(b.VersionNumber)
	}).GetVersion()
}

type ModrinthVersion struct {
	GameVersions  []string        `json:"game_versions"`
	Loaders       []string        `json:"loaders"`
	VersionNumber *semver.Version `json:"version_number"`
}

func (v ModrinthVersion) GetVersion() string {
	metadata, _ := v.VersionNumber.SetMetadata("")
	return metadata.String()
}
