package env

type Node struct {
	ID      int32
	Address string
}

var Cluster = []Node{
	{
		ID:      0,
		Address: "localhost:10000",
	},
	{
		ID:      1,
		Address: "localhost:10001",
	},
	{
		ID:      2,
		Address: "localhost:10002",
	},
	//{
	//	ID:      3,
	//	Address: "localhost:10003",
	//},
	//{
	//	ID:      4,
	//	Address: "localhost:10004",
	//},
}
