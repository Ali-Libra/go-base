package msgstruct

const (
	MsgTypeNone uint32 = iota + 1
	MsgTypeGetPictureReq
	MsgTypeGetPictureRsp
)

type MsgPack struct {
	MsgType   uint32 `msgpack:"msg_type"`
	SessionID uint32 `msgpack:"session_id"`
	Data      []byte `msgpack:"data"`
}

type MsgRegister struct {
	Name string `msgpack:"name"`
	Key  string `msgpack:"key"`
}

type MsgGetPictureReq struct {
	SrcPicture []byte `msgpack:"src_picture"`
}

type MsgGetPictureRsp struct {
	TargetPicture []byte `msgpack:"target_picture"`
}
