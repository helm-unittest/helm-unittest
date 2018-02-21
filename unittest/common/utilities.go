package common

import yaml "gopkg.in/yaml.v2"

// TrustedMarshalYAML malshal yaml without error return, if error happen it panic
func TrustedMarshalYAML(d interface{}) string {
	s, err := yaml.Marshal(d)
	if err != nil {
		panic(err)
	}
	return string(s)
}
