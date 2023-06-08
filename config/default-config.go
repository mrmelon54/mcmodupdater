package config

import (
	"github.com/MrMelon54/mcmodupdater/paths"
	"gopkg.in/yaml.v3"
	"os"
	"path"
)

func DefaultConfig() Config {
	return Config{
		Develop: DevelopConfig{
			Architectury: ArchitecturyDevelopConfig{
				Api: "https://maven.architectury.dev/dev/architectury/architectury/maven-metadata.xml",
			},
			Fabric: FabricDevelopConfig{
				Game:   "https://meta.fabricmc.net/v2/versions/game",
				Yarn:   "https://meta.fabricmc.net/v2/versions/yarn",
				Loader: "https://meta.fabricmc.net/v2/versions/loader",
				Api:    "https://maven.fabricmc.net/net/fabricmc/fabric-api/fabric-api/maven-metadata.xml",
			},
			Forge: ForgeDevelopConfig{
				Api: "https://maven.minecraftforge.net/net/minecraftforge/forge/maven-metadata.xml",
			},
			Quilt: QuiltDevelopConfig{
				Game:                 "https://meta.quiltmc.org/v3/versions/game",
				QuiltMappings:        "https://meta.quiltmc.org/v3/versions/quilt-mappings",
				QuiltMappingsOnLoom:  "https://maven.quiltmc.org/repository/release/org/quiltmc/quilt-mappings-on-loom/maven-metadata.xml",
				Loader:               "https://meta.quiltmc.org/v3/versions/loader",
				QuiltStandardLibrary: "https://maven.quiltmc.org/repository/release/org/quiltmc/qsl/maven-metadata.xml",
				QuiltedFabricApi:     "https://maven.quiltmc.org/repository/release/org/quiltmc/quilted-fabric-api/quilted-fabric-api/maven-metadata.xml",
			},
		},
		Cache: true,
	}
}

func Load() (*Config, error) {
	conf := DefaultConfig()
	vFile := path.Join(paths.UserConfigDir(), "config.yml")
	f, err := os.Open(vFile)
	if err != nil {
		return &conf, nil
	}

	err = yaml.NewDecoder(f).Decode(&conf)
	return &conf, err
}

func (c *Config) Save() error {
	vFile := path.Join(paths.UserConfigDir(), "config.yml")
	err := os.MkdirAll(path.Dir(vFile), os.ModePerm)
	if err != nil {
		return err
	}

	f, err := os.Create(vFile)
	if err != nil {
		return err
	}

	return yaml.NewEncoder(f).Encode(c)
}
