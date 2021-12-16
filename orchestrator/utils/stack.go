package utils

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"sync"
)

type Stack struct {
	data    []string       // container where data will reside
	indexes map[string]int // where our data are located
	lock    sync.Mutex     // mutex lock
}

func NewStack() *Stack {
	return &Stack{
		data:    make([]string, 0),
		indexes: make(map[string]int),
	}
}

// Push a new element into stack
func (stk *Stack) Push(value []byte) {
	stk.lock.Lock()
	defer stk.lock.Unlock()
	hex := common.Bytes2Hex(value)
	stk.indexes[hex] = len(stk.data)
	stk.data = append(stk.data, hex)
}

// Pop - pop out the last element from the stack
func (stk *Stack) Pop() ([]byte, error) {
	stk.lock.Lock()
	defer stk.lock.Unlock()
	l := len(stk.data)
	if l == 0 {
		return nil, errors.New("Pop(): Attempted pop from empty stack")
	}

	retVal := stk.data[l-1]
	stk.data = stk.data[:l-1]
	delete(stk.indexes, retVal)
	return common.Hex2Bytes(retVal), nil
}

// Top - the latest element in the cache
func (stk *Stack) Top() ([]byte, error) {
	stk.lock.Lock()
	defer stk.lock.Unlock()
	l := len(stk.data)
	if l == 0 {
		return nil, errors.New("Top(): Attempted read from empty stack")
	}

	retVal := stk.data[l-1]
	return common.Hex2Bytes(retVal), nil
}

// Contains - refers if key is inside stack or not
func (stk *Stack) Contains(hash []byte) bool {
	stk.lock.Lock()
	defer stk.lock.Unlock()
	_, found := stk.indexes[common.Bytes2Hex(hash)]
	return found
}

// Delete if the hash is found inside queue.
func (stk *Stack) Delete(hash []byte) {
	stk.lock.Lock()
	defer stk.lock.Unlock()
	key := common.Bytes2Hex(hash)
	index, found := stk.indexes[key]
	if found {
		delete(stk.indexes, key)
		stk.data = append(stk.data[:index], stk.data[index+1:]...)
	}
}

// Purge removes everything from the stack
func (stk *Stack) Purge() {
	stk.lock.Lock()
	defer stk.lock.Unlock()

	stk.data = make([]string, 0)
	stk.indexes = make(map[string]int)
}
