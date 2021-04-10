package replicant

import (
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/v1/google"
	log "github.com/sirupsen/logrus"
	"strings"
)

type CredentialHelper struct {
	gcrCredentials authn.Authenticator
	ecrCredentials authn.Authenticator
	acrCredentials authn.Authenticator
}

var savedCredentials = &CredentialHelper{}

func getCorrectAuth(registry string) authn.Authenticator {
	switch {
	case strings.Contains(registry, "gcr.io"):
		return gcrAuthenticator()
	case strings.Contains(registry, "dkr.ecr"):
		return ecrAuthenticator()
	case strings.Contains(registry, "azurecr.io"):
		return acrAuthenticator()
	default:
		return authn.Anonymous
	}
}

func gcrAuthenticator() authn.Authenticator {
	// See if we have credentials already saved.
	if savedCredentials.gcrCredentials != nil {
		return savedCredentials.gcrCredentials
	}

	// No credentials saved, create new ones.
	authenticator, err := google.NewEnvAuthenticator()
	if err != nil {
		log.Error(err)
	}
	savedCredentials.gcrCredentials = authenticator
	return authenticator
}

func ecrAuthenticator() authn.Authenticator {
	// See if we have credentials already saved.
	if savedCredentials.ecrCredentials != nil {
		return savedCredentials.ecrCredentials
	}

	// No credentials saved, create new ones.
	authenticator := &authn.Basic{
		Username: "AWS",
		Password: "aws ecr get-login-password --region <region>", //TODO
	}
	savedCredentials.ecrCredentials = authenticator
	return authenticator
}

func acrAuthenticator() authn.Authenticator {
	// See if we have credentials already saved.
	if savedCredentials.acrCredentials != nil {
		return savedCredentials.acrCredentials
	}

	// No credentials saved, create new ones.
	authenticator := &authn.Basic{
		Username: "<sp-app-id>",
		Password: "<sp-password>", //TODO
	}
	savedCredentials.acrCredentials = authenticator
	return authenticator
}
