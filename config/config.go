package config

type MetadataType string

const (
	GoogleMetadata MetadataType = "google"
)

type Config struct {
	Port      string       `env:"PORT" envDefault:"8080"`
	Type      MetadataType `env:"TYPE" envDefault:"google"`
	ProjectId string       `env:"PROJECT_ID"`
}

// Initialised by server/run.go
var Current Config
