package registry

import (
	"time"

	"gopkg.in/yaml.v3"
)

// ParseRegistry deserializes YAML bytes into a RegistryFile.
func ParseRegistry(data []byte) (*RegistryFile, error) {
	var reg RegistryFile
	if err := yaml.Unmarshal(data, &reg); err != nil {
		return nil, err
	}
	if reg.Clusters == nil {
		reg.Clusters = make(map[string][]ClaimEntry)
	}
	return &reg, nil
}

// UpdateClaimStatus finds a claim by cluster and claimRef, then updates its
// StatusMessage and LastCheckedAt. Returns true if a matching entry was found.
func UpdateClaimStatus(reg *RegistryFile, cluster, claimRef, status string) bool {
	claims, ok := reg.Clusters[cluster]
	if !ok {
		return false
	}
	for i := range claims {
		if claims[i].ClaimRef == claimRef {
			claims[i].StatusMessage = status
			claims[i].LastCheckedAt = time.Now().UTC().Format(time.RFC3339)
			reg.Clusters[cluster] = claims
			return true
		}
	}
	return false
}

// SerializeRegistry marshals a RegistryFile back to YAML bytes.
func SerializeRegistry(reg *RegistryFile) ([]byte, error) {
	return yaml.Marshal(reg)
}
