package core

/*(NULL)*/
type Server interface {
	Init()
	Listen()
}

func NewServer(node Node, looper MessageLooper) Server {
	return &server{node: node, looper: looper}
}

/*(NULL)*/
type server struct {
	node   Node
	looper MessageLooper
}

func (s *server) Init() {
}
func (s *server) Listen() {
}
