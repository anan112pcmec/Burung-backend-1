package response

type ResponseForm struct {
	Status   int32  `json:"status"`
	Services string `json:"service"`
	Payload  any
}
