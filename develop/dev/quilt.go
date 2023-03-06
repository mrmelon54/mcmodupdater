package dev

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/MrMelon54/mcmodupdater/config"
	"github.com/MrMelon54/mcmodupdater/develop"
	"github.com/MrMelon54/mcmodupdater/meta"
	libVersion "github.com/MrMelon54/mcmodupdater/meta/quilt/lib-version"
	"github.com/MrMelon54/mcmodupdater/meta/shared"
	"github.com/komkom/toml"
	"github.com/magiconair/properties"
	"io"
	"io/fs"
	"path"
)

var (
	PlatformQuilt        = develop.DevPlatform{Name: "Quilt", Branch: "quilt-", Sub: "quilt"}
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
		Cache: path.Join(cache, "quilt"),
	}
}

func (q *Quilt) Platform() develop.DevPlatform {
	return PlatformQuilt
}

func (q *Quilt) FetchCalls() []develop.DevFetch {
	return []develop.DevFetch{
		{"Game", q.fetchGame},
		{"Mappings", q.fetchQuiltMappings},
		{"Mappings on Loom", q.fetchQuiltMappingsOnLoom},
		{"Loader", q.fetchLoader},
		{"Standard Library", q.fetchQuiltStandardLibrary},
		{"Fabric Api", q.fetchQuiltedFabricApi},
	}
}

func (q *Quilt) ValidTree(tree fs.FS) bool {
	_, ok := genericCheckOnePathExists(tree, quiltLoaderMetaPaths...)
	return ok
}

func (q *Quilt) ReadVersionFile(tree fs.FS) (map[develop.PropVersion]string, error) {
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
	}
	return "", false
}

func (q *Quilt) fetchGame() (err error) {
	q.Meta.Game, err = genericPlatformFetch[meta.QuiltGameMeta](q.Conf.Game, path.Join(q.Cache, "game.json"), func(r io.Reader, m *meta.QuiltGameMeta) error {
		return json.NewDecoder(r).Decode(m)
	}, func(w io.Writer, m meta.QuiltGameMeta) error {
		return json.NewEncoder(w).Encode(m)
	})
	return err
}

func (q *Quilt) fetchQuiltMappings() (err error) {
	q.Meta.QuiltMappings, err = genericPlatformFetch[meta.QuiltMappingsMeta](q.Conf.QuiltMappings, path.Join(q.Cache, "quilt-mappings.json"), func(r io.Reader, m *meta.QuiltMappingsMeta) error {
		return json.NewDecoder(r).Decode(m)
	}, func(w io.Writer, m meta.QuiltMappingsMeta) error {
		return json.NewEncoder(w).Encode(m)
	})
	return err
}

func (q *Quilt) fetchQuiltMappingsOnLoom() (err error) {
	q.Meta.QuiltMappingsOnLoom, err = genericPlatformFetch[meta.QuiltMappingsOnLoomMeta](q.Conf.QuiltMappingsOnLoom, path.Join(q.Cache, "quilt-mappings-loom.json"), func(r io.Reader, m *meta.QuiltMappingsOnLoomMeta) error {
		return xml.NewDecoder(r).Decode(m)
	}, func(w io.Writer, m meta.QuiltMappingsOnLoomMeta) error {
		return xml.NewEncoder(w).Encode(m)
	})
	return err
}

func (q *Quilt) fetchLoader() (err error) {
	q.Meta.Loader, err = genericPlatformFetch[meta.QuiltLoaderMeta](q.Conf.Loader, path.Join(q.Cache, "loader.json"), func(r io.Reader, m *meta.QuiltLoaderMeta) error {
		return json.NewDecoder(r).Decode(m)
	}, func(w io.Writer, m meta.QuiltLoaderMeta) error {
		return json.NewEncoder(w).Encode(m)
	})
	return err
}

func (q *Quilt) fetchQuiltStandardLibrary() (err error) {
	q.Meta.QuiltStandardLibrary, err = genericPlatformFetch[meta.QuiltStandardLibraryMeta](q.Conf.QuiltStandardLibrary, path.Join(q.Cache, "qsl.json"), func(r io.Reader, m *meta.QuiltStandardLibraryMeta) error {
		return xml.NewDecoder(r).Decode(m)
	}, func(w io.Writer, m meta.QuiltStandardLibraryMeta) error {
		return xml.NewEncoder(w).Encode(m)
	})
	return err
}

func (q *Quilt) fetchQuiltedFabricApi() (err error) {
	q.Meta.QuiltedFabricApi, err = genericPlatformFetch[meta.QuiltedFabricApiMeta](q.Conf.QuiltedFabricApi, path.Join(q.Cache, "qfa.json"), func(r io.Reader, m *meta.QuiltedFabricApiMeta) error {
		return xml.NewDecoder(r).Decode(m)
	}, func(w io.Writer, m meta.QuiltedFabricApiMeta) error {
		return xml.NewEncoder(w).Encode(m)
	})
	return err
}
