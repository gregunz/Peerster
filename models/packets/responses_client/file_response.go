package responses_client

type FileResponse struct {
	Filename string `json:"filename"`
	MetaHash string `json:"meta-hash"`
	Size     uint64 `json:"size"`
}
