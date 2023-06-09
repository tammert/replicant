package replicant

import (
	"encoding/base64"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/v1/google"
	log "github.com/sirupsen/logrus"
	"os"
	"strings"
)

type CredentialHelper struct {
	gcpCredentials authn.Authenticator
	ecrCredentials authn.Authenticator
	acrCredentials authn.Authenticator
}

var savedCredentials = &CredentialHelper{}

func getCorrectAuth(registry string) authn.Authenticator {
	switch {
	case strings.Contains(registry, "gcr.io"):
		return gcpAuthenticator()
	case strings.Contains(registry, "pkg.dev"):
		return gcpAuthenticator()
	case strings.Contains(registry, "dkr.ecr"):
		return ecrAuthenticator()
	case strings.Contains(registry, "azurecr.io"):
		return acrAuthenticator()
	default:
		return authn.Anonymous
	}
}

func gcpAuthenticator() authn.Authenticator {
	// See if we have credentials already saved.
	if savedCredentials.gcpCredentials != nil {
		return savedCredentials.gcpCredentials
	}

	// No credentials saved, create new ones.
	authenticator, err := google.NewEnvAuthenticator()
	if err != nil {
		log.Fatal(err)
	}
	savedCredentials.gcpCredentials = authenticator
	return authenticator
}

func ecrAuthenticator() authn.Authenticator {
	// See if we have credentials already saved.
	if savedCredentials.ecrCredentials != nil {
		return savedCredentials.ecrCredentials
	}

	// Needs AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY and AWS_DEFAULT_REGION
	s := session.Must(session.NewSession())
	svc := ecr.New(s)

	input := &ecr.GetAuthorizationTokenInput{}
	token, err := svc.GetAuthorizationToken(input)
	if err != nil {
		log.Fatal(err)
	}

	var userName, password string
	for _, data := range token.AuthorizationData {
		decoded, err := base64.StdEncoding.DecodeString(*data.AuthorizationToken)
		if err != nil {
			log.Fatal(err)
		}
		splitter := strings.Split(string(decoded), ":")
		userName = splitter[0]
		password = splitter[1]
	}

	// No credentials saved, create new ones.
	authenticator := &authn.Basic{
		Username: userName,
		Password: password,
	}
	savedCredentials.ecrCredentials = authenticator
	return authenticator
}

func acrAuthenticator() authn.Authenticator {
	// See if we have credentials already saved.
	if savedCredentials.acrCredentials != nil {
		return savedCredentials.acrCredentials
	}

	userName, ok := os.LookupEnv("AZURE_SP_ID")
	if !ok {
		log.Fatal("AZURE_SP_ID is not set")
	}
	password, ok := os.LookupEnv("AZURE_SP_PASSWORD")
	if !ok {
		log.Fatal("AZURE_SP_PASSWORD is not set")
	}

	// No credentials saved, create new ones.
	authenticator := &authn.Basic{
		Username: userName,
		Password: password,
	}
	savedCredentials.acrCredentials = authenticator
	return authenticator
}
