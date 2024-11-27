package config

// ModuleMetadata is the metadata for the module
type ModuleMetadata struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Version     string `json:"version"`
	Creator     string `json:"creator"`
	Contact     string `json:"contact"`
	Lang        string `json:"lang"`
}

var (
	SupportedLangs = []string{"go", "python"} // Supported languages
)
