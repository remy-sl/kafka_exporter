package types

type Offset struct {
	Topic      string
	Group      string // optional; used only when getting offsets for a consumer group
	Partitions map[int]LowHighOffset
}

type LowHighOffset struct {
	Low  int64 // the earliest offset
	High int64 // the latest offset
}
