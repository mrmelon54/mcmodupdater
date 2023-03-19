package dev

import (
	"encoding/xml"
	"fmt"
	"github.com/MrMelon54/mcmodupdater/config"
	"github.com/MrMelon54/mcmodupdater/develop"
	"github.com/MrMelon54/mcmodupdater/meta"
	"github.com/magiconair/properties"
	"io"
	"io/fs"
	"path"
	"sort"
)

var PlatformArchitectury = develop.DevPlatform{Name: "Architectury", Branch: "arch-"}

type Architectury struct {
	Conf         config.ArchitecturyDevelopConfig
	Meta         *ArchitecturyMeta
	Cache        string
	SubPlatforms map[develop.DevPlatform]develop.Develop
}

type ArchitecturyMeta struct {
	done chan struct{}
	Api  meta.ArchitecturyApiMeta
}

func ForArchitectury(conf config.DevelopConfig, cache string) develop.Develop {
	return &Architectury{
		Conf:  conf.Architectury,
		Meta:  &ArchitecturyMeta{},
		Cache: path.Join(cache, "architectury"),
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

func (f *Architectury) ReadVersionFile(tree fs.FS) (map[develop.PropVersion]string, error) {
	gradlePropFile, err := tree.Open("gradle.properties")
	if err != nil {
		return nil, fmt.Errorf("open gradle.properties: %w", err)
	}
	gradlePropContent, err := io.ReadAll(gradlePropFile)
	if err != nil {
		return nil, fmt.Errorf("read gradle.properties: %w", err)
	}
	prop, err := properties.Load(gradlePropContent, 0)
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
	return a, nil
}

func (f *Architectury) LatestVersion(prop develop.PropVersion, mcVersion string) (string, bool) {
	if prop == develop.ArchitecturyVersion {
		return f.Meta.Api.Versioning.Release, true
	}
	for _, p := range f.SubPlatforms {
		if a, ok := p.LatestVersion(prop, mcVersion); ok {
			return a, true
		}
	}
	return "", false
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
	f.Meta.Api, err = genericPlatformFetch[meta.ArchitecturyApiMeta](f.Conf.Api, path.Join(f.Cache, "api.json"), func(r io.Reader, m *meta.ArchitecturyApiMeta) error {
		return xml.NewDecoder(r).Decode(m)
	}, func(w io.Writer, m meta.ArchitecturyApiMeta) error {
		return xml.NewEncoder(w).Encode(m)
	})
	return err
}
