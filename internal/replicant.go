package replicant

import (
	"github.com/Masterminds/semver/v3"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/remote/transport"
	log "github.com/sirupsen/logrus"
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
		case "semver":
			mirrorSemVerTags(image)
		case "all":
			mirrorAllTags(image)
		default:
			mirrorHighestTag(image)
		}
	}
}

// mirrorAllTags mirrors all tags.
func mirrorAllTags(ic *ImageConfig) {
	log.Infof("begin mirroring all tags from %s to %s", ic.SourceRepository, ic.DestinationRepository)

	tags := listTags(ic.SourceRepository)
	if len(tags) == 0 {
		noTagsFound(ic.SourceRepository)
		return
	}

	for _, tag := range tags {
		mirrorTag(ic, tag)
	}

	log.Infof("done mirroring all tags from %s to %s", ic.SourceRepository, ic.DestinationRepository)
}

// mirrorSemVerTags mirrors all SemVer tags.
func mirrorSemVerTags(ic *ImageConfig) {
	log.Infof("begin mirroring all SemVer tags from %s to %s", ic.SourceRepository, ic.DestinationRepository)

	tags := listTags(ic.SourceRepository)
	if len(tags) == 0 {
		noTagsFound(ic.SourceRepository)
		return
	}

	sorted := semVerSort(tags, ic.AllowPrerelease)
	if len(sorted) > 0 {
		for _, tag := range sorted {
			mirrorTag(ic, tag.Original())
		}
		log.Infof("done mirroring all SemVer tags from %s to %s", ic.SourceRepository, ic.DestinationRepository)
	} else {
		noSemverTagsFound(ic.SourceRepository)
	}
}

// mirrorHigherTags mirrors all tags that are a higher SemVer version than the highest available in the destination repository.
func mirrorHigherTags(ic *ImageConfig) {
	log.Infof("begin mirroring all tags greater than the highest in %s from %s", ic.DestinationRepository, ic.SourceRepository)

	var tagsToMirror []*semver.Version

	highestDestinationTag := findHighestTag(ic.DestinationRepository, ic.AllowPrerelease)
	if highestDestinationTag == nil {
		log.Infof("no highest tag found in %s, can't determine which tags from %s to mirror", ic.DestinationRepository, ic.SourceRepository)
		return
	}

	tags := listTags(ic.SourceRepository)
	if len(tags) == 0 {
		noTagsFound(ic.SourceRepository)
		return
	}
	sorted := semVerSort(tags, ic.AllowPrerelease)
	if len(sorted) > 0 {
		for _, tag := range sorted {
			if tag.GreaterThan(highestDestinationTag) {
				tagsToMirror = append(tagsToMirror, tag)
			}
		}
		for _, t := range tagsToMirror {
			mirrorTag(ic, t.Original())
		}
		log.Infof("done mirroring all tags greater than the highest in %s from %s", ic.DestinationRepository, ic.SourceRepository)
	} else {
		noSemverTagsFound(ic.SourceRepository)
	}

}

// mirrorHighestTag mirrors the highest SemVer tag, if it's not already in the destination repository.
func mirrorHighestTag(ic *ImageConfig) {
	log.Infof("begin mirroring highest tag from %s to %s", ic.SourceRepository, ic.DestinationRepository)

	tag := findHighestTag(ic.SourceRepository, ic.AllowPrerelease)
	if tag != nil {
		mirrorTag(ic, tag.Original())
		log.Infof("done mirroring highest tag from %s to %s", ic.SourceRepository, ic.DestinationRepository)
	} else {
		log.Infof("no suitable highest tag found in %s", ic.SourceRepository)
	}
}

func mirrorTag(ic *ImageConfig, tag string) {
	sourceReference, err := name.ParseReference(ic.SourceRepository + ":" + tag)
	if err != nil {
		log.Error(err)
	}
	destinationReference, err := name.ParseReference(ic.DestinationRepository + ":" + tag)
	if err != nil {
		log.Error(err)
	}

	_, err = remote.Head(destinationReference, getAuth(destinationReference.Context().RegistryStr()))
	if err != nil {
		if t, ok := err.(*transport.Error); ok {
			if t.StatusCode == 404 {
				log.Debugf("image %s not found, mirroring", destinationReference.String())
				image := getImage(sourceReference)
				if image != nil {
					writeImage(destinationReference, image)
					log.Debugf("mirrored %s into %s", sourceReference.String(), destinationReference.String())
				}
				return
			}
		} else {
			log.Error(err)
			//TODO: explode here
		}
	}

	log.Debugf("image %s already present in destination repository", destinationReference.String())

	// Tag is present in destination repository. If configured with `replace-tag`, check if the image ID matches.
	if ic.ReplaceTag {
		sourceImageID, err := getImage(sourceReference).ConfigName()
		if err != nil {
			log.Error(err)
		}
		destinationImageID, err := getImage(destinationReference).ConfigName()
		if err != nil {
			log.Error(err)
		}

		// Compare the image IDs.
		if sourceImageID == destinationImageID {
			log.Debug("image IDs are identical, no need to replace")
		} else {
			log.Debugf("image IDs are different, replacing %s with %s", destinationImageID, sourceImageID)
			writeImage(sourceReference, getImage(sourceReference))
		}
	}
}

func getImage(from name.Reference) v1.Image {
	image, err := remote.Image(from, getAuth(from.Context().RegistryStr()))
	if err != nil {
		if _, ok := err.(*remote.ErrSchema1); ok {
			log.Errorf("image %s uses incompatible v1 schema, skipping", from.String())
		} else {
			log.Error(err)
			//TODO: explode here
		}
	}

	return image
}

func writeImage(to name.Reference, image v1.Image) {
	err := remote.Write(to, image, getAuth(to.Context().RegistryStr()))
	if err != nil {
		log.Error(err)
		//TODO: explode here
	}
}

func getAuth(registry string) remote.Option {
	return remote.WithAuth(getCorrectAuth(registry))
}

func findHighestTag(repository string, allowPrerelease bool) *semver.Version {
	var versions = semver.Collection{}

	tags := listTags(repository)
	if len(tags) > 0 {
		versions = semVerSort(tags, allowPrerelease)
	} else {
		return nil
	}

	if len(versions) > 0 {
		return versions[len(versions)-1]
	} else {
		return nil
	}
}

func listTags(repository string) []string {
	r, err := name.NewRepository(repository)
	if err != nil {
		log.Fatal(err)
	}

	list, err := remote.List(r, getAuth(r.RegistryStr()))
	if err != nil {
		log.Error(err)
		//TODO: explode here
	}

	return list
}

// semVerSort sorts SemVer versions, removes non-SemVer values.
func semVerSort(xs []string, allowPrerelease bool) semver.Collection {
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
			if !allowPrerelease {
				log.Debugf("%s is a prerelease version, ignoring", version.String())
				continue
			}
		}

		xv = append(xv, version)
	}

	sort.Sort(xv)
	return xv
}

func noTagsFound(s string) {
	log.Infof("no tags found in %s, nothing to mirror", s)
}

func noSemverTagsFound(s string) {
	log.Infof("no SemVer tags to mirror for %s", s)
}
