package emote

// Emote represents a emote that has an image
type Emote struct {
	Verb                string
	SenderMessage       string
	SenderDescription   string
	ReceiverMessage     string
	ReceiverDescription string
}

// Gif represents a emote image
type Gif struct {
	Verb string
	URL  string
}
