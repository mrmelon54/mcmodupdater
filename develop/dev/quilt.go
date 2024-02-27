package dev

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/komkom/toml"
	"github.com/magiconair/properties"
	"github.com/mrmelon54/mcmodupdater/config"
	"github.com/mrmelon54/mcmodupdater/develop"
	"github.com/mrmelon54/mcmodupdater/meta"
	libVersion "github.com/mrmelon54/mcmodupdater/meta/quilt/lib-version"
	"github.com/mrmelon54/mcmodupdater/meta/shared"
	"github.com/mrmelon54/mcmodupdater/utils"
	"io"
	"io/fs"
)

var (
	PlatformQuilt        = develop.DevPlatform{Name: "Quilt", Sub: "quilt"}
	quiltLoaderMetaPaths = []string{
		"src/main/resources/quilt.mod.json",
		"resources/quilt.mod.json",
	}
)

type Quilt struct {
	Conf  config.QuiltDevelopConfig
	Meta  *QuiltMeta
	Cache string
}

type QuiltMeta struct {
	done                 chan struct{}
	Game                 meta.QuiltGameMeta
	QuiltMappings        meta.QuiltMappingsMeta
	QuiltMappingsOnLoom  meta.QuiltMappingsOnLoomMeta
	Loader               meta.QuiltLoaderMeta
	QuiltStandardLibrary meta.QuiltStandardLibraryMeta
	QuiltedFabricApi     meta.QuiltedFabricApiMeta
}

func ForQuilt(conf config.DevelopConfig, cache string) develop.Develop {
	return &Quilt{
		Conf:  conf.Quilt,
		Meta:  &QuiltMeta{},
		Cache: utils.PathJoin(cache, "quilt"),
	}
}

func (q *Quilt) Platform() develop.DevPlatform {
	return PlatformQuilt
}

func (q *Quilt) FetchCalls() []develop.DevFetch {
	return []develop.DevFetch{
		{"Game", q.FetchGame},
		{"Mappings", q.FetchQuiltMappings},
		{"Mappings on Loom", q.FetchQuiltMappingsOnLoom},
		{"Loader", q.FetchLoader},
		{"Standard Library", q.FetchQuiltStandardLibrary},
		{"Fabric Api", q.FetchQuiltedFabricApi},
	}
}

func (q *Quilt) ValidTree(tree fs.FS) bool {
	_, ok := genericCheckOnePathExists(tree, quiltLoaderMetaPaths...)
	return ok
}

func (q *Quilt) ReadVersionFile(tree fs.FS, name string) (map[develop.PropVersion]string, error) {
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

	gradleLibVersions, err := tree.Open("gradle/libs.versions.toml")
	if err != nil {
		return nil, fmt.Errorf("contents gradle/libs.versions.toml: %w", err)
	}
	var v libVersion.LibVersion
	err = json.NewDecoder(toml.New(gradleLibVersions)).Decode(&v)
	if err != nil {
		return nil, err
	}

	propM := v.Versions
	propMV := prop.Map()
	a := make(map[develop.PropVersion]string)
	mapProp(a, develop.ModVersion, propMV)
	mapProp(a, develop.MinecraftVersion, propM)
	mapProp(a, develop.QuiltMappingsVersion, propM)
	mapProp(a, develop.QuiltLoaderVersion, propM)
	mapProp(a, develop.QuiltFabricApiVersion, propM)
	return a, nil
}

func (q *Quilt) LatestVersion(prop develop.PropVersion, mcVersion string) (string, bool) {
	switch prop {
	case develop.QuiltLoaderVersion:
		if len(q.Meta.Loader) > 0 {
			return q.Meta.Loader[0].Version, true
		}
	case develop.QuiltFabricApiVersion:
		if a, ok := shared.LatestMavenVersion(shared.MavenMeta(q.Meta.QuiltedFabricApi), mcVersion); ok {
			return a, ok
		}
	case develop.QuiltMappingsVersion:
		if a, ok := shared.LatestYarnVersion(q.Meta.QuiltMappings, mcVersion); ok {
			return a.Version, ok
		}
	default:
	}
	return "", false
}

func (q *Quilt) FetchGame() (err error) {
	q.Meta.Game, err = genericPlatformFetch[meta.QuiltGameMeta](q.Conf.Game, utils.PathJoin(q.Cache, "game.json"), func(r io.Reader, m *meta.QuiltGameMeta) error {
		return json.NewDecoder(r).Decode(m)
	}, func(w io.Writer, m meta.QuiltGameMeta) error {
		return json.NewEncoder(w).Encode(m)
	})
	return err
}

func (q *Quilt) FetchQuiltMappings() (err error) {
	q.Meta.QuiltMappings, err = genericPlatformFetch[meta.QuiltMappingsMeta](q.Conf.QuiltMappings, utils.PathJoin(q.Cache, "quilt-mappings.json"), func(r io.Reader, m *meta.QuiltMappingsMeta) error {
		return json.NewDecoder(r).Decode(m)
	}, func(w io.Writer, m meta.QuiltMappingsMeta) error {
		return json.NewEncoder(w).Encode(m)
	})
	return err
}

func (q *Quilt) FetchQuiltMappingsOnLoom() (err error) {
	q.Meta.QuiltMappingsOnLoom, err = genericPlatformFetch[meta.QuiltMappingsOnLoomMeta](q.Conf.QuiltMappingsOnLoom, utils.PathJoin(q.Cache, "quilt-mappings-loom.xml"), func(r io.Reader, m *meta.QuiltMappingsOnLoomMeta) error {
		return xml.NewDecoder(r).Decode(m)
	}, func(w io.Writer, m meta.QuiltMappingsOnLoomMeta) error {
		return xml.NewEncoder(w).Encode(m)
	})
	return err
}

func (q *Quilt) FetchLoader() (err error) {
	q.Meta.Loader, err = genericPlatformFetch[meta.QuiltLoaderMeta](q.Conf.Loader, utils.PathJoin(q.Cache, "loader.json"), func(r io.Reader, m *meta.QuiltLoaderMeta) error {
		return json.NewDecoder(r).Decode(m)
	}, func(w io.Writer, m meta.QuiltLoaderMeta) error {
		return json.NewEncoder(w).Encode(m)
	})
	return err
}

func (q *Quilt) FetchQuiltStandardLibrary() (err error) {
	q.Meta.QuiltStandardLibrary, err = genericPlatformFetch[meta.QuiltStandardLibraryMeta](q.Conf.QuiltStandardLibrary, utils.PathJoin(q.Cache, "qsl.xml"), func(r io.Reader, m *meta.QuiltStandardLibraryMeta) error {
		return xml.NewDecoder(r).Decode(m)
	}, func(w io.Writer, m meta.QuiltStandardLibraryMeta) error {
		return xml.NewEncoder(w).Encode(m)
	})
	return err
}

func (q *Quilt) FetchQuiltedFabricApi() (err error) {
	q.Meta.QuiltedFabricApi, err = genericPlatformFetch[meta.QuiltedFabricApiMeta](q.Conf.QuiltedFabricApi, utils.PathJoin(q.Cache, "qfa.xml"), func(r io.Reader, m *meta.QuiltedFabricApiMeta) error {
		return xml.NewDecoder(r).Decode(m)
	}, func(w io.Writer, m meta.QuiltedFabricApiMeta) error {
		return xml.NewEncoder(w).Encode(m)
	})
	return err
}
