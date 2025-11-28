package pengiriman_enums

// Enums jenis pengiriman
var (
	Reguler = "Reguler"
	Fast    = "Fast"
	Instant = "Instant"
)

// Enums untuk pengiriman non ekspedisi
var (
	Waiting      = "Waiting"
	PickedUp     = "Picked Up"
	Diperjalanan = "Diperjalanan"
	Sampai       = "Sampai"
	Trouble      = "Trouble"
)

// Enums untuk pengiriman ekspedisi
var (
	WaitingEkspedisi           = "Waiting"
	DikirimEkspedisi           = "Dikirim"
	SampaiAgentEkspedisi       = "Sampai Agent"
	SampaiAgentTujuanEkspedisi = "Sampai Agent Tujuan"
	DikirimAgentEkspedisi      = "Dikirim Agent"
	SampaiEkspedisi            = "Sampai"
)
