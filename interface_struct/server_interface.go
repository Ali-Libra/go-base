package interface_struct

type ServerInterface interface {
	Listen(port string)
	Loop()
	Close()
}
