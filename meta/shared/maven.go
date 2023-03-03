package shared

import (
	"encoding/xml"
	"strings"
)

type MavenMeta struct {
	XMLName    xml.Name          `xml:"metadata"`
	Text       string            `xml:",chardata"`
	GroupId    string            `xml:"groupId"`
	ArtifactId string            `xml:"artifactId"`
	Versioning ApiMetaVersioning `xml:"versioning"`
}

type ApiMetaVersioning struct {
	Text        string                      `xml:",chardata"`
	Latest      string                      `xml:"latest"`
	Release     string                      `xml:"release"`
	Versions    MavenMetaVersioningVersions `xml:"versions"`
	LastUpdated string                      `xml:"lastUpdated"`
}

type MavenMetaVersioningVersions struct {
	Text    string   `xml:",chardata"`
	Version []string `xml:"version"`
}

func LatestMavenVersion(m MavenMeta, mc string) (string, bool) {
	var a string
	for _, i := range m.Versioning.Versions.Version {
		if strings.HasSuffix(i, "+"+mc) {
			a = i
		}
		if strings.HasSuffix(i, "-"+mc) {
			a = i
		}
	}
	return a, a != ""
}

func LatestForgeMavenVersion(m MavenMeta, mc string) (string, bool) {
	for _, i := range m.Versioning.Versions.Version {
		if strings.HasPrefix(i, mc+"-") {
			return i, true
		}
	}
	return "", false
}
