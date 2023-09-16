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
	Port               string             `env:"PORT" envDefault:"8080"`
	TlsPort            string             `env:"TLS_PORT" envDefault:"8443"`
	TlsCert            string             `env:"TLS_CERT"`
	TlsKey             string             `env:"TLS_KEY"`
	Name               string             `env:"NAME" envDefault:"lc-metadata"`
	Type               MetadataType       `env:"TYPE" envDefault:"google"`
	LogLevel           string             `env:"LOG_LEVEL" envDefault:"info"`
	ProjectId          string             `env:"PROJECT_ID,notEmpty"`
	CloudKeyfile       string             `env:"CLOUD_KEYFILE"`
	DefaultAccount     string             `env:"DEFAULT_ACCOUNT"`
	AllowOtherProjects bool               `env:"ALLOW_OTHER_PROJECTS" envDefault:"false"`
	LcmNamespace       string             `env:"LCM_NAMESPACE" envDefault:"kube-system"`
	KsaResolver        KsaBindingResolver `env:"KSA_RESOLVER" envDefault:"annotation"`
	KsaVerifyBinding   bool               `env:"KSA_VERIFY_BINDING" envDefault:"true"`
	Google             Google             `env:"GOOGLE"`
}

type Google struct {
	IdentityPool string `env:"GOOGLE_IDENTITY_POOL"`
}

// Initialised by server/run.go
var Current Config
