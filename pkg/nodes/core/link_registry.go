package core

import (
	"sync"
)

// LinkRegistry manages the mapping between Link Out and Link In nodes
type LinkRegistry struct {
	mu        sync.RWMutex
	linkNodes map[string][]*LinkInNode // key: linkID, value: array of Link In nodes
}

var (
	globalLinkRegistry *LinkRegistry
	registryOnce       sync.Once
)

// GetLinkRegistry returns the global link registry (singleton)
func GetLinkRegistry() *LinkRegistry {
	registryOnce.Do(func() {
		globalLinkRegistry = &LinkRegistry{
			linkNodes: make(map[string][]*LinkInNode),
		}
	})
	return globalLinkRegistry
}

// RegisterLinkIn registers a Link In node in the registry
func RegisterLinkIn(node *LinkInNode) error {
	registry := GetLinkRegistry()
	registry.mu.Lock()
	defer registry.mu.Unlock()

	linkID := node.GetLinkID()

	// Add to the list of nodes for this link ID
	registry.linkNodes[linkID] = append(registry.linkNodes[linkID], node)

	return nil
}

// UnregisterLinkIn removes a Link In node from the registry
func UnregisterLinkIn(node *LinkInNode) {
	registry := GetLinkRegistry()
	registry.mu.Lock()
	defer registry.mu.Unlock()

	linkID := node.GetLinkID()

	// Remove this node from the list
	nodes := registry.linkNodes[linkID]
	for i, n := range nodes {
		if n == node {
			registry.linkNodes[linkID] = append(nodes[:i], nodes[i+1:]...)
			break
		}
	}

	// Clean up empty entries
	if len(registry.linkNodes[linkID]) == 0 {
		delete(registry.linkNodes, linkID)
	}
}

// GetLinkInNodes returns all Link In nodes matching the given criteria
func GetLinkInNodes(linkIDs []string, scope string, flowID string) []*LinkInNode {
	registry := GetLinkRegistry()
	registry.mu.RLock()
	defer registry.mu.RUnlock()

	var result []*LinkInNode

	for _, linkID := range linkIDs {
		nodes, exists := registry.linkNodes[linkID]
		if !exists {
			continue
		}

		for _, node := range nodes {
			// Check scope and flow ID match
			if node.GetScope() != scope {
				continue
			}

			if scope == "flow" && node.GetFlowID() != flowID {
				continue
			}

			result = append(result, node)
		}
	}

	return result
}

// ClearLinkRegistry clears all registered link nodes (for testing)
func ClearLinkRegistry() {
	registry := GetLinkRegistry()
	registry.mu.Lock()
	defer registry.mu.Unlock()

	registry.linkNodes = make(map[string][]*LinkInNode)
}

// GetLinkStats returns statistics about registered links
func GetLinkStats() map[string]int {
	registry := GetLinkRegistry()
	registry.mu.RLock()
	defer registry.mu.RUnlock()

	stats := make(map[string]int)
	for linkID, nodes := range registry.linkNodes {
		stats[linkID] = len(nodes)
	}

	return stats
}
