package responses_client

type FileResponse struct {
	Filename string `json:"filename"`
	MetaHash string `json:"meta-hash"`
	Size     int    `json:"size"`
}
