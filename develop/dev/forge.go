package dev

import (
	"encoding/xml"
	"fmt"
	"github.com/magiconair/properties"
	"github.com/mrmelon54/mcmodupdater/config"
	"github.com/mrmelon54/mcmodupdater/develop"
	"github.com/mrmelon54/mcmodupdater/meta"
	"github.com/mrmelon54/mcmodupdater/meta/shared"
	"github.com/mrmelon54/mcmodupdater/utils"
	"io"
	"io/fs"
)

var (
	PlatformForge        = develop.DevPlatform{Name: "Forge", Sub: "forge"}
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
		Cache: utils.PathJoin(cache, "forge"),
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
		{"API", f.FetchApi},
	}
}

func (f *Forge) ValidTree(tree fs.FS) bool {
	_, ok := genericCheckOnePathExists(tree, forgeLoaderMetaPaths...)
	return ok
}

func (f *Forge) ReadVersionFile(tree fs.FS, name string) (map[develop.PropVersion]string, error) {
	if name == "" {
		name = "gradle.properties"
	}
	gradlePropFile, err := tree.Open(name)
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
		a, err := f.LatestLoaderVersion(mcVersion)
		return a, err == nil
	default:
	}
	return "", false
}

func (f *Forge) LatestLoaderVersion(mcVersion string) (string, error) {
	err := f.FetchApi()
	if err != nil {
		return "", err
	}
	version, ok := shared.LatestForgeMavenVersion(shared.MavenMeta(f.Meta.Api), mcVersion)
	if !ok {
		return "", fmt.Errorf("no forge loaders found")
	}
	return version, nil
}

func (f *Forge) FetchApi() (err error) {
	f.Meta.Api, err = genericPlatformFetch[meta.ForgeApiMeta](f.Conf.Api, utils.PathJoin(f.Cache, "api.xml"), func(r io.Reader, m *meta.ForgeApiMeta) error {
		return xml.NewDecoder(r).Decode(m)
	}, func(w io.Writer, m meta.ForgeApiMeta) error {
		return xml.NewEncoder(w).Encode(m)
	})
	return err
}
