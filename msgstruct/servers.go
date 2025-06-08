package msgstruct

type ServerType uint32

const (
	Client ServerType = iota + 1
	ApiServer
	PictureServer
)

type MsgType uint32

const (
	MsgTypeNone MsgType = iota + 1
	MsgTypeGetPictureReq
	MsgTypeGetPictureRsp
)

type MsgPack struct {
	MsgType        MsgType    `msgpack:"msg_type"`
	DispatchServer ServerType `msgpack:"to_server"`
	SessionID      uint32     `msgpack:"session_id"`
	Data           []byte     `msgpack:"data"`
}

type MsgRegister struct {
	ServerType ServerType `msgpack:"server_type"`
	Name       string     `msgpack:"name"`
	Key        string     `msgpack:"key"`
}

type MsgTransferPictureReq struct {
	SrcPicture []byte `msgpack:"src"`
	UserKey    string `msgpack:"key"`
	ImgName    string `msgpack:"img_name"`
}

type MsgTransferPictureRsp struct {
	TargetPicture []byte `msgpack:"target"`
	UserKey       string `msgpack:"key"`
	ImgName       string `msgpack:"img_name"`
}
