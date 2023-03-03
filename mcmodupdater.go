package mcmodupdater

import (
	"bufio"
	"github.com/MrMelon54/mcmodupdater/config"
	"github.com/MrMelon54/mcmodupdater/develop"
	"github.com/MrMelon54/mcmodupdater/develop/dev"
	"github.com/MrMelon54/mcmodupdater/paths"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/magiconair/properties"
	"io"
	"io/fs"
	"os"
	"path"
	"strings"
)

type McModUpdater struct {
	platforms map[string]develop.Develop
	platMap   []string
	platArch  *dev.Architectury
	cwd       string
	cache     string
	repo      *git.Repository
	branches  []*develop.BranchInfo
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

func NewMcModUpdater(conf *config.Config, repo *git.Repository, cwd string) (*McModUpdater, error) {
	cache := paths.UserCacheDir()
	err := os.MkdirAll(cache, fs.ModePerm)
	if err != nil {
		return nil, err
	}
	platCache := path.Join(cache, "platforms")
	err = os.MkdirAll(cache, fs.ModePerm)
	if err != nil {
		return nil, err
	}

	platMap := make([]string, len(dev.DevelopPlatformsFactory))
	plat := make(map[string]develop.Develop)
	for i, j := range dev.DevelopPlatformsFactory {
		d := j(conf.Develop, platCache)
		p := d.Platform()
		plat[p.Name] = d
		platMap[i] = p.Name
	}

	return &McModUpdater{
		cwd:       cwd,
		platforms: plat,
		platMap:   platMap,
		platArch:  dev.ForArchitectury(conf.Develop, platCache),
		repo:      repo,
		cache:     cache,
	}, nil
}

func (m *McModUpdater) PlatArch() *dev.Architectury           { return m.platArch }
func (m *McModUpdater) PlatMap() []string                     { return m.platMap }
func (m *McModUpdater) Platforms() map[string]develop.Develop { return m.platforms }

func (m *McModUpdater) getPlatformFromTree(tree *object.Tree) (develop.Develop, bool) {
	for _, i := range m.platforms {
		if i.ValidTree(tree) {
			return i, true
		}
	}
	return nil, false
}

func (m *McModUpdater) getPlatformFromArchTree(tree *object.Tree) (develop.Develop, bool) {
	archPlats := make(map[develop.DevPlatform]develop.Develop)
	for _, i := range m.platforms {
		if i.ValidTreeArch(tree) {
			archPlats[i.Platform()] = i
			continue
		}
	}
	m.platArch.SubPlatforms = archPlats
	return m.platArch, true
}

func (m *McModUpdater) getPlatformFromName(name string) (develop.Develop, bool) {
	n := strings.ToLower(name)
	for _, i := range m.platforms {
		if strings.ToLower(i.Platform().Name) == n {
			return i, true
		}
	}
	return nil, false
}

func (m *McModUpdater) LoadBranch(ref *plumbing.Reference) (*develop.BranchInfo, error) {
	commit, err := m.repo.CommitObject(ref.Hash())
	if err != nil {
		return nil, err
	}
	tree, err := commit.Tree()
	if err != nil {
		return nil, err
	}

	useArch := m.platArch.ValidTree(tree)

	var platform develop.Develop
	var ok bool

	if useArch {
		platform, ok = m.getPlatformFromArchTree(tree)
		if !ok {
			return nil, nil
		}
	} else {
		platform, ok = m.getPlatformFromTree(tree)
		if !ok {
			return nil, nil
		}
	}

	// Parse library version files
	versions, err := platform.ReadVersionFile(tree)
	if err != nil {
		return nil, err
	}

	return &develop.BranchInfo{
		Plumb:    ref,
		Platform: platform,
		Versions: versions,
	}, nil
}

func (m *McModUpdater) IsClean() (bool, error) {
	w, err := m.repo.Worktree()
	if err != nil {
		return false, err
	}
	s, err := w.Status()
	if err != nil {
		return false, err
	}
	return s.IsClean(), nil
}

func (m *McModUpdater) CurrentBranch() (*plumbing.Reference, error) {
	return m.repo.Head()
}

func (m *McModUpdater) GetProjectPath() string {
	return m.cwd
}

func (m *McModUpdater) VersionUpdateList(info *develop.BranchInfo) VersionUpdateList {
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

func (m *McModUpdater) useIfExists(v VersionUpdateList, branch *develop.BranchInfo, k develop.PropVersion) VersionUpdateList {
	if a, ok := branch.Versions[k]; ok {
		v = append(v, VersionUpdateItem{k, a, ""})
	}
	return v
}

func (m *McModUpdater) useIfExistsUpdate(v VersionUpdateList, branch *develop.BranchInfo, k develop.PropVersion) VersionUpdateList {
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

func (m *McModUpdater) UpdateToVersion(out io.StringWriter, ver map[develop.PropVersion]string) error {
	gProp, err := os.Open("gradle.properties")
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
