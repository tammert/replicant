package replicant

import (
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/v1/google"
	log "github.com/sirupsen/logrus"
)

func gcrAuthenticator() authn.Authenticator {
	authenticator, err := google.NewEnvAuthenticator()
	if err != nil {
		log.Error(err)
	}

	return authenticator
}
