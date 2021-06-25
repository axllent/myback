package client

// GHMDDatabase struct
type GHMDDatabase struct {
	Name string `json:"name"`
}

// Database struct
type Database struct {
	Database string `json:"database"`
	Create   string `json:"create"`
	Tables   []struct {
		Name       string `json:"name"`
		Type       string `json:"Type"`
		Rows       int    `json:"rows"`
		CreateTime string `json:"create_time"`
		Checksum   int    `json:"checksum"`
		Create     string `json:"create"`
		CreateHash string `json:"create_hash"`
	} `json:"tables"`
}
