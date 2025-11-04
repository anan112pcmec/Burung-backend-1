package open_route_direction

type RequestDirection struct {
	Koordinat    [][2]float64 `json:"coordinates"`
	Bahasa       string       `json:"language,omitempty"`
	Satuan       string       `json:"units,omitempty"` // "m" atau "km"
	Geometry     bool         `json:"geometry,omitempty"`
	Instructions bool         `json:"instructions,omitempty"`
}

type ResponseDirection struct {
	Type     string `json:"type"`
	Features []struct {
		BBox       []float64 `json:"bbox"`
		Type       string    `json:"type"`
		Properties struct {
			Segments []struct {
				Distance float64 `json:"distance"`
				Duration float64 `json:"duration"`
				Steps    []struct {
					Distance    float64 `json:"distance"`
					Duration    float64 `json:"duration"`
					Type        int     `json:"type"`
					Instruction string  `json:"instruction"`
					Name        string  `json:"name"`
					WayPoints   []int   `json:"way_points"`
				} `json:"steps"`
			} `json:"segments"`
			Summary struct {
				Distance float64 `json:"distance"`
				Duration float64 `json:"duration"`
			} `json:"summary"`
			WayPoints []int `json:"way_points"`
		} `json:"properties"`
		Geometry struct {
			Coordinates [][]float64 `json:"coordinates"`
			Type        string      `json:"type"`
		} `json:"geometry"`
	} `json:"features"`
	Metadata struct {
		Service string `json:"service"`
	} `json:"metadata"`
}
