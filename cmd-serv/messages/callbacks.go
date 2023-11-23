package messages

type Callbacks struct {
	OnMessageReceived func(clientID string, msg []byte)
	OnMessageConsumed func(msg []byte)
}
