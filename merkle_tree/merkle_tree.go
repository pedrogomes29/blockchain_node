package merkle_tree

import (
	"crypto/sha256"
)

type MerkleTree struct {
	RootNode *MerkleNode
}

type MerkleNode struct {
	Left  *MerkleNode
	Right *MerkleNode
	Data  []byte
}

// each level at the merkle tree needs to have an even number of nodes
func ensureEven(level []*MerkleNode) []*MerkleNode{
	if len(level)%2!=0{
		level = append(level,level[len(level)-1])
	}
	return level;
}

func NewMerkleTree(data [][]byte) *MerkleTree{
	var currentLevelNodes []*MerkleNode

	maxNumberOfNodes := len(data)

	if len(data)%2!=0{
		maxNumberOfNodes++
	}

	currentLevelNodes = make([]*MerkleNode,0,maxNumberOfNodes)

	for _, datum := range data {
		node := NewMerkleNode(nil,nil,datum)
		currentLevelNodes = append(currentLevelNodes, node)
	}

	for len(currentLevelNodes) > 1 { //if the current level has only one node, it is the merkle tree root
		currentLevelNodes = ensureEven(currentLevelNodes)
		for currentLevelNodeIdx := 0; currentLevelNodeIdx < len(currentLevelNodes); currentLevelNodeIdx+=2 {
			leftChild,rightChild := currentLevelNodes[currentLevelNodeIdx], currentLevelNodes[currentLevelNodeIdx+1]
			node := NewMerkleNode(leftChild,rightChild,nil)
			currentLevelNodes[currentLevelNodeIdx/2] = node
		}
		currentLevelNodes = currentLevelNodes[:len(currentLevelNodes)/2]
	}

	return &MerkleTree{currentLevelNodes[0]};
}

func NewMerkleNode(left, right *MerkleNode, data []byte) *MerkleNode {
	newNode := MerkleNode{}

	if left == nil && right == nil { //no children, node is a leaf
		newNodeData := sha256.Sum256(data)
		newNode.Data = newNodeData[:]
	} else { //at least one child, node is not a leaf
		prevHashes := append(left.Data, right.Data...)
		newNodeData := sha256.Sum256(prevHashes)
		newNode.Data = newNodeData[:]
	}

	newNode.Left = left
	newNode.Right = right

	return &newNode
}