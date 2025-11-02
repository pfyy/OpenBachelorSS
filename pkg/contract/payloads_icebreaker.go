package contract

var iceBreakerMessageRegistry = make(map[uint32]func() Content)

func init() {

}
