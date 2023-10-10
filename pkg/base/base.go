package base

type Base interface {
	Save(key string, value string, bucket string) error
	Get(key string, bucket string) (string, error)
	SaveBytes(key string, value []byte, bucket string) error
	GetBytes(key string, bucket string) ([]byte, error)
	GettAll(bucket string) ([][]byte, error)
}
