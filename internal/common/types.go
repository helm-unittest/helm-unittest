package common

// K8sManifest type for rendered manifest unmarshalled to
type K8sManifest map[string]interface{}

func (m K8sManifest) ToString() string {
	return TrustedMarshalYAML(m)
}

// RAW the key value for making content parsable as K8sManifest
const RAW string = "raw"

// Override Indention
const YAMLINDENTION int = 2
