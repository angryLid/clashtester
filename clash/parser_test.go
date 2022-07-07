package clash

import "testing"

func TestReadConfig(t *testing.T) {
	config := ReadConfig()

	t.Log(config)
}

type Writer interface {
	Write() error
}
type WriterImpl struct {
}

func (w *WriterImpl) Write() error {
	return nil
}

func UseWriter(constructor func() Writer) {

}
func TestI(t *testing.T) {
	UseWriter(func() Writer {
		return &WriterImpl{}
	})
}
