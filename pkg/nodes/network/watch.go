package network

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/edgeflow/edgeflow/internal/node"
	"github.com/fsnotify/fsnotify"
)

type WatchNode struct {
	path      string
	recursive bool
	events    []string
	pattern   string

	watcher *fsnotify.Watcher
	mu      sync.Mutex
	running bool
	msgChan chan node.Message
}

func NewWatchNode() *WatchNode {
	return &WatchNode{
		events:  []string{"create", "modify", "delete"},
		msgChan: make(chan node.Message, 100),
	}
}

func (n *WatchNode) Init(config map[string]interface{}) error {
	if path, ok := config["path"].(string); ok {
		n.path = path
	}
	if recursive, ok := config["recursive"].(bool); ok {
		n.recursive = recursive
	}
	if pattern, ok := config["pattern"].(string); ok {
		n.pattern = pattern
	}
	if events, ok := config["events"].([]interface{}); ok {
		n.events = make([]string, 0, len(events))
		for _, e := range events {
			if str, ok := e.(string); ok {
				n.events = append(n.events, str)
			}
		}
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("failed to create watcher: %w", err)
	}
	n.watcher = watcher

	if err := n.addPath(n.path); err != nil {
		return err
	}

	go n.watchLoop()

	return nil
}

func (n *WatchNode) addPath(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("failed to stat path: %w", err)
	}

	if !info.IsDir() {
		return n.watcher.Add(path)
	}

	if err := n.watcher.Add(path); err != nil {
		return err
	}

	if n.recursive {
		return filepath.Walk(path, func(walkPath string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() && walkPath != path {
				return n.watcher.Add(walkPath)
			}
			return nil
		})
	}

	return nil
}

func (n *WatchNode) watchLoop() {
	n.mu.Lock()
	n.running = true
	n.mu.Unlock()

	for {
		select {
		case event, ok := <-n.watcher.Events:
			if !ok {
				return
			}
			n.handleEvent(event)

		case err, ok := <-n.watcher.Errors:
			if !ok {
				return
			}
			n.msgChan <- node.Message{
				Payload: map[string]interface{}{
					"error": err.Error(),
				},
				Topic: "error",
			}
		}
	}
}

func (n *WatchNode) handleEvent(event fsnotify.Event) {
	if n.pattern != "" {
		matched, _ := filepath.Match(n.pattern, filepath.Base(event.Name))
		if !matched {
			return
		}
	}

	var eventType string
	switch {
	case event.Op&fsnotify.Create == fsnotify.Create:
		eventType = "create"
	case event.Op&fsnotify.Write == fsnotify.Write:
		eventType = "modify"
	case event.Op&fsnotify.Remove == fsnotify.Remove:
		eventType = "delete"
	case event.Op&fsnotify.Rename == fsnotify.Rename:
		eventType = "rename"
	case event.Op&fsnotify.Chmod == fsnotify.Chmod:
		eventType = "chmod"
	default:
		return
	}

	shouldSend := false
	for _, e := range n.events {
		if e == eventType {
			shouldSend = true
			break
		}
	}

	if !shouldSend {
		return
	}

	n.msgChan <- node.Message{
		Payload: map[string]interface{}{
			"file":  event.Name,
			"event": eventType,
			"time":  time.Now().Unix(),
		},
		Topic: "file-event",
	}
}

func (n *WatchNode) Execute(ctx context.Context, msg node.Message) (node.Message, error) {
	select {
	case outMsg := <-n.msgChan:
		return outMsg, nil
	case <-time.After(100 * time.Millisecond):
		return node.Message{}, nil
	}
}

func (n *WatchNode) Cleanup() error {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.running = false
	if n.watcher != nil {
		return n.watcher.Close()
	}
	return nil
}

func NewWatchExecutor() node.Executor {
	return NewWatchNode()
}
