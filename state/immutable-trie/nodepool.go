package itrie

import (
	"runtime"
	"sync"
)

const (
	batchAlloc  = 1024
	preBuffSize = 32
)

type NodePool struct {
	valueNodes sync.Pool
	shortNodes sync.Pool
	fullNodes  sync.Pool
}

func NewNodePool() *NodePool {
	return &NodePool{}
}

func (np *NodePool) GetValueNode() *ValueNode {
	if node, ok := np.valueNodes.Get().(*ValueNode); ok && node != nil {
		// GC mark
		node._inpool = false

		return node
	}

	node := new(ValueNode)

	node.buf = make([]byte, 0, preBuffSize)
	node.hash = false

	// GC mark
	node._inpool = false
	runtime.SetFinalizer(node, np.PutValueNode)

	return node
}

func (np *NodePool) PutValueNode(node *ValueNode) {
	if node._inpool {
		// object already in pool, allow GC to collect
		runtime.SetFinalizer(node, nil)

		return
	}

	node._inpool = true

	node.buf = node.buf[0:0]
	node.hash = false

	np.valueNodes.Put(node)
}

func (np *NodePool) GetShortNode() *ShortNode {
	if node, ok := np.shortNodes.Get().(*ShortNode); ok && node != nil {
		// object unmark in pool
		node._inpool = false

		return node
	}

	node := new(ShortNode)
	node.hash = make([]byte, 0, preBuffSize)
	node.key = make([]byte, 0, preBuffSize)
	node.child = nil

	// GC mark
	node._inpool = false
	runtime.SetFinalizer(node, np.PutShortNode)

	return node
}

func (np *NodePool) PutShortNode(node *ShortNode) {
	if node._inpool {
		// object already in pool, allow GC to collect
		runtime.SetFinalizer(node, nil)

		return
	}

	// object mark in pool
	node._inpool = true

	node.key = node.key[0:0]
	node.hash = node.hash[0:0]
	node.child = nil

	np.shortNodes.Put(node)
}

func (np *NodePool) GetFullNode() *FullNode {
	if node, ok := np.fullNodes.Get().(*FullNode); ok && node != nil {
		node._inpool = false

		return node
	}

	node := new(FullNode)
	node.hash = make([]byte, 0, preBuffSize)
	node.epoch = 0
	node.value = nil

	for i := 0; i < 16; i++ {
		node.children[i] = nil
	}

	// GC mark
	node._inpool = false
	runtime.SetFinalizer(node, np.PutFullNode)

	return node
}

func (np *NodePool) PutFullNode(node *FullNode) {
	if node._inpool {
		// object already in pool, allow GC to collect
		runtime.SetFinalizer(node, nil)

		return
	}

	node._inpool = true

	node.hash = node.hash[0:0]
	node.epoch = 0
	node.value = nil

	for i := 0; i < 16; i++ {
		node.children[i] = nil
	}

	np.fullNodes.Put(node)
}
