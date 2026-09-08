// Package nodeaudit_test sweeps every registered node type for defects that do
// not depend on any external resource: nil factories, and panics on the kinds of
// message an executor can actually be handed at runtime.
//
// A panic here is severe: nodes process messages on their own goroutines, so an
// unrecovered panic inside an executor terminates the whole EdgeFlow process.
package nodeaudit_test

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/EdgxCloud/EdgeFlow/internal/node"
	aiNodes "github.com/EdgxCloud/EdgeFlow/pkg/nodes/ai"
	coreNodes "github.com/EdgxCloud/EdgeFlow/pkg/nodes/core"
	dashboardNodes "github.com/EdgxCloud/EdgeFlow/pkg/nodes/dashboard"
	databaseNodes "github.com/EdgxCloud/EdgeFlow/pkg/nodes/database"
	gpioNodes "github.com/EdgxCloud/EdgeFlow/pkg/nodes/gpio"
	_ "github.com/EdgxCloud/EdgeFlow/pkg/nodes/industrial"
	messagingNodes "github.com/EdgxCloud/EdgeFlow/pkg/nodes/messaging"
	networkNodes "github.com/EdgxCloud/EdgeFlow/pkg/nodes/network"
	parserNodes "github.com/EdgxCloud/EdgeFlow/pkg/nodes/parser"
	storageNodes "github.com/EdgxCloud/EdgeFlow/pkg/nodes/storage"
	_ "github.com/EdgxCloud/EdgeFlow/pkg/nodes/wireless"
)

var once sync.Once

func allNodes(t *testing.T) []*node.NodeInfo {
	t.Helper()
	reg := node.GetGlobalRegistry()
	once.Do(func() {
		_ = coreNodes.RegisterAllNodes(reg)
		_ = dashboardNodes.RegisterAll(reg)
		_ = gpioNodes.RegisterAllNodes(reg)
		networkNodes.RegisterAllNodes(reg)
		databaseNodes.RegisterAllNodes(reg)
		storageNodes.RegisterAllNodes(reg)
		_ = messagingNodes.RegisterAllNodes(reg)
		_ = aiNodes.RegisterAllNodes(reg)
		_ = parserNodes.RegisterNodes(reg)
	})
	infos := reg.List()
	sort.Slice(infos, func(i, j int) bool { return infos[i].Type < infos[j].Type })
	return infos
}

// schemaDefaults builds a config from the defaults a node advertises.
func schemaDefaults(info *node.NodeInfo) map[string]interface{} {
	cfg := map[string]interface{}{}
	for _, p := range info.Properties {
		if p.Default != nil {
			cfg[p.Name] = p.Default
		}
	}
	return cfg
}

func guard(what string, fn func()) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic in %s: %v", what, r)
		}
	}()
	fn()
	return nil
}

func TestEveryNodeTypeHasAFactory(t *testing.T) {
	for _, info := range allNodes(t) {
		if info.Factory == nil {
			t.Errorf("%s: registered with a nil factory", info.Type)
			continue
		}
		if err := guard(info.Type+".Factory", func() {
			if info.Factory() == nil {
				t.Errorf("%s: factory returned a nil executor", info.Type)
			}
		}); err != nil {
			t.Errorf("%s: %v", info.Type, err)
		}
	}
}

// No node may panic on an ordinary message. Nodes that cannot initialise without
// a broker, device or credentials are skipped: the runtime aborts on an Init
// error and never calls Execute on them.
func TestNoNodePanicsOnOrdinaryInput(t *testing.T) {
	// Payloads mirror what Node.handleMessage can hand an executor. It
	// guarantees a non-nil map, so nil is deliberately not probed.
	payloads := []map[string]interface{}{
		{},
		{"value": 1.0, "payload": "test", "topic": "t"},
		{"value": "text", "payload": map[string]interface{}{"nested": true}},
	}

	for _, info := range allNodes(t) {
		if info.Factory == nil {
			continue
		}
		info := info
		t.Run(info.Type, func(t *testing.T) {
			ex := info.Factory()
			if ex == nil {
				t.Skip("nil executor")
			}

			var initErr error
			if err := guard("Init", func() { initErr = ex.Init(schemaDefaults(info)) }); err != nil {
				t.Fatalf("%v", err)
			}
			if initErr != nil {
				t.Skipf("needs external configuration: %v", initErr)
			}

			for i, p := range payloads {
				ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
				err := guard(fmt.Sprintf("Execute(payload#%d)", i), func() {
					// Execution errors are expected without real resources;
					// only a panic is a defect.
					_, _ = ex.Execute(ctx, node.Message{Type: node.MessageTypeData, Payload: p})
				})
				cancel()
				if err != nil {
					t.Fatalf("%v", err)
				}
			}

			if err := guard("Cleanup", func() { _ = ex.Cleanup() }); err != nil {
				t.Errorf("%v", err)
			}
		})
	}
}
