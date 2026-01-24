# Contributing to EdgeFlow

ممنون که می‌خواید به EdgeFlow کمک کنید! 🎉

## 📋 فهرست

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Making Changes](#making-changes)
- [Pull Request Process](#pull-request-process)
- [Coding Standards](#coding-standards)
- [Creating Nodes](#creating-nodes)

---

## Code of Conduct

- با احترام رفتار کنید
- انتقادات سازنده باشد
- روی مسئله تمرکز کنید، نه فرد
- از تنوع استقبال کنید

---

## Getting Started

### Issues

- قبل از شروع کار، یک issue باز کنید
- برای bug ها، جزئیات کامل بدید
- برای feature ها، use case توضیح بدید

### Good First Issues

به دنبال issues با برچسب `good first issue` باشید:
- مستندات
- تست‌ها
- نودهای ساده

---

## Development Setup

### پیش‌نیازها

```bash
# Go 1.21+
go version

# Node.js 18+ (برای frontend)
node --version

# Make
make --version
```

### Clone و Setup

```bash
# Fork the repo on GitHub
# Then clone your fork
git clone https://github.com/YOUR_USERNAME/edgeflow.git
cd edgeflow

# Add upstream remote
git remote add upstream https://github.com/edgeflow/edgeflow.git

# Install dependencies
go mod download
cd web && npm install && cd ..

# Run in development mode
make dev
```

---

## Making Changes

### Branch Naming

```
feature/short-description
bugfix/issue-number-description
docs/what-changed
refactor/what-changed
```

### Commit Messages

از [Conventional Commits](https://www.conventionalcommits.org/) استفاده کنید:

```
feat: add telegram node
fix: resolve memory leak in executor
docs: update API documentation
test: add tests for http node
refactor: simplify node registry
chore: update dependencies
```

### Workflow

```bash
# Create branch
git checkout -b feature/my-feature

# Make changes
# ...

# Run tests
make test

# Run linter
make lint

# Commit
git add .
git commit -m "feat: add my feature"

# Push
git push origin feature/my-feature

# Create Pull Request on GitHub
```

---

## Pull Request Process

### قبل از PR

- [ ] تست‌ها pass میشن (`make test`)
- [ ] Linter error نداره (`make lint`)
- [ ] مستندات به‌روز شده
- [ ] Commit messages درست هستن

### Template

```markdown
## Description
[توضیح تغییرات]

## Type of Change
- [ ] Bug fix
- [ ] New feature
- [ ] Breaking change
- [ ] Documentation update

## Testing
[چطور تست شده]

## Screenshots (if applicable)
[اسکرین‌شات‌ها]

## Checklist
- [ ] Tests pass
- [ ] Linter passes
- [ ] Documentation updated
```

### Review Process

1. حداقل یک approve لازمه
2. CI باید pass بشه
3. Conflicts حل بشه
4. Squash and merge

---

## Coding Standards

### Go

```go
// Package comment
package nodes

import (
    // Standard library
    "context"
    "fmt"
    
    // Third-party
    "github.com/gofiber/fiber/v2"
    
    // Internal
    "github.com/edgeflow/edgeflow/internal/node"
)

// Node interface implementation
// HTTPRequestNode makes HTTP requests
type HTTPRequestNode struct {
    BaseNode
    URL    string `json:"url"`
    Method string `json:"method"`
}

// Execute runs the node
func (n *HTTPRequestNode) Execute(ctx *node.ExecutionContext) (*node.Message, error) {
    // Implementation
}
```

### Rules

- از `gofmt` و `goimports` استفاده کنید
- نام‌ها واضح باشند
- کامنت برای public functions
- Error handling صحیح
- Context propagation

### Frontend (TypeScript)

```typescript
// Use interfaces
interface NodeProps {
  id: string;
  type: string;
  data: NodeData;
}

// Functional components
const CustomNode: React.FC<NodeProps> = ({ id, type, data }) => {
  // Implementation
};

// Named exports
export { CustomNode };
```

---

## Creating Nodes

### ساختار یک نود

```go
package network

import (
    "github.com/edgeflow/edgeflow/internal/node"
)

// Register the node
func init() {
    node.Register("my-node", NewMyNode)
}

// MyNode does something useful
type MyNode struct {
    node.BaseNode
    
    // Properties (shown in UI)
    Property1 string `json:"property1" title:"Property 1" description:"Description"`
    Property2 int    `json:"property2" title:"Property 2" default:"10"`
}

// NewMyNode creates a new instance
func NewMyNode() node.Node {
    return &MyNode{}
}

// GetSchema returns the node schema for UI
func (n *MyNode) GetSchema() node.NodeSchema {
    return node.NodeSchema{
        Type:        "my-node",
        Name:        "My Node",
        Description: "Does something useful",
        Category:    "Network",
        Inputs: []node.PortSchema{
            {Name: "input", Type: "any"},
        },
        Outputs: []node.PortSchema{
            {Name: "output", Type: "any"},
        },
        Properties: []node.PropertySchema{
            {Name: "property1", Type: "string", Required: true},
            {Name: "property2", Type: "number", Default: 10},
        },
    }
}

// Validate checks if the node configuration is valid
func (n *MyNode) Validate() error {
    if n.Property1 == "" {
        return fmt.Errorf("property1 is required")
    }
    return nil
}

// Execute runs the node logic
func (n *MyNode) Execute(ctx *node.ExecutionContext) (*node.Message, error) {
    input := ctx.Input
    
    // Do something with input
    result := process(input.Payload, n.Property1)
    
    return &node.Message{
        Payload: result,
        Meta: map[string]interface{}{
            "processedBy": "my-node",
        },
    }, nil
}
```

### تست نود

```go
package network_test

import (
    "testing"
    
    "github.com/stretchr/testify/assert"
    "github.com/edgeflow/edgeflow/pkg/nodes/network"
)

func TestMyNode_Execute(t *testing.T) {
    node := &network.MyNode{
        Property1: "test",
        Property2: 42,
    }
    
    ctx := &node.ExecutionContext{
        Input: &node.Message{Payload: "hello"},
    }
    
    result, err := node.Execute(ctx)
    
    assert.NoError(t, err)
    assert.NotNil(t, result)
    assert.Equal(t, "expected", result.Payload)
}
```

---

## Questions?

- 💬 [Discord](https://discord.gg/edgeflow)
- 📧 [Email](mailto:contribute@edgeflow.io)
- 🐛 [GitHub Issues](https://github.com/edgeflow/edgeflow/issues)

---

ممنون از کمک شما! 🙏
