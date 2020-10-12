package common

// K8sManifest type for rendered manifest unmarshaled to
type K8sManifest map[string]interface{}

// RAW the key value for making content parsable as K8sManifest
const RAW string = "raw"
