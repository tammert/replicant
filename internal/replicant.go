package replicant

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

// mirrorSemVerTags mirrors all SemVer tags.
func mirrorSemVerTags(config Config) {
	for _, image := range config.Images {
		tags := semVerSort(listTags(image.UpstreamRepository))
		for _, tag := range tags {
			log.Debug(tag)
		}
	}
}

// mirrorHigherTags mirrors all tags that are a higher SemVer version than the highest available in the downstream repository.
func mirrorHigherTags(config Config) {

}

// mirrorHighestTag mirrors the highest SemVer tag, if it's not already in the downstream repository.
func mirrorHighestTag(config Config) {
	for _, image := range config.Images {
		tag := findHighestTag(image.UpstreamRepository)

		upstreamReference, err := name.ParseReference(image.UpstreamRepository + ":" + tag.Original())
		if err != nil {
			log.Error(err)
		}
		downstreamReference, err := name.ParseReference(image.DownstreamRepository + ":" + tag.Original())
		if err != nil {
			log.Error(err)
		}

		_, err = remote.Head(downstreamReference, getAuth(downstreamReference))
		if err != nil {
			t, ok := err.(*transport.Error)
			if ok {
				if t.StatusCode == 404 {
					log.Debug("image not found downstream, mirroring")
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
}

func getImage(from name.Reference) v1.Image {
	image, err := remote.Image(from, getAuth(from))
	if err != nil {
		log.Error(err)
	}

	return image
}

func writeImage(to name.Reference, image v1.Image) {
	err := remote.Write(to, image, getAuth(to))
	if err != nil {
		log.Error(err)
	}
}

func getAuth(ref name.Reference) remote.Option {
	return remote.WithAuth(getCorrectAuth(ref.Context().RegistryStr()))
}

func findHighestTag(repository string) *semver.Version {
	tags := listTags(repository)
	versions := semVerSort(tags)
	return versions[len(versions)-1]
}

func mirrorImage(from name.Reference, to name.Reference) {
	writeImage(to, getImage(from))
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
