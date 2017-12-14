package madtun_chan

type Hub struct {
	chans map[int32]*Chan
}

func NewHub() *Hub {
	return &Hub{}
}

func (h *Hub) Add(ch *Chan) {

}

func (h *Hub) Find() *Chan {
	return nil
}
