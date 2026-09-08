// Package routing_test covers the node runtime's message-passing guarantees:
// payload isolation between branches, accurate execution traces, and
// port-based conditional routing.
package routing_test

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/EdgxCloud/EdgeFlow/internal/engine"
	"github.com/EdgxCloud/EdgeFlow/internal/node"
	coreNodes "github.com/EdgxCloud/EdgeFlow/pkg/nodes/core"
)

var registerOnce sync.Once

func registry(t *testing.T) *node.Registry {
	t.Helper()
	reg := node.GetGlobalRegistry()
	registerOnce.Do(func() {
		if err := coreNodes.RegisterAllNodes(reg); err != nil {
			t.Fatalf("register core nodes: %v", err)
		}
	})
	return reg
}

type event struct {
	name string
	in   map[string]interface{}
	out  map[string]interface{}
}

type harness struct {
	flow *engine.Flow
	mu   sync.Mutex
	evs  []event
}

func newHarness(name string) *harness {
	h := &harness{flow: engine.NewFlow(name, "")}
	h.flow.SetExecutionCallback(func(e node.ExecutionEvent) {
		h.mu.Lock()
		defer h.mu.Unlock()
		h.evs = append(h.evs, event{e.NodeName, e.Input, e.Output})
	})
	return h
}

func (h *harness) add(t *testing.T, reg *node.Registry, name, typ string, cfg map[string]interface{}) string {
	t.Helper()
	n, err := reg.CreateNode(typ, name)
	if err != nil {
		t.Fatalf("create node %s: %v", typ, err)
	}
	n.Config = cfg
	if err := h.flow.AddNode(n); err != nil {
		t.Fatalf("add node %s: %v", name, err)
	}
	return n.ID
}

// send starts the flow, delivers payloads to the entry node, and returns the
// recorded execution events once the flow goes quiet.
func (h *harness) send(t *testing.T, entry string, payloads ...map[string]interface{}) []event {
	t.Helper()
	if err := h.flow.Start(context.Background()); err != nil {
		t.Fatalf("start flow: %v", err)
	}
	defer h.flow.Stop()

	n, err := h.flow.GetNode(entry)
	if err != nil {
		t.Fatalf("entry node: %v", err)
	}
	for _, p := range payloads {
		if err := n.Send(node.Message{Type: node.MessageTypeData, Payload: p}); err != nil {
			t.Fatalf("send: %v", err)
		}
	}
	time.Sleep(700 * time.Millisecond)

	h.mu.Lock()
	defer h.mu.Unlock()
	return append([]event(nil), h.evs...)
}

func sinksHit(evs []event, source string) []string {
	out := []string{}
	for _, e := range evs {
		if e.name != source {
			out = append(out, e.name)
		}
	}
	return out
}

func jsonOf(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprintf("<%v>", err)
	}
	return string(b)
}

func setRule(prop string, to interface{}, typ string) map[string]interface{} {
	return map[string]interface{}{"t": "set", "p": prop, "to": to, "tot": typ}
}

// A node's recorded input must be its pre-execution state. Executors mutate the
// payload in place and return it, so without an input snapshot the debug view
// reports the output for both fields.
func TestExecutionTraceInputIsPreExecutionState(t *testing.T) {
	reg := registry(t)
	h := newHarness("trace")

	setID := h.add(t, reg, "set", "set", map[string]interface{}{
		"rules": []interface{}{setRule("temperature", 21.5, "num")}})
	mathID := h.add(t, reg, "math", "math", map[string]interface{}{
		"operation": "multiply", "operand": 2.0, "property": "temperature"})
	if err := h.flow.Connect(setID, mathID); err != nil {
		t.Fatalf("connect: %v", err)
	}

	evs := h.send(t, setID, map[string]interface{}{"origin": "seed"})
	if len(evs) != 2 {
		t.Fatalf("expected 2 execution events, got %d", len(evs))
	}

	byName := map[string]event{}
	for _, e := range evs {
		byName[e.name] = e
	}

	if _, leaked := byName["set"].in["temperature"]; leaked {
		t.Errorf("set node input contains its own output: %s", jsonOf(byName["set"].in))
	}
	if got := byName["set"].out["temperature"]; got == nil {
		t.Errorf("set node output missing temperature: %s", jsonOf(byName["set"].out))
	}
	if in, out := jsonOf(byName["math"].in), jsonOf(byName["set"].out); in != out {
		t.Errorf("math input should equal set output\n  in : %s\n  out: %s", in, out)
	}
}

// Sibling branches must each own their payload. Sharing one map across branches
// corrupts data and crashes the process with a concurrent map write.
func TestFanOutBranchesAreIsolated(t *testing.T) {
	reg := registry(t)
	h := newHarness("fanout")

	src := h.add(t, reg, "src", "set", map[string]interface{}{
		"rules": []interface{}{setRule("reading", 10.0, "num")}})
	const branches = 6
	for i := 0; i < branches; i++ {
		name := fmt.Sprintf("b%d", i)
		id := h.add(t, reg, name, "change", map[string]interface{}{
			"rules": []interface{}{setRule("owner", name, "str")}})
		if err := h.flow.Connect(src, id); err != nil {
			t.Fatalf("connect: %v", err)
		}
	}

	payloads := make([]map[string]interface{}, 0, 50)
	for i := 0; i < 50; i++ {
		payloads = append(payloads, map[string]interface{}{"origin": i})
	}

	for _, e := range h.send(t, src, payloads...) {
		if e.name == "src" {
			continue
		}
		if owner, ok := e.out["owner"].(string); ok && owner != e.name {
			t.Errorf("branch %s saw another branch data: owner=%q", e.name, owner)
		}
	}
}

// The if node owns two output ports: 0 for true, 1 for false.
func TestIfNodeRoutesToSelectedPort(t *testing.T) {
	for _, tc := range []struct {
		name string
		temp float64
		want string
	}{
		{"condition true routes to port 0", 100, "hot"},
		{"condition false routes to port 1", 10, "cold"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			reg := registry(t)
			h := newHarness("if")
			ifID := h.add(t, reg, "if", "if", map[string]interface{}{
				"field": "t", "operator": "gt", "value": 50.0})
			hot := h.add(t, reg, "hot", "debug", map[string]interface{}{"active": false})
			cold := h.add(t, reg, "cold", "debug", map[string]interface{}{"active": false})
			h.flow.ConnectPort(ifID, hot, 0)
			h.flow.ConnectPort(ifID, cold, 1)

			evs := h.send(t, ifID, map[string]interface{}{"t": tc.temp})
			got := sinksHit(evs, "if")
			if len(got) != 1 || got[0] != tc.want {
				t.Errorf("t=%v: want only %q to fire, got %v", tc.temp, tc.want, got)
			}
		})
	}
}

// Each switch rule owns the output port of the same index.
func TestSwitchNodeRoutesToMatchingPort(t *testing.T) {
	for _, tc := range []struct {
		value float64
		want  string
	}{
		{5, "low"},
		{40, "mid"},
		{90, "high"},
	} {
		t.Run(fmt.Sprintf("value_%v", tc.value), func(t *testing.T) {
			reg := registry(t)
			h := newHarness("switch")
			swID := h.add(t, reg, "sw", "switch", map[string]interface{}{
				"property": "value",
				"checkall": false, // stop at first match
				"rules": []interface{}{
					map[string]interface{}{"t": "lt", "v": 20.0},
					map[string]interface{}{"t": "lt", "v": 60.0},
					map[string]interface{}{"t": "gte", "v": 60.0},
				}})
			h.flow.ConnectPort(swID, h.add(t, reg, "low", "debug", map[string]interface{}{"active": false}), 0)
			h.flow.ConnectPort(swID, h.add(t, reg, "mid", "debug", map[string]interface{}{"active": false}), 1)
			h.flow.ConnectPort(swID, h.add(t, reg, "high", "debug", map[string]interface{}{"active": false}), 2)

			evs := h.send(t, swID, map[string]interface{}{"value": tc.value})
			got := sinksHit(evs, "sw")
			if len(got) != 1 || got[0] != tc.want {
				t.Errorf("value=%v: want only %q to fire, got %v", tc.value, tc.want, got)
			}
		})
	}
}

// Sending to a node that was never started must report an error, not panic on a
// nil context.
func TestSendBeforeStartReturnsError(t *testing.T) {
	reg := registry(t)
	n, err := reg.CreateNode("debug", "orphan")
	if err != nil {
		t.Fatalf("create node: %v", err)
	}
	if err := n.Send(node.Message{Type: node.MessageTypeData}); err == nil {
		t.Error("expected an error when sending to a node that is not running")
	}
}

// The legacy function DSL must evaluate arithmetic against payload fields
// rather than treating the whole expression as a property name.
func TestFunctionNodeArithmeticOnPayloadField(t *testing.T) {
	reg := registry(t)
	info, err := reg.Get("function")
	if err != nil {
		t.Fatalf("get function node: %v", err)
	}
	ex := info.Factory()
	if err := ex.Init(map[string]interface{}{
		"code": "msg.payload.doubled = msg.payload.value * 2\nmsg.payload.plus = msg.payload.value + 10\nreturn msg",
	}); err != nil {
		t.Fatalf("init: %v", err)
	}

	out, err := ex.Execute(context.Background(), node.Message{
		Type:    node.MessageTypeData,
		Payload: map[string]interface{}{"value": 4.0},
	})
	if err != nil {
		t.Fatalf("execute: %v", err)
	}
	if got := out.Payload["doubled"]; got != 8.0 {
		t.Errorf("doubled = %v (%T), want 8", got, got)
	}
	if got := out.Payload["plus"]; got != 14.0 {
		t.Errorf("plus = %v (%T), want 14", got, got)
	}
}
