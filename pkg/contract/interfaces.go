package contract

type Content interface {
	ContentType() uint32

	Marshal() ([]byte, error)
	Unmarshal(payload []byte) error
}
