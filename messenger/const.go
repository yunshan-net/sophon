package messenger

type MessageType uint8

const (
	MSG_USAGE MessageType = iota
	MSG_PERF
	MSG_GEO
	MSG_FLOW
	MSG_CONSOLE_LOG
	MSG_TYPE
	MSG_FPS

	MSG_VTAP_USAGE
)
