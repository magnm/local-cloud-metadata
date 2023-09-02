package config

type MetadataType string

const (
	GoogleMetadata MetadataType = "google"
)

type KsaBindingResolver string

const (
	// KsaBindingResolverCRD means using an in-cluster CRD to resolve the cloud service account bound to the KSA.
	KsaBindingResolverCRD KsaBindingResolver = "crd"
	// KsaBindingResolverCloud means finding the cloud service account by querying the cloud's IAM policy.
	KsaBindingResolverCloud KsaBindingResolver = "cloud"
)

type Config struct {
	Port       string       `env:"PORT" envDefault:"8080"`
	Type       MetadataType `env:"TYPE" envDefault:"google"`
	ProjectId  string       `env:"PROJECT_ID,notEmpty"`
	KsaBinding string       `env:"KSA_BINDING" envDefault:"crd"`
}

// Initialised by server/run.go
var Current Config
