package dev

import (
	"errors"
	"github.com/MrMelon54/mcmodupdater/config"
	"github.com/MrMelon54/mcmodupdater/develop"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path"
	"time"
)

var ErrOutdatedCache = errors.New("outdated cache")

var (
	DevelopPlatformsFactory = []func(config.DevelopConfig, string) develop.Develop{
		// Architectury MUST be handled separately
		ForFabric,
		ForForge,
		ForQuilt,
		ForNeoForge,
		//TODO: add LiteLoader options.. https://dl.liteloader.com/versions/versions.json
	}
	Platforms = []develop.DevPlatform{
		PlatformFabric,
		PlatformForge,
		PlatformQuilt,
		PlatformNeoForge,
	}
)

func genericPlatformFetch[T any](url, cache string, cbR func(io.Reader, *T) error, cbW func(io.Writer, T) error) (t T, err error) {
	err = genericPlatformCacheLoad[T](cache, &t, cbR)
	if err == nil {
		return
	}
	resp, err := http.Get(url)
	if err != nil {
		return
	}
	err = cbR(resp.Body, &t)
	if err != nil {
		return
	}
	err = genericPlatformCacheSave[T](cache, cbW, t)
	return
}

func genericPlatformCacheLoad[T any](p string, t *T, cbR func(io.Reader, *T) error) error {
	open, err := os.Open(p)
	if err != nil {
		return err
	}
	stat, err := open.Stat()
	if time.Now().Sub(stat.ModTime()).Abs() > time.Hour {
		err = ErrOutdatedCache
	}
	if err != nil {
		return err
	}
	return cbR(open, t)
}

func genericPlatformCacheSave[T any](p string, cbW func(io.Writer, T) error, t T) error {
	// if path is empty then don't write
	if p == "" {
		return nil
	}

	err := os.MkdirAll(path.Dir(p), fs.ModePerm)
	if err != nil {
		return err
	}
	create, err := os.Create(p)
	if err != nil {
		return err
	}
	return cbW(create, t)
}

func mapProp(out map[develop.PropVersion]string, target develop.PropVersion, in map[string]string) {
	if v, ok := in[target.Key()]; ok {
		out[target] = v
		return
	}
}

func genericCheckOnePathExists(tree fs.FS, name ...string) (string, bool) {
	for _, i := range name {
		if genericCheckPathExists(tree, i) {
			return i, true
		}
	}
	return "", false
}

func genericCheckPathExists(tree fs.FS, name string) bool {
	_, err := fs.Stat(tree, name)
	return err == nil
}
