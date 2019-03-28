package env

type Node struct {
	ID      int32
	Address string
}

var Cluster = []Node{
	{
		ID:      0,
		Address: "localhost:4000",
	},
	{
		ID:      1,
		Address: "localhost:4001",
	},
	{
		ID:      2,
		Address: "localhost:4002",
	},
	//{
	//	ID:	3,
	//	Address: "localhost:4003",
	//},
	//{
	//	ID:	4,
	//	Address: "localhost:4004",
	//},
}
