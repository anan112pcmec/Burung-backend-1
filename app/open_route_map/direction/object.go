package open_route_direction

type RequestDirection struct {
	Koordinat    [][2]float64 `json:"coordinates"`
	Bahasa       string       `json:"language,omitempty"`
	Satuan       string       `json:"units,omitempty"` // "m" atau "km"
	Geometry     bool         `json:"geometry,omitempty"`
	Instructions bool         `json:"instructions,omitempty"`
}

type ResponseDirection struct {
	Routes []struct {
		Summary struct {
			Distance float64 `json:"distance"`
			Duration float64 `json:"duration"`
		} `json:"summary"`
		Geometry string `json:"geometry"`
	} `json:"routes"`
	Metadata struct {
		Service string `json:"service"`
	} `json:"metadata"`
}
