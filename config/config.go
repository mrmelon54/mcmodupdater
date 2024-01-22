package config

type Config struct {
	Develop DevelopConfig `yaml:"develop"`
	Cache   bool          `yaml:"cache"`
}

type DevelopConfig struct {
	Architectury ArchitecturyDevelopConfig `yaml:"architectury"`
	Fabric       FabricDevelopConfig       `yaml:"fabric"`
	Forge        ForgeDevelopConfig        `yaml:"forge"`
	Quilt        QuiltDevelopConfig        `yaml:"quilt"`
	NeoForge     NeoForgeDevelopConfig     `yaml:"neoforge"`
}

type ArchitecturyDevelopConfig struct {
	Api string `yaml:"api"`
}

type FabricDevelopConfig struct {
	Game   string `yaml:"game"`
	Yarn   string `yaml:"yarn"`
	Loader string `yaml:"loader"`
	Api    string `yaml:"api"`
}

type ForgeDevelopConfig struct {
	Api string `yaml:"api"`
}

type QuiltDevelopConfig struct {
	Game                 string `yaml:"game"`
	QuiltMappings        string `yaml:"quiltMappings"`
	QuiltMappingsOnLoom  string `yaml:"quiltMappingsOnLoom"`
	Loader               string `yaml:"loader"`
	QuiltStandardLibrary string `yaml:"quiltStandardLibrary"`
	QuiltedFabricApi     string `yaml:"quiltedFabricApi"`
}

type NeoForgeDevelopConfig struct {
	Api string `yaml:"api"`
}
