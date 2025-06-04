package interface_struct

const (
	PictureServer uint32 = iota + 1
	ClientServer
)

type ServerInterface interface {
	Listen(port string)
	Loop()
	Close()
}
