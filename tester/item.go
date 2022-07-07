package tester

const (
	ItemLatency = 1 << iota
	ItemUpload
	ItemDownload
	ItemUserInfo
)

type ItemHandler struct {
	Item    int
	Handler func() error
}
