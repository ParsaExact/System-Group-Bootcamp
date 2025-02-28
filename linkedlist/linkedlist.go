package linkedlist

type Node struct {
	Value int
	Next  *Node
}

func AddNode(head *Node, value int) *Node {
	if head == nil {
		return &Node{Value: value}
	}

	current := head
	for current.Next != nil {
		current = current.Next
	}
	current.Next = &Node{Value: value}
	return head
}

func MergeNodes(head *Node) *Node {
	if head == nil {
		return nil
	}

	var ans *Node
	sum := 0

	for node := head; node != nil; node = node.Next {
		if node.Value == 0 {
			if sum != 0 {
				ans = AddNode(ans, sum)
			}
			sum = 0
		} else {
			sum += node.Value
		}
	}

	if sum != 0 {
		ans = AddNode(ans, sum)
	}

	return ans
}
