package open_route_matrix

type RequestMatrix struct {
	Lokasi  [][2]float64 `json:"locations"`
	Metrics []string     `json:"metrics,omitempty"` // misal: ["distance", "duration"]
	Satuan  string       `json:"units,omitempty"`
}

type ResponseMatrix struct {
	Distances [][]float64 `json:"distances"`
	Durations [][]float64 `json:"durations"`
	Sources   []struct {
		Location        [2]float64 `json:"location"`
		SnappedDistance float64    `json:"snapped_distance"`
	} `json:"sources"`
	Destinations []struct {
		Location        [2]float64 `json:"location"`
		SnappedDistance float64    `json:"snapped_distance"`
	} `json:"destinations"`
}
