package msgstruct

type ServerRegister struct {
	Name string `msgpack:"name"`
	Key  string `msgpack:"key"`
}
