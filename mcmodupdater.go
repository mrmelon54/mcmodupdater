package mcmodupdater

import (
	"bufio"
	"fmt"
	"github.com/MrMelon54/mcmodupdater/config"
	"github.com/MrMelon54/mcmodupdater/develop"
	"github.com/MrMelon54/mcmodupdater/develop/dev"
	"github.com/MrMelon54/mcmodupdater/paths"
	"github.com/MrMelon54/mcmodupdater/utils"
	"github.com/magiconair/properties"
	"io"
	"io/fs"
	"os"
	"strings"
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
		platCache = utils.PathJoin(cache, "platforms")
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

	if platform == nil {
		return nil, fmt.Errorf("cannot find valid platform")
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

func (m *McModUpdater) VersionUpdateList(info *develop.PlatformVersions) VersionUpdateList {
	v := make(VersionUpdateList, 0, 11)
	v = m.useIfExists(v, info, develop.ModVersion)
	v = m.useIfExists(v, info, develop.MinecraftVersion)
	v = m.useIfExistsUpdate(v, info, develop.ArchitecturyVersion)
	v = m.useIfExistsUpdate(v, info, develop.FabricLoaderVersion)
	v = m.useIfExistsUpdate(v, info, develop.FabricApiVersion)
	v = m.useIfExistsUpdate(v, info, develop.YarnMappingsVersion)
	v = m.useIfExistsUpdate(v, info, develop.ForgeVersion)
	v = m.useIfExistsUpdate(v, info, develop.ForgeMappingsVersion)
	v = m.useIfExistsUpdate(v, info, develop.QuiltLoaderVersion)
	v = m.useIfExistsUpdate(v, info, develop.QuiltFabricApiVersion)
	v = m.useIfExistsUpdate(v, info, develop.QuiltMappingsVersion)
	return v
}

func (m *McModUpdater) useIfExists(v VersionUpdateList, branch *develop.PlatformVersions, k develop.PropVersion) VersionUpdateList {
	if a, ok := branch.Versions[k]; ok {
		v = append(v, VersionUpdateItem{k, a, ""})
	}
	return v
}

func (m *McModUpdater) useIfExistsUpdate(v VersionUpdateList, branch *develop.PlatformVersions, k develop.PropVersion) VersionUpdateList {
	if a, ok := branch.Versions[k]; ok {
		if l, ok := branch.Platform.LatestVersion(k, branch.Versions[develop.MinecraftVersion]); ok {
			if a != l {
				v = append(v, VersionUpdateItem{k, a, l})
				return v
			}
		}
		v = append(v, VersionUpdateItem{k, a, ""})
		return v
	}
	return v
}

func (m *McModUpdater) UpdateToVersion(out io.StringWriter, tree fs.StatFS, ver map[develop.PropVersion]string) error {
	gProp, err := tree.Open("gradle.properties")
	if err != nil {
		return err
	}
	//goland:noinspection GoUnhandledErrorResult
	defer gProp.Close()

	scanner := bufio.NewScanner(gProp)
	for scanner.Scan() {
		t := scanner.Text()
		if strings.TrimSpace(t) != "" {
			if oneProp, err := properties.LoadString(t); err == nil {
				k := oneProp.Keys()
				if len(k) == 1 {
					if p, ok := develop.PropVersionFromKey(k[0]); ok {
						if p2, ok := ver[p]; ok {
							_, _ = out.WriteString(p.Key() + "=" + p2 + "\n")
							continue
						}
					}
				}
			}
		}
		_, _ = out.WriteString(t + "\n")
		continue
	}
	return nil
}
