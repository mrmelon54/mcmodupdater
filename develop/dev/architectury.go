package dev

import (
	"encoding/json"
	"fmt"
	"github.com/magiconair/properties"
	"github.com/mrmelon54/mcmodupdater/config"
	"github.com/mrmelon54/mcmodupdater/develop"
	"github.com/mrmelon54/mcmodupdater/meta/shared"
	"github.com/mrmelon54/mcmodupdater/utils"
	"io"
	"io/fs"
	"sort"
)

var PlatformArchitectury = develop.DevPlatform{Name: "Architectury"}

type Architectury struct {
	Conf         config.ArchitecturyDevelopConfig
	Meta         *ArchitecturyMeta
	Cache        string
	SubPlatforms map[develop.DevPlatform]develop.Develop
}

type ArchitecturyMeta struct {
	done chan struct{}
	Api  shared.ModrinthVersionList
}

func ForArchitectury(conf config.DevelopConfig, cache string) develop.Develop {
	return &Architectury{
		Conf:  conf.Architectury,
		Meta:  &ArchitecturyMeta{},
		Cache: utils.PathJoin(cache, "architectury"),
	}
}

func (f *Architectury) Platform() develop.DevPlatform {
	return PlatformArchitectury
}

func (f *Architectury) FetchCalls() []develop.DevFetch {
	return []develop.DevFetch{{"Architectury", f.fetchArchApi}}
}

func (f *Architectury) ValidTree(tree fs.FS) bool {
	if !genericCheckPathExists(tree, "settings.gradle") {
		return false
	}
	if !genericCheckPathExists(tree, "common/build.gradle") {
		return false
	}

	// probably architectury, now detect the sub-platforms
	for _, i := range Platforms {
		if i == PlatformArchitectury {
			continue
		}
		sub, err := fs.Sub(tree, i.Name)
		if err != nil {
			continue
		}
		if subPlat, ok := f.SubPlatforms[i]; ok {
			if !subPlat.ValidTree(sub) {
				delete(f.SubPlatforms, i)
			}
		}
	}

	return true
}

func (f *Architectury) ReadVersionFile(tree fs.FS, name string) (map[develop.PropVersion]string, error) {
	if name == "" {
		name = "gradle.properties"
	}
	gradlePropFile, err := tree.Open(name)
	if err != nil {
		return nil, fmt.Errorf("open gradle.properties: %w", err)
	}
	return f.ReadVersions(gradlePropFile)
}

func (f *Architectury) ReadVersions(r io.Reader) (map[develop.PropVersion]string, error) {
	gradlePropContent, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	prop, err := properties.Load(gradlePropContent, properties.UTF8)
	if err != nil {
		return nil, err
	}

	propM := prop.Map()
	a := make(map[develop.PropVersion]string)
	mapProp(a, develop.ModVersion, propM)
	mapProp(a, develop.MinecraftVersion, propM)
	mapProp(a, develop.ArchitecturyVersion, propM)
	if _, ok := f.SubPlatforms[PlatformFabric]; ok {
		mapProp(a, develop.FabricLoaderVersion, propM)
		mapProp(a, develop.FabricApiVersion, propM)
	}
	if _, ok := f.SubPlatforms[PlatformForge]; ok {
		mapProp(a, develop.ForgeVersion, propM)
	}
	if _, ok := f.SubPlatforms[PlatformQuilt]; ok {
		mapProp(a, develop.QuiltLoaderVersion, propM)
		mapProp(a, develop.QuiltFabricApiVersion, propM)
	}
	if _, ok := f.SubPlatforms[PlatformNeoForge]; ok {
		mapProp(a, develop.NeoForgeVersion, propM)
	}
	return a, nil
}

func (f *Architectury) LatestVersion(prop develop.PropVersion, mcVersion string) (string, bool) {
	latestArchApi := f.Meta.Api.FilterGameVersions(mcVersion).GetLatest()
	if prop == develop.ArchitecturyVersion {
		return latestArchApi, true
	}
	for _, p := range f.SubPlatforms {
		if a, ok := p.LatestVersion(prop, mcVersion); ok {
			return a, true
		}
	}
	return "", false
}

func (f *Architectury) LatestLoaderVersion(_ string) (string, error) {
	return "", fmt.Errorf("no loader defined")
}

func (f *Architectury) SubPlatformNames() []string {
	a := make([]string, len(f.SubPlatforms))
	z := 0
	for k := range f.SubPlatforms {
		a[z] = k.Name
		z++
	}
	sort.Strings(a)
	return a
}

func (f *Architectury) fetchArchApi() (err error) {
	f.Meta.Api, err = genericPlatformFetch[shared.ModrinthVersionList](f.Conf.Api, utils.PathJoin(f.Cache, "api.json"), func(r io.Reader, m *shared.ModrinthVersionList) error {
		return json.NewDecoder(r).Decode(m)
	}, func(w io.Writer, m shared.ModrinthVersionList) error {
		return json.NewEncoder(w).Encode(m)
	})
	return err
}
