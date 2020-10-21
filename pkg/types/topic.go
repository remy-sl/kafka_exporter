package types

type Topic struct {
	Name       string
	Partitions []Partition
}

type Partition struct {
	ID       int
	Leader   int
	Replicas []int
	Isrs     []int // in-sync replicas
}
