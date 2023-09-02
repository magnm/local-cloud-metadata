package config

type MetadataType string

const (
	GoogleMetadata MetadataType = "google"
)

type KsaBindingResolver string

const (
	// KsaBindingResolverAnnotation means finding the cloud service account by looking at the annotations of the KSA.
	KsaBindingResolverAnnotation KsaBindingResolver = "annotation"
	// KsaBindingResolverCRD means using an in-cluster CRD to resolve the cloud service account bound to the KSA.
	KsaBindingResolverCRD KsaBindingResolver = "crd"
)

type Config struct {
	Port         string             `env:"PORT" envDefault:"8080"`
	TlsPort      string             `env:"TLS_PORT" envDefault:"8443"`
	TlsCert      string             `env:"TLS_CERT"`
	TlsKey       string             `env:"TLS_KEY"`
	Type         MetadataType       `env:"TYPE" envDefault:"google"`
	ProjectId    string             `env:"PROJECT_ID,notEmpty"`
	CloudKeyfile string             `env:"CLOUD_KEYFILE"`
	KsaResolver  KsaBindingResolver `env:"KSA_RESOLVER" envDefault:"annotation"`
}

// Initialised by server/run.go
var Current Config
