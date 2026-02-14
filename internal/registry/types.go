package registry

import "gopkg.in/yaml.v3"

// ClaimEntry represents a single crossplane claim tracked in the registry.
type ClaimEntry struct {
	Name          string `yaml:"name"`
	Namespace     string `yaml:"namespace"`
	ClaimRef      string `yaml:"claimRef"`
	StatusMessage string `yaml:"statusMessage"`
	LastCheckedAt string `yaml:"lastCheckedAt"`
}

// RegistryFile holds the full registry: a mapping of cluster names to their claim entries.
type RegistryFile struct {
	Clusters map[string][]ClaimEntry
}

func (r *RegistryFile) UnmarshalYAML(value *yaml.Node) error {
	return value.Decode(&r.Clusters)
}

func (r RegistryFile) MarshalYAML() (any, error) {
	return r.Clusters, nil
}
