package response

type ResponseForm struct {
	Status   int32  `json:"status"`
	Services string `json:"service"`
	Message  string `json:"pesan"`
	Payload  any    `json:"response_payload"`
}
