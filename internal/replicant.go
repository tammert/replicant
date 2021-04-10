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

// mirrorHighestTag mirrors the highest SemVer tag.
func mirrorHighestTag(config Config) {
	for _, image := range config.Images {
		tag := findHighestTag(image.UpstreamRepository)

		upstream := image.UpstreamRepository + ":" + tag.Original()
		upstreamReference, err := name.ParseReference(upstream)
		if err != nil {
			log.Error(err)
		}

		downstream := image.DownstreamRepository + ":" + tag.Original()
		downstreamReference, err := name.ParseReference(downstream)
		if err != nil {
			log.Error(err)
		}

		auth := remote.WithAuth(getCorrectAuth(downstreamReference.Context().RegistryStr()))
		_, err = remote.Head(downstreamReference, auth)
		if err != nil {
			t, ok := err.(*transport.Error)
			if ok {
				if t.StatusCode == 404 {
					log.Debug("image not found downstream, mirroring")
					cloneToRepo(upstreamReference, downstreamReference)
					return
				}
			} else {
				log.Error(err)
			}
		}

		// Image is present downstream.
		if viper.GetBool("replace-tag") { // No need to do all this if you're not going to replace the image anyway!
			upstreamImage := getImage(upstreamReference)
			upstreamImageID, err := upstreamImage.ConfigName()
			if err != nil {
				log.Error(err)
			}

			downstreamImage := getImage(downstreamReference)
			downstreamImageID, err := downstreamImage.ConfigName()
			if err != nil {
				log.Error(err)
			}

			// Compare the image IDs.
			if upstreamImageID == downstreamImageID {
				log.Debug("images are identical!")
			} else {
				cloneToRepo(upstreamReference, downstreamReference) //TODO: don't get the image here again, pass it along
			}
		}
	}
}

func getImage(reference name.Reference) v1.Image {
	auth := remote.WithAuth(getCorrectAuth(reference.Context().RegistryStr()))
	image, err := remote.Image(reference, auth)
	if err != nil {
		log.Error(err)
	}

	return image
}

func findHighestTag(repository string) *semver.Version {
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
