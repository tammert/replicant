package replicant

import (
	"fmt"
	"github.com/blang/semver/v4"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/remote/transport"
	log "github.com/sirupsen/logrus"
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
			log.Debug(tag)
		}
	}
}

// mirrorHighestTag mirrors the highest SemVer tag.
func mirrorHighestTag(config Config) {
	for _, image := range config.Images {
		tag := findHighestTag(image.UpstreamRepository)

		from := image.UpstreamRepository + ":" + image.TagPrefix + tag.String()
		fromReference, err := name.ParseReference(from)
		if err != nil {
			log.Error(err)
		}

		to := image.DownstreamRepository + ":" + image.TagPrefix + tag.String()
		toReference, err := name.ParseReference(to)
		if err != nil {
			log.Error(err)
		}

		checkDownstreamForTag(toReference)
		cloneToRepo(fromReference, toReference)
	}
}

func checkDownstreamForTag(to name.Reference) {
	auth := remote.WithAuth(getCorrectAuth(to.Context().RegistryStr()))
	descriptor, err := remote.Get(to, auth)
	if err != nil {
		t, ok := err.(*transport.Error)
		if ok {
			if t.StatusCode == 404 {
				log.Debug("not found!")
			}
		}
		log.Error(err)
	}
	fmt.Println(descriptor)
}

func findHighestTag(repository string) semver.Version {
	tags := listTags(repository)
	versions := semVerSort(tags)
	return versions[len(versions)-1]
}

func cloneToRepo(from name.Reference, to name.Reference) {
	// Grab the upstream image.
	image, err := remote.Image(from)
	if err != nil {
		log.Fatal(err)
	}

	//TODO: check if downstream repository already contains the image
	//TODO: if image is already downstream, compare SHA
	//TODO: add flag/config to specify whether to override if SHA is different -> default no?

	auth := remote.WithAuth(getCorrectAuth(to.Context().RegistryStr()))
	err = remote.Write(to, image, auth)
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
