package pubsub

type Action int

const (
	Register Action = iota
	Unregister
	Broadcast
)

type msg struct {
	channel chan interface{}
	data    interface{}
	action  Action
}

type PubSub struct {
	bus         chan msg
	subscribers map[chan interface{}]bool
}

func (p *PubSub) Start() {

	go func() {

		for m := range p.bus {
			switch m.action {
			case Register:
				p.subscribers[m.channel] = true
			case Unregister:
				delete(p.subscribers, m.channel)
			case Broadcast:
				for k := range p.subscribers {
					k <- m.data
				}

			}
		}

	}()

}

func (p *PubSub) Stop() {
	close(p.bus)
}

func (p *PubSub) Register(c chan interface{}) {
	p.bus <- msg{
		channel: c,
		action:  Register,
	}
}

func (p *PubSub) UnRegister(c chan interface{}) {
	p.bus <- msg{
		channel: c,
		action:  Unregister,
	}
}

func (p *PubSub) Broadcast(data interface{}) {
	p.bus <- msg{
		data:   data,
		action: Broadcast,
	}
}

func NewPubSub() *PubSub {
	return &PubSub{
		bus:         make(chan msg),
		subscribers: map[chan interface{}]bool{},
	}
}
