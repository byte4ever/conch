package observer

import (
	"fmt"
	"testing"
	"time"
)

func TestNewMyObs(t *testing.T) {
	s := NewTopic1()
	c := make(chan interface{}, 1)

	oc := NewMyObs("OC")
	_ = s.RegisterChan(oc, c)
	o1 := NewMyObs("O1")
	o2 := NewMyObs("O2")
	o3 := NewMyObs("O3")
	_ = s.Register(o1)
	_ = s.Register(o2)
	_ = s.Register(o3)
	s.SetA(1)
	time.Sleep(1 * time.Millisecond)
	s.SetB(2)
	time.Sleep(1 * time.Millisecond)
	s.SetC(3)
	time.Sleep(1 * time.Millisecond)
	_ = s.Unregister(o1)
	s.SetA(123)
	time.Sleep(4 * time.Millisecond)
	oc.OnNotify(<-c)
	oc.OnNotify(<-c)
	oc.OnNotify(<-c)
	oc.OnNotify(<-c)
}

type MyObs struct {
	name string
}

func NewMyObs(name string) *MyObs {
	return &MyObs{name: name}
}

func (m MyObs) OnNotify(i interface{}) {
	switch s := i.(type) {
	case ValueChangedA:
		fmt.Println(m.name, " got A ", s)
	case ValueChangedB:
		fmt.Println(m.name, " got B ", s)
	case ValueChangedC:
		fmt.Println(m.name, " got C ", s)
	}
}

type Topic1 struct {
	Observable
	a, b, c int
}

func NewTopic1() *Topic1 {
	return &Topic1{
		Observable: NewObservable(),
	}
}

type ValueChangedA int
type ValueChangedB int
type ValueChangedC int

func (m *Topic1) SetA(v int) {
	m.a = v
	m.Notify(ValueChangedA(v))
}

func (m *Topic1) SetB(v int) {
	m.b = v
	m.Notify(ValueChangedB(v))
}

func (m *Topic1) SetC(v int) {
	m.c = v
	m.Notify(ValueChangedC(v))
}
