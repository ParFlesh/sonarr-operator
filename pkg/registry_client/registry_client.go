package registry_client

import (
	"github.com/docker/distribution/manifest/schema2"
	"github.com/heroku/docker-registry-client/registry"
	"strings"
)

const (
	// DefaultTag defines the default tag used when performing images related actions and no tag or digest is specified
	DefaultTag = "latest"
	// DefaultHostname is the default built-in hostname
	DefaultHostname = "docker.io"
	// DefaultRepoPrefix is the prefix used for default repositories in default host
	DefaultRepoPrefix = "library"
)

type RegistryClientProviderInterface interface {
	New(registryURL, username, password string) (RegistryClientInterface, error)
}

type RegistryClientProvider struct{}

func (r *RegistryClientProvider) New(registryURL, username, password string) (RegistryClientInterface, error) {
	client, err := registry.New(registryURL, username, password)
	if err != nil {
		return &RegistryClient{}, err
	}

	return &RegistryClient{client: client}, nil
}

type MockRegistryClientProvider struct {
	Client RegistryClientInterface
}

func (r *MockRegistryClientProvider) New(registryURL, username, password string) (RegistryClientInterface, error) {
	return r.Client, nil
}

type RegistryClientInterface interface {
	ManifestV2(repository, reference string) (*schema2.DeserializedManifest, error)
	Tags(repository string) (tags []string, err error)
}

type RegistryClient struct {
	client *registry.Registry
}

func (r *RegistryClient) ManifestV2(repository, reference string) (*schema2.DeserializedManifest, error) {
	return r.client.ManifestV2(repository, reference)
}

func (r *RegistryClient) Tags(repository string) (tags []string, err error) {
	return r.client.Tags(repository)
}

type MockRegistryClient struct {
	ManifestV2Output *schema2.DeserializedManifest
	ManifestV2Error  error
	TagsOutput       []string
	TagsError        error
}

func (r *MockRegistryClient) ManifestV2(repository, reference string) (*schema2.DeserializedManifest, error) {
	return r.ManifestV2Output, r.ManifestV2Error
}

func (r *MockRegistryClient) Tags(repository string) (tags []string, err error) {
	return r.TagsOutput, r.TagsError
}

func SplitImageName(imageName string) (string, string, string) {
	nameParts := strings.Split(imageName, "/")
	var repo, user, image, tag string
	if len(nameParts) > 2 {
		repo = strings.Join(nameParts[:len(nameParts)-2], "/")
		user = nameParts[len(nameParts)-2]
		image = nameParts[len(nameParts)-1]
	} else if len(nameParts) == 2 {
		repo = DefaultHostname
		user = nameParts[0]
		image = nameParts[1]
	} else if len(nameParts) == 1 {
		repo = DefaultHostname
		user = DefaultRepoPrefix
		image = nameParts[0]
	}

	imageParts := strings.Split(image, ":")
	if len(imageParts) == 2 {
		image = imageParts[0]
		tag = imageParts[1]
	} else if len(imageParts) == 1 {
		image = imageParts[0]
		tag = DefaultTag
	}

	return repo, strings.Join([]string{user, image}, "/"), tag
}
