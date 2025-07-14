package domain_test

import (
	"context"
	"fmt"

	"github.com/byte4ever/conch/domain"
)

// MyProcessable implements the Processable interface for demonstration.
type MyProcessable struct {
	param  string
	result string
	err    error
}

func (p *MyProcessable) GetParam() string { return p.param }
func (p *MyProcessable) SetValue(v string) { p.result = v; p.err = nil }
func (p *MyProcessable) GetValue() string { return p.result }
func (p *MyProcessable) SetError(e error) { p.err = e; p.result = "" }
func (p *MyProcessable) GetError() error { return p.err }

// Example demonstrates the usage of the Processable interface.
func Example() {
	ctx := context.Background()
	elem := &MyProcessable{param: "test"}
	
	// Define a processor function
	processor := func(ctx context.Context, elem *MyProcessable) {
		// Process the element
		result := "processed: " + elem.GetParam()
		elem.SetValue(result)
	}
	
	// Apply the processor
	processor(ctx, elem)
	
	// Output the result
	fmt.Println(elem.GetValue())
	
	// Output:
	// processed: test
}
