package dev

import (
	"encoding/json"
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
	PlatformFabric        = develop.DevPlatform{Name: "Fabric", Sub: "fabric"}
	fabricLoaderMetaPaths = []string{
		"src/main/resources/fabric.mod.json",
		"resources/fabric.mod.json",
	}
)

type Fabric struct {
	Conf  config.FabricDevelopConfig
	Meta  *FabricMeta
	Cache string
}

type FabricMeta struct {
	done   chan struct{}
	Game   meta.FabricGameMeta
	Yarn   meta.FabricYarnMeta
	Loader meta.FabricLoaderMeta
	Api    meta.FabricApiMeta
}

func ForFabric(conf config.DevelopConfig, cache string) develop.Develop {
	return &Fabric{
		Conf:  conf.Fabric,
		Meta:  &FabricMeta{},
		Cache: utils.PathJoin(cache, "fabric"),
	}
}

func (f *Fabric) Platform() develop.DevPlatform {
	return PlatformFabric
}

func (f *Fabric) FetchCalls() []develop.DevFetch {
	return []develop.DevFetch{
		{"Game", f.FetchGame},
		{"Yarn", f.FetchYarn},
		{"Loader", f.FetchLoader},
		{"API", f.FetchApi},
	}
}

func (f *Fabric) ValidTree(tree fs.FS) bool {
	_, ok := genericCheckOnePathExists(tree, fabricLoaderMetaPaths...)
	return ok
}

func (f *Fabric) ReadVersionFile(tree fs.FS, name string) (map[develop.PropVersion]string, error) {
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
	mapProp(a, develop.YarnMappingsVersion, propM)
	mapProp(a, develop.FabricLoaderVersion, propM)
	mapProp(a, develop.FabricApiVersion, propM)
	return a, nil
}

func (f *Fabric) LatestVersion(prop develop.PropVersion, mcVersion string) (string, bool) {
	switch prop {
	case develop.FabricLoaderVersion:
		if len(f.Meta.Loader) > 0 {
			return f.Meta.Loader[0].Version, true
		}
	case develop.FabricApiVersion:
		if a, ok := shared.LatestMavenVersion(shared.MavenMeta(f.Meta.Api), mcVersion); ok {
			return a, true
		}
	case develop.YarnMappingsVersion:
		if a, ok := shared.LatestYarnVersion(f.Meta.Yarn, mcVersion); ok {
			return a.Version, ok
		}
	default:
	}
	return "", false
}

func (f *Fabric) FetchGame() (err error) {
	f.Meta.Game, err = genericPlatformFetch[meta.FabricGameMeta](f.Conf.Game, utils.PathJoin(f.Cache, "game.json"), func(r io.Reader, m *meta.FabricGameMeta) error {
		return json.NewDecoder(r).Decode(m)
	}, func(w io.Writer, m meta.FabricGameMeta) error {
		return json.NewEncoder(w).Encode(m)
	})
	return err
}

func (f *Fabric) FetchYarn() (err error) {
	f.Meta.Yarn, err = genericPlatformFetch[meta.FabricYarnMeta](f.Conf.Yarn, utils.PathJoin(f.Cache, "yarn.json"), func(r io.Reader, m *meta.FabricYarnMeta) error {
		return json.NewDecoder(r).Decode(m)
	}, func(w io.Writer, m meta.FabricYarnMeta) error {
		return json.NewEncoder(w).Encode(m)
	})
	return err
}

func (f *Fabric) FetchLoader() (err error) {
	f.Meta.Loader, err = genericPlatformFetch[meta.FabricLoaderMeta](f.Conf.Loader, utils.PathJoin(f.Cache, "loader.json"), func(r io.Reader, m *meta.FabricLoaderMeta) error {
		return json.NewDecoder(r).Decode(m)
	}, func(w io.Writer, m meta.FabricLoaderMeta) error {
		return json.NewEncoder(w).Encode(m)
	})
	return err
}

func (f *Fabric) FetchApi() (err error) {
	f.Meta.Api, err = genericPlatformFetch[meta.FabricApiMeta](f.Conf.Api, utils.PathJoin(f.Cache, "api.xml"), func(r io.Reader, m *meta.FabricApiMeta) error {
		return xml.NewDecoder(r).Decode(m)
	}, func(w io.Writer, m meta.FabricApiMeta) error {
		return xml.NewEncoder(w).Encode(m)
	})
	return err
}
