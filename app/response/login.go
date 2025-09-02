package response

type LoginResponse struct {
	Nama     string `json:"nama"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Status   string `json:"status"`
}

type LoginUserResp struct {
	ID int64 `json:"Id_user"`
	LoginResponse
}

type LoginSellerResp struct {
	ID int32
	LoginResponse
	Jenis string
}
