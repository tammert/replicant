package replicant

import (
	"fmt"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/spf13/viper"
	"log"
)

func CloneToRepo() {
	imageUri := viper.GetString("image")

	r, err := name.ParseReference(imageUri)
	if err != nil {
		log.Fatal(err)
	}

	image, err := remote.Image(r)
	if err != nil {
		log.Fatal(err)
	}

	localRef, err := name.ParseReference("localhost:5000/" + imageUri)
	if err != nil {
		log.Fatal(err)
	}

	err = remote.Write(localRef, image)
	if err != nil {
		log.Fatal(err)
	}
}

func ListTags() {
	repository, err := name.NewRepository(viper.GetString("repository"))
	if err != nil {
		log.Fatal(err)
	}

	list, err := remote.List(repository)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(list)
}
