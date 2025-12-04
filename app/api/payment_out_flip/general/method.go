package payment_out_general

import "fmt"

type ValidasiResponse interface {
	Validation() bool
}

func (Response401Error) Validation() bool {
	return false
}

func (Response422Error) Validation() bool {
	return false
}

func (ResponseGetBalance) Validation() bool {
	return true
}

func (ResponseGetBank) Validation() bool {
	return true
}

func (ResponseMaintenanceInfo) Validation() bool {
	return true
}

func HandleResponse[T ValidasiResponse](resp T) error {
	if resp.Validation() {
		fmt.Println("Valid response")
		return nil
	} else {
		fmt.Println("Error response")
		return fmt.Errorf("gagal bukan data response berhasil")
	}
}
