package cloud

type Cloud interface {
	UploadImage([]byte, string) error
	ObjectRecognition(string) ([]string, error)
	GetURL(string) (string, error)
	Close()
}
