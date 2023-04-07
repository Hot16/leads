package models

type Lead struct {
	ID        uint32
	VzId      uint32
	AtlId     uint32
	AtlStatus string
	Date      string
	RawData   string
	Phone     string
}
