package payment_va

import (
	"fmt"
	"strings"
)

func ParseVirtualAccount(data any) (string, error) {
	switch v := data.(type) {
	case map[string]interface{}:
		if _, ok := v["bca_va_number"]; ok {
			fmt.Println("Detected: BCA Virtual Account (via map)")
			return "bca", nil
		}
		if _, ok := v["permata_va_number"]; ok {
			fmt.Println("Detected: Permata Virtual Account (via map)")
			return "permata", nil
		}

		if vaNums, ok := v["va_numbers"].([]interface{}); ok && len(vaNums) > 0 {
			if first, ok := vaNums[0].(map[string]interface{}); ok {
				if bank, ok := first["bank"].(string); ok {
					bank = strings.ToLower(bank)
					fmt.Printf("Detected: %s Virtual Account (via map)\n", strings.ToUpper(bank))
					return bank, nil
				}
			}
		}

		return "", fmt.Errorf("map tidak cocok dengan format VA yang dikenal: %v", v)

	case BcaVirtualAccountResponse:
		fmt.Println("Detected: BCA Virtual Account (via struct)")
		return "bca", nil

	case BniVirtualAccountResponse:
		fmt.Println("Detected: BNI Virtual Account (via struct)")
		return "bni", nil

	case BriVirtualAccountResponse:
		fmt.Println("Detected: BRI Virtual Account (via struct)")
		return "bri", nil

	case PermataVirtualAccount:
		fmt.Println("Detected: Permata Virtual Account (via struct)")
		return "permata", nil

	default:
		return "", fmt.Errorf("tipe tidak dikenali: %T", data)
	}
}
