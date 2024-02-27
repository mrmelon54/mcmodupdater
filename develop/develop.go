package develop

import (
	"io/fs"
)

type Develop interface {
	Platform() DevPlatform
	FetchCalls() []DevFetch
	ValidTree(tree fs.FS) bool
	ReadVersionFile(tree fs.FS, name string) (map[PropVersion]string, error)
	LatestVersion(prop PropVersion, mcVersion string) (string, bool)
	LatestLoaderVersion(mcVersion string) (string, error)
}

type DevPlatform struct {
	Name string
	Sub  string
}

type DevFetch struct {
	Name string
	Call func() error
}

type PlatformVersions struct {
	Platform Develop
	Versions map[PropVersion]string
}
