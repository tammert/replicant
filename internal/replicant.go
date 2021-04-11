package replicant

//TODO: better logging

import (
	"github.com/Masterminds/semver/v3"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/remote/transport"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"sort"
)

func Run(configFile string) {
	config := ReadConfig(configFile)
	for _, image := range config.Images {
		switch image.Mode {
		case "highest":
			mirrorHighestTag(image)
		case "higher":
			mirrorHigherTags(image)
		case "all":
			mirrorAllTags(image)
		case "semver":
			mirrorSemVerTags(image)
		default:
			mirrorHighestTag(image)
		}
	}
}

// mirrorAllTags mirrors all tags.
func mirrorAllTags(image *ImageConfig) {
	tags := listTags(image.UpstreamRepository)
	for _, tag := range tags {
		mirrorTag(image, tag)
	}
}

// mirrorSemVerTags mirrors all SemVer tags.
func mirrorSemVerTags(image *ImageConfig) {
	tags := semVerSort(listTags(image.UpstreamRepository))
	for _, tag := range tags {
		mirrorTag(image, tag.Original())
	}
}

// mirrorHigherTags mirrors all tags that are a higher SemVer version than the highest available in the downstream repository.
func mirrorHigherTags(image *ImageConfig) {
	var tagsToMirror []*semver.Version

	highestDownstreamTag := findHighestTag(image.DownstreamRepository)
	upstreamTags := semVerSort(listTags(image.UpstreamRepository))
	for _, tag := range upstreamTags {
		if tag.GreaterThan(highestDownstreamTag) {
			tagsToMirror = append(tagsToMirror, tag)
		}
	}
	for _, t := range tagsToMirror {
		mirrorTag(image, t.Original())
	}
}

// mirrorHighestTag mirrors the highest SemVer tag, if it's not already in the downstream repository.
func mirrorHighestTag(image *ImageConfig) {
	tag := findHighestTag(image.UpstreamRepository)
	mirrorTag(image, tag.Original())
}

func mirrorTag(image *ImageConfig, tag string) {
	upstreamReference, err := name.ParseReference(image.UpstreamRepository + ":" + tag)
	if err != nil {
		log.Error(err)
	}
	downstreamReference, err := name.ParseReference(image.DownstreamRepository + ":" + tag)
	if err != nil {
		log.Error(err)
	}

	_, err = remote.Head(downstreamReference, getAuth(downstreamReference.Context().RegistryStr()))
	if err != nil {
		if t, ok := err.(*transport.Error); ok {
			if t.StatusCode == 404 {
				log.Debugf("image %s not found, mirroring", downstreamReference.String())
				mirrorImage(upstreamReference, downstreamReference)
				return
			}
		} else {
			log.Error(err)
		}
	}

	// Tag is present downstream. If configured with `replace-tag`, check if the image ID matches.
	if viper.GetBool("replace-tag") {
		upstreamImageID, err := getImage(upstreamReference).ConfigName()
		if err != nil {
			log.Error(err)
		}
		downstreamImageID, err := getImage(downstreamReference).ConfigName()
		if err != nil {
			log.Error(err)
		}

		// Compare the image IDs.
		if upstreamImageID == downstreamImageID {
			log.Debug("image IDs are identical, no need to replace")
		} else {
			log.Debugf("image IDs are different, replacing %s with %s", downstreamImageID, upstreamImageID)
			writeImage(upstreamReference, getImage(upstreamReference))
		}
	}
}

//TODO: save this on disk instead of in-mem?
func getImage(from name.Reference) v1.Image {
	image, err := remote.Image(from, getAuth(from.Context().RegistryStr()))
	if err != nil {
		if _, ok := err.(*remote.ErrSchema1); ok {
			log.Errorf("image %s uses incompatible v1 schema, skipping", from.String())
		} else {
			log.Fatal(err)
		}
	}

	return image
}

func writeImage(to name.Reference, image v1.Image) {
	err := remote.Write(to, image, getAuth(to.Context().RegistryStr()))
	if err != nil {
		log.Error(err)
	}
}

func getAuth(registry string) remote.Option {
	return remote.WithAuth(getCorrectAuth(registry))
}

func findHighestTag(repository string) *semver.Version {
	tags := listTags(repository)
	versions := semVerSort(tags)
	return versions[len(versions)-1]
}

func mirrorImage(from name.Reference, to name.Reference) {
	image := getImage(from)
	if image != nil {
		writeImage(to, image)
	}
}

func listTags(repository string) []string {
	r, err := name.NewRepository(repository)
	if err != nil {
		log.Fatal(err)
	}

	list, err := remote.List(r, getAuth(repository))
	if err != nil {
		log.Fatal(err)
	}

	return list
}

// semVerSort sorts SemVer versions, removes non-SemVer values.
func semVerSort(xs []string) semver.Collection {
	var xv semver.Collection

	for _, v := range xs {
		version, err := semver.NewVersion(v)
		if err != nil {
			if err == semver.ErrInvalidSemVer {
				log.Debugf("%s is not a SemVer version, ignoring", v)
				continue
			} else {
				log.Error(err)
				continue
			}
		}
		// Tags with only numbers will incorrectly be parsed by NewVersion(), do a dirty 'verification' on Major number.
		if version.Major() > 1024 {
			log.Debugf("%s is probably not a SemVer version, ignoring", version.String())
			continue
		}
		// Handle prerelease versions.
		if version.Prerelease() != "" {
			if !viper.GetBool("allow-prerelease") {
				log.Debugf("%s is a prerelease version, ignoring", version.String())
				continue
			}
		}

		xv = append(xv, version)
	}

	sort.Sort(xv)
	return xv
}
