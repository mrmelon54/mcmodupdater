package mcmodupdater

import (
	"github.com/MrMelon54/mcmodupdater/config"
	"github.com/MrMelon54/mcmodupdater/develop"
	"github.com/MrMelon54/mcmodupdater/develop/dev"
	"github.com/MrMelon54/mcmodupdater/paths"
	"io/fs"
	"os"
	"path"
)

type McModUpdater struct {
	cwd       string
	cache     string
	platforms map[develop.DevPlatform]develop.Develop
	platArch  *dev.Architectury
}

type VersionUpdateList []VersionUpdateItem

func (v VersionUpdateList) ChangeToLatest() map[develop.PropVersion]string {
	a := make(map[develop.PropVersion]string)
	for _, i := range v {
		if i.Latest != "" {
			a[i.Property] = i.Latest
		} else {
			a[i.Property] = i.Current
		}
	}
	return a
}

type VersionUpdateItem struct {
	Property        develop.PropVersion
	Current, Latest string
}

func NewMcModUpdater(conf *config.Config) (*McModUpdater, error) {
	var cache, platCache string
	if conf.Cache {
		cache = paths.UserCacheDir()
		platCache = path.Join(cache, "platforms")
		err := os.MkdirAll(cache, fs.ModePerm)
		if err != nil {
			return nil, err
		}
	}

	platMap := make([]string, len(dev.DevelopPlatformsFactory))
	plat := make(map[develop.DevPlatform]develop.Develop)
	for i, j := range dev.DevelopPlatformsFactory {
		d := j(conf.Develop, platCache)
		p := d.Platform()
		plat[p] = d
		platMap[i] = p.Name
	}

	return &McModUpdater{
		cache:     cache,
		platforms: plat,
		platArch:  dev.ForArchitectury(conf.Develop, platCache).(*dev.Architectury),
	}, nil
}

func (m *McModUpdater) PlatArch() *dev.Architectury                        { return m.platArch }
func (m *McModUpdater) Platforms() map[develop.DevPlatform]develop.Develop { return m.platforms }

func (m *McModUpdater) detectPlatformFromTree(tree fs.StatFS) (develop.Develop, bool) {
	for _, i := range m.platforms {
		if i.ValidTree(tree) {
			return i, true
		}
	}
	return nil, false
}

func (m *McModUpdater) LoadTree(tree fs.StatFS) (*develop.PlatformVersions, error) {
	useArch := m.platArch.ValidTree(tree)

	var platform develop.Develop

	if useArch {
		platform = m.platArch
		m.platArch.SubPlatforms = make(map[develop.DevPlatform]develop.Develop, len(m.platforms))
		for k, v := range m.platforms {
			m.platArch.SubPlatforms[k] = v
		}
	} else {
		for _, i := range m.platforms {
			if i.ValidTree(tree) {
				platform = i
			}
		}
	}

	versions, err := platform.ReadVersionFile(tree)
	if err != nil {
		return nil, err
	}

	return &develop.PlatformVersions{
		Platform: platform,
		Versions: versions,
	}, nil
}
