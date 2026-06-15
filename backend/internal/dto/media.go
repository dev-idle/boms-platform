package dto

type CloudinaryUploadSignatureResponse struct {
	CloudName      string `json:"cloud_name"`
	APIKey         string `json:"api_key"`
	Timestamp      int64  `json:"timestamp"`
	Signature      string `json:"signature"`
	Folder         string `json:"folder"`
	UploadURL      string `json:"upload_url"`
	AllowedFormats string `json:"allowed_formats"`
	UniqueFilename string `json:"unique_filename"`
	MaxBytes       int64  `json:"max_bytes"`
}
