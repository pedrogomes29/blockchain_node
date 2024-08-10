package merkle_tree

import (
	"encoding/hex"
	"fmt"
	"testing"
)

func TestNewMerkleNode(t *testing.T) {
	data := [][]byte{
		[]byte("node1"),
		[]byte("node2"),
		[]byte("node3"),
	}

	// Level 1
	n1 := NewMerkleNode(nil, nil, data[0])
	n2 := NewMerkleNode(nil, nil, data[1])
	n3 := NewMerkleNode(nil, nil, data[2])
	n4 := NewMerkleNode(nil, nil, data[2])

	// Level 2
	n5 := NewMerkleNode(n1, n2, nil)
	n6 := NewMerkleNode(n3, n4, nil)

	// Level 3
	n7 := NewMerkleNode(n5, n6, nil)

	level1Node1Hash := "64b04b718d8b7c5b6fd17f7ec221945c034cfce3be4118da33244966150c4bd4"
	level1Node2Hash := "08bd0d1426f87a78bfc2f0b13eccdf6f5b58dac6b37a7b9441c1a2fab415d76c"
	RootNodeHash := "4e3e44e55926330ab6c31892f980f8bfd1a6e910ff1ebc3f778211377f35227e"

	if level1Node1Hash != hex.EncodeToString(n5.Data){
		t.Fatalf("Level 1 hash 1 is incorrect")

	}
	if level1Node2Hash != hex.EncodeToString(n6.Data){
		t.Fatalf("Level 1 hash 1 is incorrect")

	}
	if RootNodeHash != hex.EncodeToString(n7.Data){
		t.Fatalf("Root hash is incorrect")
	}
}

func TestNewMerkleTree(t *testing.T) {
	data := [][]byte{
		[]byte("node1"),
		[]byte("node2"),
		[]byte("node3"),
	}
	// Level 1
	n1 := NewMerkleNode(nil, nil, data[0])
	n2 := NewMerkleNode(nil, nil, data[1])
	n3 := NewMerkleNode(nil, nil, data[2])
	n4 := NewMerkleNode(nil, nil, data[2])

	// Level 2
	n5 := NewMerkleNode(n1, n2, nil)
	n6 := NewMerkleNode(n3, n4, nil)

	// Level 3
	n7 := NewMerkleNode(n5, n6, nil)

	rootHash := fmt.Sprintf("%x", n7.Data)
	mTree := NewMerkleTree(data)

	if rootHash != fmt.Sprintf("%x", mTree.RootNode.Data){
		t.Fatalf("Root hash is incorrect")
	}
}