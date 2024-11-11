package hub

type Room struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Players  []string `json:"players"`
	Capacity int      `json:"capacity"`
}

type Message struct {
	Type      string      `json:"type"`
	Room      string      `json:"room"`
	Username  string      `json:"username"`
	Content   interface{} `json:"content"`
	Timestamp int64       `json:"timestamp"`
}
