package payment_out_general

type Response401Error struct {
	Name    string `json:"name"`
	Message string `json:"message"`
	Status  int32  `json:"status"`
}

type Response422Error struct {
	Message string `json:"string"`
}

type ResponseGetBalance struct {
	Balance int64 `json:"balance"`
}

type ResponseGetBank []struct {
	BankCode string `json:"bank_code"`
	Name     string `json:"name"`
	Fee      int64  `json:"fee"`
	Queue    int64  `json:"queue"`
	Status   string `json:"status"`
}

type ResponseMaintenanceInfo struct {
	Maintenance string `json:"maintenance"`
}
