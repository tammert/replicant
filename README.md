# replicant
Replicant mirrors container images between repositories, using the [go-containerregistry](https://github.com/google/go-containerregistry) library. This means a Docker daemon is **not needed** for Replicant to mirror images; the library interacts directly with the registry APIs.

## Configuration options (per source)
* `source`: the 'from' repository
* `destination`: the 'to' repository
* `mode`: see *mirroring modes* below
* `allow-prerelease`: if `true`, prerelease versions (as per the SemVer specification) will also be eligible for mirroring for SemVer *modes*
* `replace-tag`: if `true`, will check if the image ID for an equal tag is the same for the source and the destination. If not, will replace the tag in the destination repository

### Mirroring modes
Replicant supports 4 types of mirroring in the `mode` field:
1) `highest`: the highest SemVer image tag is mirrored
2) `higher`: all SemVer image tags greater than the highest in the destination repository are mirrored
3) `semver`: all SemVer image tags are mirrored
4) `all`: **all** image tags are mirrored

## Configuration file
Replicant needs a configuration file to function, in the following format:
```yaml
images:
  image-name-1:
    source: docker.io/project/image
    destination: gcr.io/private-repo/project/image
    mode: highest|higher|semver|all # when not specified, defaults to `highest`
    allow-prerelease: true|false # when not specified, defaults to false
    replace-tag: true|false # when not specified, defaults to false
  image-name-2:
    ...
  image-name-n:
    ...
```

## Global configuration options
Global configuration can either be set via environment variables or as long/short flags to the binary:

|Description|Environment variable|Long flag|Short flag|Type|Default|
|---|---|---|---|---|---|
|Reference to the YAML config file|REPLICANT_CONFIG_FILE|--config-file|-c|string|/config/replicant.yaml|
|Enable debug logging|REPLICANT_DEBUG|--debug|-d|bool|false|

## Supported registries
* read from any repository anonymously
* read/write from private Google Container Registry (GCR)
* read/write from private Azure Container Registry (ACR) [untested]
* read/write from private Amazon Elastic Container Registry (ECR) [untested]

Currently, only one set of credentials per registry *type* is supported. This means that, for example, all configured GCR repositories in both `source` and `destination` will use the same credentials. Furthermore, at the current time, GCR/ACR/ECR will always require valid credentials, even when they're publicly readable.

### Registry authentication
#### GCR
Replicant uses `NewEnvAuthenticator()` to get credentials automatically. More information on that [here](https://cloud.google.com/docs/authentication/production#automatically). When running on GKE, you can combine the above with [Workload Identity](https://cloud.google.com/kubernetes-engine/docs/how-to/workload-identity).
#### ACR
Supports logging in with [service principal + password](https://docs.microsoft.com/en-us/azure/container-registry/container-registry-auth-service-principal#authenticate-with-the-service-principal). Set `AZURE_SP_ID` and `AZURE_SP_PASSWORD` in your environment with the correct values.
#### ECR
Replicant uses [aws-sdk-go](https://github.com/aws/aws-sdk-go) to grab a short-lived (12 hour) token to [authenticate with ECR](https://docs.aws.amazon.com/AmazonECR/latest/userguide/registry_auth.html#registry-auth-token). Set `AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY` and `AWS_DEFAULT_REGION` in your environment with the correct values.

## Running Replicant
1) Build from source: `make build-binary`
2) Download the binary from a GitHub release
3) Use the official [Docker image](https://hub.docker.com/r/tammert/replicant)
4) Use the official [Helm chart](https://github.com/tammert/helm-charts/tree/main/replicant)

## Gotchas
* doesn't work with Docker image V1 schema
* all images are kept in memory, so Replicant will need *at least* as much memory as the (compressed?) size of the largest image you want to mirror
* Docker Hub has rate limit for pulls in place; take this into account when selecting the mode when the source is Docker Hub
