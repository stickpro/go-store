package media

type SaveMediumDTO struct {
	Name     string `json:"name"`
	FileType string `json:"file_type"`
	Path     string `json:"path"`
	Size     int64  `json:"size"`
	Data     []byte `json:"data"`
}
