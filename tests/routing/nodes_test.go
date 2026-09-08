package routing_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/EdgxCloud/EdgeFlow/internal/node"
	networkNodes "github.com/EdgxCloud/EdgeFlow/pkg/nodes/network"
)

// panicExecutor blows up on every message, standing in for any node with a
// latent bug.
type panicExecutor struct{}

func (panicExecutor) Init(map[string]interface{}) error { return nil }
func (panicExecutor) Cleanup() error                    { return nil }
func (panicExecutor) Execute(context.Context, node.Message) (node.Message, error) {
	panic("boom")
}

// A panicking executor must fail its own node, not terminate the process. Nodes
// run on their own goroutines, so an unrecovered panic would be fatal.
func TestPanicInExecutorIsContained(t *testing.T) {
	n := node.NewNode("panicky", "panicky", node.NodeTypeFunction, panicExecutor{})

	var (
		gotStatus string
		gotErr    string
		done      = make(chan struct{})
	)
	n.SetExecutionCallback(func(e node.ExecutionEvent) {
		gotStatus = e.Status
		gotErr = e.Error
		close(done)
	})

	if err := n.Start(context.Background()); err != nil {
		t.Fatalf("start: %v", err)
	}
	defer n.Stop()

	if err := n.Send(node.Message{Type: node.MessageTypeData,
		Payload: map[string]interface{}{"v": 1}}); err != nil {
		t.Fatalf("send: %v", err)
	}

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("no execution event; the panic was not converted into a node error")
	}

	if gotStatus != "error" {
		t.Errorf("status = %q, want %q", gotStatus, "error")
	}
	if gotErr == "" {
		t.Error("expected the panic to be reported as the node error")
	}
}

// recordingExecutor captures the payload it was handed.
type recordingExecutor struct{ got chan map[string]interface{} }

func (recordingExecutor) Init(map[string]interface{}) error { return nil }
func (recordingExecutor) Cleanup() error                    { return nil }
func (r recordingExecutor) Execute(_ context.Context, msg node.Message) (node.Message, error) {
	select {
	case r.got <- msg.Payload:
	default:
	}
	// Writing into the payload is what most executors do first; it panics if the
	// runtime handed over a nil map.
	msg.Payload["touched"] = true
	return msg, nil
}

// Executors index into the payload directly, so the runtime must never hand
// them a nil map.
func TestNilPayloadIsNormalisedBeforeExecute(t *testing.T) {
	rec := recordingExecutor{got: make(chan map[string]interface{}, 1)}
	n := node.NewNode("recorder", "recorder", node.NodeTypeFunction, rec)

	if err := n.Start(context.Background()); err != nil {
		t.Fatalf("start: %v", err)
	}
	defer n.Stop()

	// A message carrying no payload at all.
	if err := n.Send(node.Message{Type: node.MessageTypeData}); err != nil {
		t.Fatalf("send: %v", err)
	}

	select {
	case got := <-rec.got:
		if got == nil {
			t.Fatal("executor received a nil payload; it must be normalised to an empty map")
		}
	case <-time.After(3 * time.Second):
		t.Fatal("executor was never invoked")
	}
}

// The filter node owns two outputs: port 0 for matches, port 1 for non-matches.
func TestFilterNodeRoutesToMatchAndNoMatchPorts(t *testing.T) {
	for _, tc := range []struct {
		value float64
		want  string
	}{
		{10, "match"},   // 10 > 5
		{1, "no-match"}, // 1 not > 5
	} {
		t.Run(fmt.Sprintf("value_%v", tc.value), func(t *testing.T) {
			reg := registry(t)
			h := newHarness("filter")
			id := h.add(t, reg, "filter", "filter", map[string]interface{}{
				"property": "value", "operator": "gt", "value": 5.0})
			h.flow.ConnectPort(id, h.add(t, reg, "match", "debug", map[string]interface{}{"active": false}), 0)
			h.flow.ConnectPort(id, h.add(t, reg, "no-match", "debug", map[string]interface{}{"active": false}), 1)

			evs := h.send(t, id, map[string]interface{}{"value": tc.value})
			got := sinksHit(evs, "filter")
			if len(got) != 1 || got[0] != tc.want {
				t.Errorf("value=%v: want only %q to fire, got %v", tc.value, tc.want, got)
			}
		})
	}
}

// The editor renders QoS as a select, whose value is a string. Decoding must
// accept that as well as the numeric form used by imported flows.
func TestMQTTQoSAcceptsStringAndNumber(t *testing.T) {
	for _, tc := range []struct {
		raw  string
		want networkNodes.QoSLevel
	}{
		{`"0"`, 0},
		{`"1"`, 1},
		{`"2"`, 2},
		{`0`, 0},
		{`2`, 2},
		{`""`, 0},
	} {
		var q networkNodes.QoSLevel
		if err := json.Unmarshal([]byte(tc.raw), &q); err != nil {
			t.Errorf("qos %s: unexpected error: %v", tc.raw, err)
			continue
		}
		if q != tc.want {
			t.Errorf("qos %s = %d, want %d", tc.raw, q, tc.want)
		}
	}

	var q networkNodes.QoSLevel
	if err := json.Unmarshal([]byte(`"high"`), &q); err == nil {
		t.Error("expected a non-numeric qos string to be rejected")
	}
}

// A node must initialise from the defaults its own schema advertises, otherwise
// dropping it on the canvas and accepting the defaults yields a node that
// refuses to start.
func TestNodesInitialiseFromTheirSchemaDefaults(t *testing.T) {
	reg := registry(t)

	// Nodes legitimately requiring credentials, an address or user-supplied
	// rules cannot initialise bare and are out of scope here.
	needsUserConfig := map[string]bool{
		"switch": true, "link-in": true, "link-out": true,
	}

	for _, typ := range []string{
		"trigger", "delay", "set", "change", "math", "template", "hash",
		"base64", "filter", "range", "rbe", "regex", "statistics", "split",
		"join", "debug", "rate-limit", "if", "compress",
	} {
		if needsUserConfig[typ] {
			continue
		}
		info, err := reg.Get(typ)
		if err != nil {
			t.Errorf("%s: not registered: %v", typ, err)
			continue
		}
		cfg := map[string]interface{}{}
		for _, p := range info.Properties {
			if p.Default != nil {
				cfg[p.Name] = p.Default
			}
		}
		if err := info.Factory().Init(cfg); err != nil {
			t.Errorf("%s: Init with its own schema defaults failed: %v", typ, err)
		}
	}
}
