package kontainerdrivermetadata

import "encoding/json"

const (
	DataJSONLocation = "/var/lib/rancher-data/driver-metadata/data.json"
)

type Data struct {
	// K3S specific data, opaque and defined by the config file in kdm
	K3S map[string]interface{} `json:"k3s,omitempty"`
	// Rke2 specific data, defined by the config file in kdm
	RKE2 map[string]interface{} `json:"rke2,omitempty"`
}

func FromData(b []byte) (Data, error) {
	d := &Data{}

	if err := json.Unmarshal(b, d); err != nil {
		return Data{}, err
	}
	return *d, nil
}
