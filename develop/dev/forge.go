package dev

import (
	"encoding/xml"
	"fmt"
	"github.com/MrMelon54/mcmodupdater/config"
	"github.com/MrMelon54/mcmodupdater/develop"
	"github.com/MrMelon54/mcmodupdater/meta"
	"github.com/MrMelon54/mcmodupdater/meta/shared"
	"github.com/magiconair/properties"
	"io"
	"io/fs"
	"path"
)

var (
	PlatformForge        = develop.DevPlatform{Name: "Forge", Branch: "forge-", Sub: "forge"}
	forgeLoaderMetaPaths = []string{
		"src/main/resources/META-INF/mods.toml",
		"resources/META-INF/mods.toml",
	}
)

type Forge struct {
	Conf  config.ForgeDevelopConfig
	Meta  *ForgeMeta
	Cache string
}

func ForForge(conf config.DevelopConfig, cache string) develop.Develop {
	return &Forge{
		Conf:  conf.Forge,
		Meta:  &ForgeMeta{},
		Cache: path.Join(cache, "forge"),
	}
}

type ForgeMeta struct {
	done chan struct{}
	Api  meta.ForgeApiMeta
}

func (f *Forge) Platform() develop.DevPlatform {
	return PlatformForge
}

func (f *Forge) FetchCalls() []develop.DevFetch {
	return []develop.DevFetch{
		{"API", f.fetchApi},
	}
}

func (f *Forge) ValidTree(tree fs.FS) bool {
	_, ok := genericCheckOnePathExists(tree, forgeLoaderMetaPaths...)
	return ok
}

func (f *Forge) ReadVersionFile(tree fs.FS) (map[develop.PropVersion]string, error) {
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
	mapProp(a, develop.ForgeVersion, propM)
	mapProp(a, develop.ForgeMappingsVersion, propM)
	return a, nil
}

func (f *Forge) LatestVersion(prop develop.PropVersion, mcVersion string) (string, bool) {
	switch prop {
	case develop.ForgeVersion:
		if a, ok := shared.LatestForgeMavenVersion(shared.MavenMeta(f.Meta.Api), mcVersion); ok {
			return a, true
		}
	}
	return "", false
}

func (f *Forge) fetchApi() (err error) {
	f.Meta.Api, err = genericPlatformFetch[meta.ForgeApiMeta](f.Conf.Api, path.Join(f.Cache, "api.json"), func(r io.Reader, m *meta.ForgeApiMeta) error {
		return xml.NewDecoder(r).Decode(m)
	}, func(w io.Writer, m meta.ForgeApiMeta) error {
		return xml.NewEncoder(w).Encode(m)
	})
	return err
}
