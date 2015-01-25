// Package core provides Blocks, Sources and the means to connect them together. Blocks communicate with
// one another by passing Messages.
package core

import (
	"encoding/json"
	"sync"

	"github.com/nikhan/go-fetch"
)

const (
	NONE = iota
	KEY_VALUE
	STREAM
	PRIORITY
	ARRAY
	VALUE
)

// MessageMap maps a block's inbound routes onto the Messages they contain
type MessageMap map[RouteIndex]Message

// Message is the container for data sent between blocks
type Message interface{}

// RouteIndex is the index into a MessageMap. The 0th index corresponds to that block's 0th Input or Output
type RouteIndex int

// SourceType is used to indicate what kind of source a block can connect to
type SourceType int

// Connections are used to connect blocks together
type Connection chan Message

// Interrupt is a function that interrupts a running block in order to change its state.
// If the interrupt returns false, the block will quit.
type Interrupt func() bool

// Kernel is a block's core function that operates on an inbound message. It works by populating
// the outbound MessageMap, and can be interrupted on its Interrupt channel.
type Kernel func(MessageMap, MessageMap, MessageMap, Source, chan Interrupt) Interrupt

// A Pin contains information about a particular input or output
type Pin struct {
	Name string
}

// A Spec defines a block's input and output Pins, and the block's Kernel.
type Spec struct {
	Name    string
	Inputs  []Pin
	Outputs []Pin
	Source  SourceType
	Kernel  Kernel
}

// Input is an inbound route to a block. A Input holds the channel that allows Messages
// to be passed into the block. A Input's Path is applied to the inbound Message before populating the
// MessageMap and calling the Kernel. A Input can be set to a Value, instead of waiting for an inbound message.
type Input struct {
	Name  string
	Value interface{}
	C     chan Message
}

func (r Input) MarshalJSON() ([]byte, error) {
	out := make(map[string]interface{})
	value := make(map[string]interface{})
	out["name"] = r.Name

	switch q := r.Value.(type) {
	case *fetch.Query:
		value["fetch"] = q.String()
	default:
		value["json"] = q
	}

	out["value"] = value

	return json.Marshal(out)
}

// An Output holds a set of Connections. Each Connection refers to a Input.C. Every outbound
// mesage is sent on every Connection in the Connections set.
type Output struct {
	Name        string                  `json:"name"`
	Connections map[Connection]struct{} `json:"-"`
}

// A ManifestPair is a unique reference to an Output/Connection pair
type ManifestPair struct {
	int
	Connection
}

// A block's Manifest is the set of Connections
type Manifest map[ManifestPair]struct{}

// A block's BlockState is the pair of input/output MessageMaps, and the Manifest
type BlockState struct {
	inputValues    MessageMap
	outputValues   MessageMap
	internalValues MessageMap
	manifest       Manifest
	Processed      bool
}

// a Source is esssentially a lockable piece of memory that can be accessed safely by mulitple blocks.
// The Lock and Unlock methods are usually implemented using a sync.Mutex
// TODO Source -> Source
type Source interface {
	Lock()
	Unlock()
	Describe() map[string]string
	Serve()
	Stop()
	SetSourceParameter(key, value string)
	GetType() SourceType
}

// A block's BlockRouting is the set of Input and Output routes, and the Interrupt channel
type BlockRouting struct {
	Inputs        []Input
	Outputs       []Output
	Source        Source
	InterruptChan chan Interrupt
	sync.RWMutex
}

// A Block is comprised of a state, a set of routes, and a kernel
type Block struct {
	state      BlockState
	routing    BlockRouting
	kernel     Kernel
	sourceType SourceType
}