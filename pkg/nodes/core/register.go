package core

import (
	"github.com/edgeflow/edgeflow/internal/node"
)

// RegisterAllNodes registers all core nodes with the registry
func RegisterAllNodes(registry *node.Registry) error {
	// Register Inject Node
	if err := registry.Register(&node.NodeInfo{
		Type:        "inject",
		Name:        "Inject",
		Category:    node.NodeTypeInput,
		Description: "Manually trigger flows or inject timestamps at intervals",
		Icon:        "clock",
		Color:       "#3b82f6",
		Properties: []node.PropertySchema{
			{
				Name:        "intervalType",
				Label:       "Trigger Interval",
				Type:        "select",
				Default:     "minutes",
				Options:     []string{"seconds", "minutes", "hours", "days", "months"},
				Required:    true,
				Description: "Select time unit for triggering",
			},
			{
				Name:        "intervalValue",
				Label:       "Value",
				Type:        "number",
				Default:     1,
				Required:    true,
				Description: "Enter value based on selected interval type",
			},
			{
				Name:        "repeat",
				Label:       "Repeat",
				Type:        "boolean",
				Default:     true,
				Required:    false,
				Description: "Repeat at intervals (or fire once)",
			},
			{
				Name:        "payload",
				Label:       "Payload",
				Type:        "object",
				Default:     map[string]interface{}{},
				Required:    false,
				Description: "Message payload to send",
			},
			{
				Name:        "topic",
				Label:       "Topic",
				Type:        "string",
				Default:     "",
				Required:    false,
				Description: "Message topic",
			},
		},
		Outputs: []node.PortSchema{
			{
				Name:        "output",
				Label:       "Output",
				Type:        "any",
				Description: "Message output",
			},
		},
		Factory: func() node.Executor {
			return NewInjectNode()
		},
	}); err != nil {
		return err
	}

	// Register Debug Node
	if err := registry.Register(&node.NodeInfo{
		Type:        "debug",
		Name:        "Debug",
		Category:    node.NodeTypeOutput,
		Description: "Display messages in debug sidebar for troubleshooting",
		Icon:        "bug",
		Color:       "#10b981",
		Properties: []node.PropertySchema{
			{
				Name:        "output_to",
				Label:       "Output To",
				Type:        "select",
				Default:     "console",
				Required:    true,
				Options:     []string{"console", "log"},
				Description: "Output destination",
			},
			{
				Name:        "complete",
				Label:       "Complete Message",
				Type:        "boolean",
				Default:     false,
				Required:    false,
				Description: "Display full message or only payload",
			},
		},
		Inputs: []node.PortSchema{
			{
				Name:        "input",
				Label:       "Input",
				Type:        "any",
				Description: "Message input",
			},
		},
		Factory: func() node.Executor {
			return NewDebugNode()
		},
	}); err != nil {
		return err
	}

	// Register If Node
	if err := registry.Register(&node.NodeInfo{
		Type:        "if",
		Name:        "If",
		Category:    node.NodeTypeProcessing,
		Description: "Route messages based on conditions",
		Icon:        "git-branch",
		Color:       "#f59e0b",
		Properties: []node.PropertySchema{
			{
				Name:        "field",
				Label:       "Field",
				Type:        "string",
				Default:     "",
				Required:    true,
				Description: "Field to check",
			},
			{
				Name:        "operator",
				Label:       "Operator",
				Type:        "select",
				Default:     "eq",
				Required:    true,
				Options:     []string{"eq", "ne", "gt", "lt", "gte", "lte", "contains", "exists"},
				Description: "Comparison operator",
			},
			{
				Name:        "value",
				Label:       "Value",
				Type:        "string",
				Default:     "",
				Required:    false,
				Description: "Value to compare",
			},
		},
		Inputs: []node.PortSchema{
			{
				Name:        "input",
				Label:       "Input",
				Type:        "any",
				Description: "Message input",
			},
		},
		Outputs: []node.PortSchema{
			{
				Name:        "true",
				Label:       "True",
				Type:        "any",
				Description: "Output when condition is true",
			},
			{
				Name:        "false",
				Label:       "False",
				Type:        "any",
				Description: "Output when condition is false",
			},
		},
		Factory: func() node.Executor {
			return NewIfNode()
		},
	}); err != nil {
		return err
	}

	// Register Delay Node
	if err := registry.Register(&node.NodeInfo{
		Type:        "delay",
		Name:        "Delay",
		Category:    node.NodeTypeFunction,
		Description: "Delay message processing",
		Icon:        "timer",
		Color:       "#8b5cf6",
		Properties: []node.PropertySchema{
			{
				Name:        "duration",
				Label:       "Duration",
				Type:        "string",
				Default:     "1s",
				Required:    true,
				Description: "Delay duration (e.g. 1s, 500ms, 2m)",
			},
			{
				Name:        "timeout",
				Label:       "Timeout",
				Type:        "string",
				Default:     "1m",
				Required:    false,
				Description: "Maximum wait time",
			},
		},
		Inputs: []node.PortSchema{
			{
				Name:        "input",
				Label:       "Input",
				Type:        "any",
				Description: "Message input",
			},
		},
		Outputs: []node.PortSchema{
			{
				Name:        "output",
				Label:       "Output",
				Type:        "any",
				Description: "Message output (after delay)",
			},
		},
		Factory: func() node.Executor {
			return NewDelayNode()
		},
	}); err != nil {
		return err
	}

	// Register Schedule Node
	if err := registry.Register(&node.NodeInfo{
		Type:        "schedule",
		Name:        "Schedule",
		Category:    node.NodeTypeInput,
		Description: "Trigger flows based on cron expressions",
		Icon:        "clock",
		Color:       "#10b981",
		Properties: []node.PropertySchema{
			{
				Name:        "cron",
				Label:       "Cron Expression",
				Type:        "string",
				Default:     "0 * * * * *",
				Required:    true,
				Description: "Cron expression (with seconds: second minute hour day month weekday)",
			},
			{
				Name:        "payload",
				Label:       "Payload",
				Type:        "object",
				Default:     map[string]interface{}{},
				Required:    false,
				Description: "Message payload to send",
			},
			{
				Name:        "topic",
				Label:       "Topic",
				Type:        "string",
				Default:     "",
				Required:    false,
				Description: "Message topic",
			},
			{
				Name:        "timezone",
				Label:       "Timezone",
				Type:        "string",
				Default:     "Local",
				Required:    false,
				Description: "Timezone for cron (e.g. UTC, America/New_York, Local)",
			},
		},
		Outputs: []node.PortSchema{
			{
				Name:        "output",
				Label:       "Output",
				Type:        "any",
				Description: "Scheduled message output",
			},
		},
		Factory: func() node.Executor {
			return NewScheduleNode()
		},
	}); err != nil {
		return err
	}

	// Register Function Node
	if err := registry.Register(&node.NodeInfo{
		Type:        "function",
		Name:        "Function",
		Category:    node.NodeTypeFunction,
		Description: "Execute custom JavaScript code",
		Icon:        "code",
		Color:       "#ec4899",
		Properties: []node.PropertySchema{
			{
				Name:        "code",
				Label:       "Code",
				Type:        "code",
				Default:     "// return msg;\n",
				Required:    true,
				Description: "JavaScript code to execute",
			},
			{
				Name:        "output_key",
				Label:       "Output Key",
				Type:        "string",
				Default:     "result",
				Required:    false,
				Description: "Key to store result in payload",
			},
		},
		Inputs: []node.PortSchema{
			{
				Name:        "input",
				Label:       "Input",
				Type:        "any",
				Description: "Message input",
			},
		},
		Outputs: []node.PortSchema{
			{
				Name:        "output",
				Label:       "Output",
				Type:        "any",
				Description: "Message output",
			},
		},
		Factory: func() node.Executor {
			return NewFunctionNode()
		},
	}); err != nil {
		return err
	}

	// Register Exec Node
	if err := registry.Register(&node.NodeInfo{
		Type:        "exec",
		Name:        "Exec",
		Category:    node.NodeTypeFunction,
		Description: "Execute shell commands",
		Icon:        "terminal",
		Color:       "#14b8a6",
		Properties: []node.PropertySchema{
			{
				Name:        "command",
				Label:       "Command",
				Type:        "string",
				Default:     "",
				Required:    true,
				Description: "Shell command to execute",
			},
			{
				Name:        "appendPayload",
				Label:       "Append Payload",
				Type:        "boolean",
				Default:     false,
				Required:    false,
				Description: "Append payload to command",
			},
			{
				Name:        "useSpawn",
				Label:       "Use Spawn",
				Type:        "boolean",
				Default:     false,
				Required:    false,
				Description: "Keep process alive",
			},
			{
				Name:        "timeout",
				Label:       "Timeout (seconds)",
				Type:        "number",
				Default:     10,
				Required:    false,
				Description: "Maximum execution time (seconds)",
			},
		},
		Inputs: []node.PortSchema{
			{
				Name:        "input",
				Label:       "Input",
				Type:        "any",
				Description: "Message input",
			},
		},
		Outputs: []node.PortSchema{
			{
				Name:        "stdout",
				Label:       "Stdout",
				Type:        "any",
				Description: "Standard output",
			},
			{
				Name:        "stderr",
				Label:       "Stderr",
				Type:        "any",
				Description: "Error output",
			},
			{
				Name:        "return",
				Label:       "Return Code",
				Type:        "any",
				Description: "Return code",
			},
		},
		Factory: func() node.Executor {
			return NewExecNode()
		},
	}); err != nil {
		return err
	}

	// Register Python Node
	if err := registry.Register(&node.NodeInfo{
		Type:        "python",
		Name:        "Python",
		Category:    node.NodeTypeFunction,
		Description: "Execute Python code",
		Icon:        "code",
		Color:       "#3776ab",
		Properties: []node.PropertySchema{
			{
				Name:        "code",
				Label:       "Python Code",
				Type:        "code",
				Default:     "# Write your Python code here\nprint('Hello from Python')\n",
				Required:    true,
				Description: "Python code to execute",
			},
			{
				Name:        "pythonPath",
				Label:       "Python Path",
				Type:        "string",
				Default:     "python3",
				Required:    false,
				Description: "Python executable path",
			},
			{
				Name:        "timeout",
				Label:       "Timeout (seconds)",
				Type:        "number",
				Default:     30,
				Required:    false,
				Description: "Maximum execution time (seconds)",
			},
			{
				Name:        "useVirtualEnv",
				Label:       "Use Virtual Environment",
				Type:        "boolean",
				Default:     false,
				Required:    false,
				Description: "Use virtual environment",
			},
			{
				Name:        "venvPath",
				Label:       "Virtual Env Path",
				Type:        "string",
				Default:     "",
				Required:    false,
				Description: "Virtual environment path",
			},
		},
		Inputs: []node.PortSchema{
			{
				Name:        "input",
				Label:       "Input",
				Type:        "any",
				Description: "Message input (available as {{msg.payload}})",
			},
		},
		Outputs: []node.PortSchema{
			{
				Name:        "output",
				Label:       "Output",
				Type:        "any",
				Description: "Message output",
			},
		},
		Factory: func() node.Executor {
			return NewPythonNode()
		},
	}); err != nil {
		return err
	}

	// ========== PHASE 8: WEEK 1-3 NODES ==========

	// Register Template Node (Week 1)
	if err := registry.Register(&node.NodeInfo{
		Type:        "template",
		Name:        "Template",
		Category:    node.NodeTypeFunction,
		Description: "Render Mustache templates",
		Icon:        "file-text",
		Color:       "#06b6d4",
		Properties: []node.PropertySchema{
			{
				Name:        "template",
				Label:       "Template",
				Type:        "code",
				Default:     "Hello {{msg.payload}}!",
				Required:    true,
				Description: "Mustache template",
			},
			{
				Name:        "field",
				Label:       "Set Property",
				Type:        "string",
				Default:     "payload",
				Required:    false,
				Description: "Field to set result",
			},
			{
				Name:        "syntax",
				Label:       "Syntax",
				Type:        "select",
				Default:     "mustache",
				Options:     []string{"mustache", "plain"},
				Description: "Template syntax",
			},
		},
		Inputs: []node.PortSchema{
			{Name: "input", Label: "Input", Type: "any"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Output", Type: "any"},
		},
		Factory: func() node.Executor {
			return NewTemplateNode()
		},
	}); err != nil {
		return err
	}

	// Register Change Node (Week 1)
	if err := registry.Register(&node.NodeInfo{
		Type:        "change",
		Name:        "Change",
		Category:    node.NodeTypeFunction,
		Description: "Set, change, move or delete message properties",
		Icon:        "edit",
		Color:       "#ef4444",
		Properties: []node.PropertySchema{
			{
				Name:        "rules",
				Label:       "Rules",
				Type:        "array",
				Default:     []interface{}{},
				Description: "List of change rules",
			},
		},
		Inputs: []node.PortSchema{
			{Name: "input", Label: "Input", Type: "any"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Output", Type: "any"},
		},
		Factory: func() node.Executor {
			return NewChangeNode()
		},
	}); err != nil {
		return err
	}

	// Register Range Node (Week 1)
	if err := registry.Register(&node.NodeInfo{
		Type:        "range",
		Name:        "Range",
		Category:    node.NodeTypeFunction,
		Description: "Scale numeric values between ranges",
		Icon:        "sliders",
		Color:       "#8b5cf6",
		Properties: []node.PropertySchema{
			{
				Name:        "action",
				Label:       "Action",
				Type:        "select",
				Default:     "scale",
				Options:     []string{"scale", "clamp", "wrap"},
				Description: "Action type",
			},
			{
				Name:        "minIn",
				Label:       "Min Input",
				Type:        "number",
				Default:     0.0,
				Description: "Minimum input",
			},
			{
				Name:        "maxIn",
				Label:       "Max Input",
				Type:        "number",
				Default:     100.0,
				Description: "Maximum input",
			},
			{
				Name:        "minOut",
				Label:       "Min Output",
				Type:        "number",
				Default:     0.0,
				Description: "Minimum output",
			},
			{
				Name:        "maxOut",
				Label:       "Max Output",
				Type:        "number",
				Default:     1.0,
				Description: "Maximum output",
			},
		},
		Inputs: []node.PortSchema{
			{Name: "input", Label: "Input", Type: "number"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Output", Type: "number"},
		},
		Factory: func() node.Executor {
			return NewRangeNode()
		},
	}); err != nil {
		return err
	}

	// Register Switch Node (Week 2)
	if err := registry.Register(&node.NodeInfo{
		Type:        "switch",
		Name:        "Switch",
		Category:    node.NodeTypeFunction,
		Description: "Route messages based on rules",
		Icon:        "git-branch",
		Color:       "#f59e0b",
		Properties: []node.PropertySchema{
			{
				Name:        "property",
				Label:       "Property",
				Type:        "string",
				Default:     "payload",
				Description: "Property to check",
			},
			{
				Name:        "rules",
				Label:       "Rules",
				Type:        "array",
				Default:     []interface{}{},
				Description: "List of routing rules",
			},
			{
				Name:        "checkall",
				Label:       "Check All Rules",
				Type:        "boolean",
				Default:     true,
				Description: "Check all rules or stop at first match",
			},
		},
		Inputs: []node.PortSchema{
			{Name: "input", Label: "Input", Type: "any"},
		},
		Outputs: []node.PortSchema{},
		Factory: func() node.Executor {
			return NewSwitchNode()
		},
	}); err != nil {
		return err
	}

	// Register Split Node (Week 2)
	if err := registry.Register(&node.NodeInfo{
		Type:        "split",
		Name:        "Split",
		Category:    node.NodeTypeFunction,
		Description: "Split array, object or string into separate messages",
		Icon:        "scissors",
		Color:       "#14b8a6",
		Properties: []node.PropertySchema{
			{
				Name:        "arraySplt",
				Label:       "Split Array",
				Type:        "boolean",
				Default:     true,
				Description: "Split array",
			},
			{
				Name:        "arraySplitType",
				Label:       "Array Split Type",
				Type:        "select",
				Default:     "each",
				Options:     []string{"each", "len"},
				Description: "Array split type",
			},
			{
				Name:        "arraySpltLen",
				Label:       "Array Split Length",
				Type:        "number",
				Default:     1,
				Description: "Length of each segment",
			},
		},
		Inputs: []node.PortSchema{
			{Name: "input", Label: "Input", Type: "any"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Output", Type: "any"},
		},
		Factory: func() node.Executor {
			return NewSplitNode()
		},
	}); err != nil {
		return err
	}

	// Register Join Node (Week 2)
	if err := registry.Register(&node.NodeInfo{
		Type:        "join",
		Name:        "Join",
		Category:    node.NodeTypeFunction,
		Description: "Join sequence of messages into array, object or string",
		Icon:        "link",
		Color:       "#06b6d4",
		Properties: []node.PropertySchema{
			{
				Name:        "mode",
				Label:       "Mode",
				Type:        "select",
				Default:     "auto",
				Options:     []string{"auto", "manual", "reduce", "merge"},
				Description: "Join mode",
			},
			{
				Name:        "build",
				Label:       "Build",
				Type:        "select",
				Default:     "array",
				Options:     []string{"array", "object", "string", "buffer"},
				Description: "Output type",
			},
			{
				Name:        "count",
				Label:       "Message Count",
				Type:        "number",
				Default:     0,
				Description: "Message count for manual mode",
			},
			{
				Name:        "joiner",
				Label:       "Join Character",
				Type:        "string",
				Default:     "\\n",
				Description: "Join character for string",
			},
		},
		Inputs: []node.PortSchema{
			{Name: "input", Label: "Input", Type: "any"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Output", Type: "any"},
		},
		Factory: func() node.Executor {
			return NewJoinNode()
		},
	}); err != nil {
		return err
	}

	// Register Catch Node (Week 3)
	if err := registry.Register(&node.NodeInfo{
		Type:        "catch",
		Name:        "Catch",
		Category:    node.NodeTypeFunction,
		Description: "Catch errors from other nodes",
		Icon:        "alert-triangle",
		Color:       "#dc2626",
		Properties: []node.PropertySchema{
			{
				Name:        "scope",
				Label:       "Scope",
				Type:        "select",
				Default:     "all",
				Options:     []string{"all", "flow", "nodes"},
				Description: "Monitoring scope",
			},
			{
				Name:        "nodeIds",
				Label:       "Node IDs",
				Type:        "array",
				Default:     []interface{}{},
				Description: "Target node IDs (for scope=nodes)",
			},
			{
				Name:        "uncaught",
				Label:       "Uncaught Only",
				Type:        "boolean",
				Default:     false,
				Description: "Only uncaught errors",
			},
		},
		Inputs: []node.PortSchema{
			{Name: "input", Label: "Input", Type: "any"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Output", Type: "any"},
		},
		Factory: func() node.Executor {
			return NewCatchNode()
		},
	}); err != nil {
		return err
	}

	// Register Status Node (Week 3)
	if err := registry.Register(&node.NodeInfo{
		Type:        "status",
		Name:        "Status",
		Category:    node.NodeTypeFunction,
		Description: "Monitor node status changes",
		Icon:        "activity",
		Color:       "#22c55e",
		Properties: []node.PropertySchema{
			{
				Name:        "scope",
				Label:       "Scope",
				Type:        "select",
				Default:     "all",
				Options:     []string{"all", "flow", "nodes"},
				Description: "Monitoring scope",
			},
			{
				Name:        "nodeIds",
				Label:       "Node IDs",
				Type:        "array",
				Default:     []interface{}{},
				Description: "Target node IDs (for scope=nodes)",
			},
		},
		Inputs: []node.PortSchema{
			{Name: "input", Label: "Input", Type: "any"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Output", Type: "any"},
		},
		Factory: func() node.Executor {
			return NewStatusNode()
		},
	}); err != nil {
		return err
	}

	// Register Complete Node (Week 3)
	if err := registry.Register(&node.NodeInfo{
		Type:        "complete",
		Name:        "Complete",
		Category:    node.NodeTypeFunction,
		Description: "Trigger when node or flow completes",
		Icon:        "check-circle",
		Color:       "#3b82f6",
		Properties: []node.PropertySchema{
			{
				Name:        "scope",
				Label:       "Scope",
				Type:        "select",
				Default:     "flow",
				Options:     []string{"all", "flow", "nodes"},
				Description: "Monitoring scope",
			},
			{
				Name:        "nodeIds",
				Label:       "Node IDs",
				Type:        "array",
				Default:     []interface{}{},
				Description: "Target node IDs (for scope=nodes)",
			},
		},
		Inputs: []node.PortSchema{
			{Name: "input", Label: "Input", Type: "any"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Output", Type: "any"},
		},
		Factory: func() node.Executor {
			return NewCompleteNode()
		},
	}); err != nil {
		return err
	}

	// Register Set Node (Property Manipulation)
	if err := registry.Register(&node.NodeInfo{
		Type:        "set",
		Name:        "Set",
		Category:    node.NodeTypeFunction,
		Description: "Set, delete, move message properties",
		Icon:        "settings",
		Color:       "#f59e0b",
		Properties: []node.PropertySchema{
			{
				Name:        "rules",
				Label:       "Rules",
				Type:        "array",
				Default:     []interface{}{},
				Description: "List of set/delete/move rules",
			},
		},
		Inputs: []node.PortSchema{
			{Name: "input", Label: "Input", Type: "any"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Output", Type: "any"},
		},
		Factory: func() node.Executor {
			return NewSetNode()
		},
	}); err != nil {
		return err
	}

	// Register Link In Node
	if err := registry.Register(&node.NodeInfo{
		Type:        "link-in",
		Name:        "Link In",
		Category:    node.NodeTypeFunction,
		Description: "Receive messages from Link Out nodes (virtual wiring)",
		Icon:        "link",
		Color:       "#6366f1",
		Properties: []node.PropertySchema{
			{
				Name:        "linkId",
				Label:       "Link ID",
				Type:        "string",
				Required:    true,
				Description: "Unique identifier for this link",
			},
			{
				Name:        "scope",
				Label:       "Scope",
				Type:        "select",
				Default:     "global",
				Options:     []string{"global", "flow"},
				Required:    false,
				Description: "Link scope (global or flow)",
			},
			{
				Name:        "flowId",
				Label:       "Flow ID",
				Type:        "string",
				Required:    false,
				Description: "Flow identifier (required for flow scope)",
			},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Output", Type: "any", Description: "Linked messages"},
		},
		Factory: func() node.Executor {
			return NewLinkInNode()
		},
	}); err != nil {
		return err
	}

	// Register Link Out Node
	if err := registry.Register(&node.NodeInfo{
		Type:        "link-out",
		Name:        "Link Out",
		Category:    node.NodeTypeFunction,
		Description: "Send messages to Link In nodes (virtual wiring)",
		Icon:        "link",
		Color:       "#6366f1",
		Properties: []node.PropertySchema{
			{
				Name:        "linkIds",
				Label:       "Link IDs",
				Type:        "array",
				Required:    false,
				Description: "Target link IDs (can be multiple)",
			},
			{
				Name:        "linkId",
				Label:       "Link ID",
				Type:        "string",
				Required:    false,
				Description: "Single target link ID",
			},
			{
				Name:        "scope",
				Label:       "Scope",
				Type:        "select",
				Default:     "global",
				Options:     []string{"global", "flow"},
				Required:    false,
				Description: "Link scope (global or flow)",
			},
			{
				Name:        "flowId",
				Label:       "Flow ID",
				Type:        "string",
				Required:    false,
				Description: "Flow identifier (required for flow scope)",
			},
			{
				Name:        "mode",
				Label:       "Mode",
				Type:        "select",
				Default:     "continue",
				Options:     []string{"continue", "return"},
				Required:    false,
				Description: "Continue passing to outputs or return only",
			},
		},
		Inputs: []node.PortSchema{
			{Name: "input", Label: "Input", Type: "any", Description: "Messages to link"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Output", Type: "any", Description: "Pass-through (continue mode only)"},
		},
		Factory: func() node.Executor {
			return NewLinkOutNode()
		},
	}); err != nil {
		return err
	}

	// Register Trigger Node
	if err := registry.Register(&node.NodeInfo{
		Type:        "trigger",
		Name:        "Trigger",
		Category:    node.NodeTypeFunction,
		Description: "Send message, then optionally send second message after delay",
		Icon:        "timer",
		Color:       "#f59e0b",
		Properties: []node.PropertySchema{
			{
				Name:        "op",
				Label:       "Operation",
				Type:        "select",
				Default:     "send-then-send",
				Options:     []string{"send", "send-then-send", "send-then-nothing"},
				Required:    false,
				Description: "Operation mode",
			},
			{
				Name:        "initialPayload",
				Label:       "Initial Payload",
				Type:        "any",
				Default:     nil,
				Required:    false,
				Description: "Payload to send immediately (null = pass through)",
			},
			{
				Name:        "secondPayload",
				Label:       "Second Payload",
				Type:        "any",
				Default:     nil,
				Required:    false,
				Description: "Payload to send after delay (null = no second message)",
			},
			{
				Name:        "delay",
				Label:       "Delay",
				Type:        "string",
				Default:     "250ms",
				Required:    false,
				Description: "Delay before second message (e.g. 250ms, 1s, 5s)",
			},
			{
				Name:        "duration",
				Label:       "Duration",
				Type:        "string",
				Default:     "",
				Required:    false,
				Description: "Duration for send-then-nothing mode",
			},
			{
				Name:        "extend",
				Label:       "Extend Delay",
				Type:        "boolean",
				Default:     false,
				Required:    false,
				Description: "Extend delay on new messages",
			},
			{
				Name:        "reset",
				Label:       "Reset",
				Type:        "boolean",
				Default:     false,
				Required:    false,
				Description: "Reset on second message",
			},
			{
				Name:        "resetOnNewMsg",
				Label:       "Reset on New Message",
				Type:        "boolean",
				Default:     false,
				Required:    false,
				Description: "Reset timer on any new message",
			},
		},
		Inputs: []node.PortSchema{
			{Name: "input", Label: "Input", Type: "any", Description: "Trigger input"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Output", Type: "any", Description: "Triggered output"},
		},
		Factory: func() node.Executor {
			return NewTriggerNode()
		},
	}); err != nil {
		return err
	}

	// Register RBE (Report By Exception) Node
	if err := registry.Register(&node.NodeInfo{
		Type:        "rbe",
		Name:        "RBE",
		Category:    node.NodeTypeFunction,
		Description: "Report by exception - only pass changed values",
		Icon:        "filter",
		Color:       "#8b5cf6",
		Properties: []node.PropertySchema{
			{
				Name:        "property",
				Label:       "Property",
				Type:        "string",
				Default:     "payload",
				Description: "Property to check for changes",
			},
			{
				Name:        "mode",
				Label:       "Mode",
				Type:        "select",
				Default:     "value",
				Options:     []string{"value", "deadband", "narrowband"},
				Description: "Comparison mode",
			},
			{
				Name:        "bandgap",
				Label:       "Band Gap",
				Type:        "number",
				Default:     0.0,
				Description: "Tolerance for numeric comparisons",
			},
			{
				Name:        "invert",
				Label:       "Invert",
				Type:        "boolean",
				Default:     false,
				Description: "Block changes, pass same values",
			},
		},
		Inputs: []node.PortSchema{
			{Name: "input", Label: "Input", Type: "any"},
		},
		Outputs: []node.PortSchema{
			{Name: "output", Label: "Output", Type: "any"},
		},
		Factory: func() node.Executor {
			return NewRBENode()
		},
	}); err != nil {
		return err
	}

	// Register Comment Node
	if err := registry.Register(&node.NodeInfo{
		Type:        "comment",
		Name:        "Comment",
		Category:    node.NodeTypeFunction,
		Description: "Add documentation notes to flows",
		Icon:        "message-square",
		Color:       "#fbbf24",
		Properties: []node.PropertySchema{
			{
				Name:        "text",
				Label:       "Comment Text",
				Type:        "string",
				Default:     "",
				Description: "Documentation text",
			},
			{
				Name:        "color",
				Label:       "Color",
				Type:        "string",
				Default:     "#fbbf24",
				Description: "Display color",
			},
		},
		Inputs:  []node.PortSchema{},
		Outputs: []node.PortSchema{},
		Factory: func() node.Executor {
			return NewCommentNode()
		},
	}); err != nil {
		return err
	}

	return nil
}
