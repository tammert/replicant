package replicant

import (
	"github.com/blang/semver/v4"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	log "github.com/sirupsen/logrus"
	"strings"
)

func Run(configFile string) {
	config := ReadConfig(configFile)
	mirrorHighestTag(config)
}

// mirrorAllTags mirrors all tags.
func mirrorAllTags(config Config) {
	for _, image := range config.Images {
		tags := listTags(image.UpstreamRepository)
		for _, tag := range tags {
			cloneToRepo(image.UpstreamRepository+":"+tag, image.DownstreamRepository+":"+tag)
		}
	}
}

// mirrorHighestTag mirrors the highest SemVer tag.
func mirrorHighestTag(config Config) {
	for _, image := range config.Images {
		tag := findHighestTag(image.UpstreamRepository)
		from := image.UpstreamRepository + ":" + image.TagPrefix + tag.String()
		to := image.DownstreamRepository + ":" + image.TagPrefix + tag.String()
		cloneToRepo(from, to)
	}
}

func findHighestTag(repository string) semver.Version {
	tags := listTags(repository)
	versions := semVerSort(tags)
	return versions[len(versions)-1]
}

func cloneToRepo(from string, to string) {
	upstream, err := name.ParseReference(from)
	if err != nil {
		log.Fatal(err)
	}

	// Grab the upstream image.
	image, err := remote.Image(upstream)
	if err != nil {
		log.Fatal(err)
	}

	private, err := name.ParseReference(to)
	if err != nil {
		log.Fatal(err)
	}

	//TODO: check if downstream repository already contains the image
	//TODO: if image is already downstream, compare SHA
	//TODO: add flag/config to specify whether to override if SHA is different -> default no?

	// Check if downstream repository is in GCR.
	if strings.Contains(private.Context().RegistryStr(), "gcr.io") {
		auth := remote.WithAuth(gcrAuthenticator())
		err = remote.Write(private, image, auth)
	} else {
		// Write the image to the private repository.
		err = remote.Write(private, image)
	}
	if err != nil {
		log.Fatal(err)
	}
}

func listTags(repository string) []string {
	r, err := name.NewRepository(repository)
	if err != nil {
		log.Fatal(err)
	}

	list, err := remote.List(r)
	if err != nil {
		log.Fatal(err)
	}

	return list
}

// semVerSort sorts SemVer versions, removes non-SemVer values.
func semVerSort(xs []string) semver.Versions {
	var xv semver.Versions

	for _, v := range xs {
		version, err := semver.ParseTolerant(v)
		if err != nil {
			log.Debugf("%s is not a SemVer version, ignoring", v)
			continue
		}
		// Tags with only numbers will incorrectly be parsed by ParseTolerant, do a dirty 'verification' on Major number.
		if version.Major > 1024 {
			log.Debugf("%s is probably not a SemVer version, ignoring", version.String())
			continue
		}
		// Handle prerelease versions. TODO: allow via flag/config
		if version.Pre != nil {
			log.Debugf("%s is a prerelease version, ignoring", version.String())
			continue
		}

		xv = append(xv, version)
	}

	semver.Sort(xv)
	return xv
}
