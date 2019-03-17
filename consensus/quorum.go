package consensus

func Reached(count, nodesNumber int) bool {
	return count >= (nodesNumber/2)+1
}
