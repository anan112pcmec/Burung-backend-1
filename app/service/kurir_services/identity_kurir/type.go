package identity_kurir

type IdentitasKurir struct {
	IDKurir  int64  `json:"id_kurir"`
	Username string `json:"username_kurir"`
	Email    string `json:"email_kurir"`
}

func (i IdentitasKurir) Validate() bool {
	if i.IDKurir == 0 {
		return false
	}

	if i.Username == "" {
		return false
	}

	if i.Email == "" {
		return false
	}

	return true
}
