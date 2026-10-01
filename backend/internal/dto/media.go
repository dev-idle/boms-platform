package dto

// CloudinaryUploadSignatureResponse is one signed upload: the browser sends
// Params with the file exactly as given, and checks the image it gets back is
// in Folder.
type CloudinaryUploadSignatureResponse struct {
	CloudName string            `json:"cloud_name"`
	APIKey    string            `json:"api_key"`
	Signature string            `json:"signature"`
	UploadURL string            `json:"upload_url"`
	Folder    string            `json:"folder"`
	Params    map[string]string `json:"params"`
	MaxBytes  int64             `json:"max_bytes"`
}
