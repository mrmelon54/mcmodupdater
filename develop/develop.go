package develop

import (
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

type Develop interface {
	Platform() DevPlatform
	FetchCalls() []DevFetch
	ValidTree(tree *object.Tree) bool
	ValidTreeArch(tree *object.Tree) bool
	ReadVersionFile(tree *object.Tree) (map[PropVersion]string, error)
	LatestVersion(prop PropVersion, mcVersion string) (string, bool)
}

type DevPlatform struct {
	Name   string
	Branch string
}

type DevFetch struct {
	Name string
	Call func() error
}

type BranchInfo struct {
	Plumb    *plumbing.Reference
	Platform Develop
	Versions map[PropVersion]string
}
