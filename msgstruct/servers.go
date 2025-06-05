package msgstruct

type ServerType uint32

const (
	ClientServer ServerType = iota + 1
	PictureServer
)

type MsgType uint32

const (
	MsgTypeNone MsgType = iota + 1
	MsgTypeGetPictureReq
	MsgTypeGetPictureRsp
)

type MsgPack struct {
	MsgType   MsgType `msgpack:"msg_type"`
	SessionID uint32  `msgpack:"session_id"`
	Data      []byte  `msgpack:"data"`
}

type MsgRegister struct {
	ServerType ServerType `msgpack:"server_type"`
	Name       string     `msgpack:"name"`
	Key        string     `msgpack:"key"`
}

type MsgTransferPictureReq struct {
	SrcPicture []byte `msgpack:"src"`
}

type MsgTransferPictureRsp struct {
	TargetPicture []byte `msgpack:"target"`
}
