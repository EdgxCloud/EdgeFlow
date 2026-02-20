package main

import (
	"time"

	"github.com/edgeflow/edgeflow/internal/logger"
	"github.com/edgeflow/edgeflow/internal/storage"
	"go.uber.org/zap"
)

// seedDefaultFlows creates example IoT workflow templates when the data directory is empty.
// These give new users ready-to-run examples on first launch.
func seedDefaultFlows(store *storage.FileStorage) {
	flows, err := store.ListFlows()
	if err != nil {
		logger.Warn("Could not list flows for seeding", zap.Error(err))
		return
	}
	if len(flows) > 0 {
		return // already have flows, skip seeding
	}

	logger.Info("No flows found — seeding default example workflows...")

	defaults := []*storage.Flow{
		temperatureMonitoringFlow(),
		dataProcessingPipelineFlow(),
		heartbeatSystemMonitorFlow(),
	}

	for _, f := range defaults {
		f.CreatedAt = time.Now()
		f.UpdatedAt = time.Now()
		if err := store.SaveFlow(f); err != nil {
			logger.Error("Failed to seed flow", zap.String("name", f.Name), zap.Error(err))
		} else {
			logger.Info("Seeded default flow", zap.String("name", f.Name), zap.String("id", f.ID))
		}
	}
}

// -----------------------------------------------------------------------
// Flow 1: Temperature Monitoring & Alerting
// -----------------------------------------------------------------------
// Inject (every 10s) → Function-Node (simulate temp) → If (>35°C)
//   → true:  Template (alert) → Debug
//   → false: Debug (normal)
// + Comment
func temperatureMonitoringFlow() *storage.Flow {
	const (
		flowID    = "flow-default-temp-monitor"
		nInject   = "node-temp-inject"
		nFunc     = "node-temp-simulate"
		nIf       = "node-temp-threshold"
		nTemplate = "node-temp-alert-tmpl"
		nDbgAlert = "node-temp-debug-alert"
		nDbgNorm  = "node-temp-debug-normal"
		nComment  = "node-temp-comment"
		cInFunc   = "conn-temp-1"
		cFuncIf   = "conn-temp-2"
		cIfTmpl   = "conn-temp-3"
		cTmplDbg  = "conn-temp-4"
		cIfNorm   = "conn-temp-5"
	)

	return &storage.Flow{
		ID:          flowID,
		Name:        "Temperature Monitoring & Alerting",
		Description: "Simulates an IoT temperature sensor that triggers alerts when readings exceed 35°C. Demonstrates inject, function-node, conditional routing, template formatting, and debug output.",
		Status:      "idle",
		Nodes: []map[string]interface{}{
			{
				"id":   nInject,
				"type": "inject",
				"name": "Sensor Read",
				"config": map[string]interface{}{
					"intervalType":  "seconds",
					"intervalValue": 10,
					"repeat":        true,
					"payload":       map[string]interface{}{},
					"topic":         "sensor/temperature",
					"position":      map[string]interface{}{"x": 100, "y": 200},
				},
			},
			{
				"id":   nFunc,
				"type": "function-node",
				"name": "Simulate Temperature",
				"config": map[string]interface{}{
					"code": "msg.payload.temperature = 15 + msg.payload._ts * 0\nset sensor_id = \"TMP-001\"\nset unit = \"C\"\nset location = \"Server Room A\"\nreturn msg",
					"position": map[string]interface{}{"x": 300, "y": 200},
				},
			},
			{
				"id":   nIf,
				"type": "if",
				"name": "Temp > 35°C?",
				"config": map[string]interface{}{
					"field":    "temperature",
					"operator": "gt",
					"value":    "35",
					"position": map[string]interface{}{"x": 500, "y": 200},
				},
			},
			{
				"id":   nTemplate,
				"type": "template",
				"name": "Format Alert",
				"config": map[string]interface{}{
					"template": "ALERT: Temperature {{temperature}}°C at {{location}} exceeds threshold (35°C). Sensor: {{sensor_id}}",
					"field":    "alert_message",
					"syntax":   "mustache",
					"position": map[string]interface{}{"x": 700, "y": 100},
				},
			},
			{
				"id":   nDbgAlert,
				"type": "debug",
				"name": "Alert Output",
				"config": map[string]interface{}{
					"output_to": "console",
					"complete":  true,
					"position":  map[string]interface{}{"x": 900, "y": 100},
				},
			},
			{
				"id":   nDbgNorm,
				"type": "debug",
				"name": "Normal Reading",
				"config": map[string]interface{}{
					"output_to": "console",
					"complete":  false,
					"position":  map[string]interface{}{"x": 700, "y": 300},
				},
			},
			{
				"id":   nComment,
				"type": "comment",
				"name": "About This Flow",
				"config": map[string]interface{}{
					"text":     "Temperature Monitoring & Alerting\n\nThis flow simulates an IoT temperature sensor.\n- Inject node triggers every 10 seconds\n- Function generates simulated temperature data\n- If node checks if temperature exceeds 35°C\n- Alerts are formatted and sent to debug output\n\nTo test: Click Deploy, then check the Debug panel.",
					"color":    "#3b82f6",
					"position": map[string]interface{}{"x": 100, "y": 400},
				},
			},
		},
		Connections: []map[string]interface{}{
			{"id": cInFunc, "source": nInject, "target": nFunc},
			{"id": cFuncIf, "source": nFunc, "target": nIf},
			{"id": cIfTmpl, "source": nIf, "target": nTemplate},
			{"id": cTmplDbg, "source": nTemplate, "target": nDbgAlert},
			{"id": cIfNorm, "source": nIf, "target": nDbgNorm},
		},
	}
}

// -----------------------------------------------------------------------
// Flow 2: Data Processing Pipeline
// -----------------------------------------------------------------------
// Inject (every 30s) → Function-Node (multi-sensor) → Change (enrich) → Filter (humidity>80)
//   → match:    Template (warning) → Debug
//   → no_match: Debug (normal)
// + Comment
func dataProcessingPipelineFlow() *storage.Flow {
	const (
		flowID    = "flow-default-data-pipeline"
		nInject   = "node-pipe-inject"
		nFunc     = "node-pipe-sensors"
		nChange   = "node-pipe-enrich"
		nFilter   = "node-pipe-humidity"
		nTemplate = "node-pipe-warn-tmpl"
		nDbgWarn  = "node-pipe-debug-warn"
		nDbgNorm  = "node-pipe-debug-normal"
		nComment  = "node-pipe-comment"
		cInFunc   = "conn-pipe-1"
		cFuncChg  = "conn-pipe-2"
		cChgFilt  = "conn-pipe-3"
		cFiltTmpl = "conn-pipe-4"
		cTmplDbg  = "conn-pipe-5"
		cFiltNorm = "conn-pipe-6"
	)

	return &storage.Flow{
		ID:          flowID,
		Name:        "Data Processing Pipeline",
		Description: "Multi-sensor data enrichment and filtering pipeline. Demonstrates inject, function-node for data generation, change node for property enrichment, and filter-based routing.",
		Status:      "idle",
		Nodes: []map[string]interface{}{
			{
				"id":   nInject,
				"type": "inject",
				"name": "Sensor Array",
				"config": map[string]interface{}{
					"intervalType":  "seconds",
					"intervalValue": 30,
					"repeat":        true,
					"payload":       map[string]interface{}{},
					"topic":         "sensors/array",
					"position":      map[string]interface{}{"x": 100, "y": 200},
				},
			},
			{
				"id":   nFunc,
				"type": "function-node",
				"name": "Generate Sensor Data",
				"config": map[string]interface{}{
					"code": "set temperature = 22.5\nset humidity = 65\nset pressure = 1013.25\nset light = 450\nreturn msg",
					"position": map[string]interface{}{"x": 300, "y": 200},
				},
			},
			{
				"id":   nChange,
				"type": "change",
				"name": "Enrich Data",
				"config": map[string]interface{}{
					"rules": []interface{}{
						map[string]interface{}{
							"t":    "set",
							"p":    "device_id",
							"to":   "EDGE-001",
							"tot":  "str",
						},
						map[string]interface{}{
							"t":    "set",
							"p":    "location",
							"to":   "Building A - Floor 2",
							"tot":  "str",
						},
						map[string]interface{}{
							"t":    "set",
							"p":    "status",
							"to":   "active",
							"tot":  "str",
						},
					},
					"position": map[string]interface{}{"x": 500, "y": 200},
				},
			},
			{
				"id":   nFilter,
				"type": "filter",
				"name": "High Humidity?",
				"config": map[string]interface{}{
					"property":  "humidity",
					"operator":  "gt",
					"value":     "80",
					"valueType": "number",
					"position":  map[string]interface{}{"x": 700, "y": 200},
				},
			},
			{
				"id":   nTemplate,
				"type": "template",
				"name": "Format Warning",
				"config": map[string]interface{}{
					"template": "WARNING: High humidity detected!\nDevice: {{device_id}} at {{location}}\nHumidity: {{humidity}}%\nTemperature: {{temperature}}°C",
					"field":    "warning_message",
					"syntax":   "mustache",
					"position": map[string]interface{}{"x": 900, "y": 100},
				},
			},
			{
				"id":   nDbgWarn,
				"type": "debug",
				"name": "Warning Output",
				"config": map[string]interface{}{
					"output_to": "console",
					"complete":  true,
					"position":  map[string]interface{}{"x": 1100, "y": 100},
				},
			},
			{
				"id":   nDbgNorm,
				"type": "debug",
				"name": "Normal Data",
				"config": map[string]interface{}{
					"output_to": "console",
					"complete":  false,
					"position":  map[string]interface{}{"x": 900, "y": 300},
				},
			},
			{
				"id":   nComment,
				"type": "comment",
				"name": "About This Flow",
				"config": map[string]interface{}{
					"text":     "Data Processing Pipeline\n\nThis flow demonstrates a multi-stage data processing pipeline.\n- Inject triggers every 30 seconds\n- Function generates multi-sensor readings\n- Change node enriches data with device info\n- Filter routes based on humidity threshold (>80%)\n\nTry modifying the humidity value in the function node to trigger warnings.",
					"color":    "#10b981",
					"position": map[string]interface{}{"x": 100, "y": 400},
				},
			},
		},
		Connections: []map[string]interface{}{
			{"id": cInFunc, "source": nInject, "target": nFunc},
			{"id": cFuncChg, "source": nFunc, "target": nChange},
			{"id": cChgFilt, "source": nChange, "target": nFilter},
			{"id": cFiltTmpl, "source": nFilter, "target": nTemplate},
			{"id": cTmplDbg, "source": nTemplate, "target": nDbgWarn},
			{"id": cFiltNorm, "source": nFilter, "target": nDbgNorm},
		},
	}
}

// -----------------------------------------------------------------------
// Flow 3: Heartbeat & System Monitor
// -----------------------------------------------------------------------
// Schedule (every minute) → Function-Node (status) → RBE (change detect)
//   → Template (format) → Debug
// + Comment
func heartbeatSystemMonitorFlow() *storage.Flow {
	const (
		flowID    = "flow-default-heartbeat"
		nSchedule = "node-hb-schedule"
		nFunc     = "node-hb-status"
		nRBE      = "node-hb-rbe"
		nTemplate = "node-hb-template"
		nDebug    = "node-hb-debug"
		nComment  = "node-hb-comment"
		cSchFunc  = "conn-hb-1"
		cFuncRBE  = "conn-hb-2"
		cRBETmpl  = "conn-hb-3"
		cTmplDbg  = "conn-hb-4"
	)

	return &storage.Flow{
		ID:          flowID,
		Name:        "Heartbeat & System Monitor",
		Description: "Periodic system health check with change detection. Demonstrates scheduled triggers, system status simulation, RBE (Report By Exception) filtering, and template-based reporting.",
		Status:      "idle",
		Nodes: []map[string]interface{}{
			{
				"id":   nSchedule,
				"type": "schedule",
				"name": "Every Minute",
				"config": map[string]interface{}{
					"cron":     "0 * * * * *",
					"payload":  map[string]interface{}{},
					"topic":    "system/heartbeat",
					"timezone": "Local",
					"position": map[string]interface{}{"x": 100, "y": 200},
				},
			},
			{
				"id":   nFunc,
				"type": "function-node",
				"name": "System Status",
				"config": map[string]interface{}{
					"code": "set status = \"healthy\"\nset cpu_percent = 42\nset memory_percent = 68\nset disk_percent = 55\nset uptime_hours = 720\nset hostname = \"edgeflow-pi\"\nreturn msg",
					"position": map[string]interface{}{"x": 300, "y": 200},
				},
			},
			{
				"id":   nRBE,
				"type": "rbe",
				"name": "Status Changed?",
				"config": map[string]interface{}{
					"property": "status",
					"mode":     "value",
					"bandgap":  0,
					"invert":   false,
					"position": map[string]interface{}{"x": 500, "y": 200},
				},
			},
			{
				"id":   nTemplate,
				"type": "template",
				"name": "Format Report",
				"config": map[string]interface{}{
					"template": "System Health Report\n====================\nHost: {{hostname}}\nStatus: {{status}}\nCPU: {{cpu_percent}}%\nMemory: {{memory_percent}}%\nDisk: {{disk_percent}}%\nUptime: {{uptime_hours}} hours",
					"field":    "report",
					"syntax":   "mustache",
					"position": map[string]interface{}{"x": 700, "y": 200},
				},
			},
			{
				"id":   nDebug,
				"type": "debug",
				"name": "Status Output",
				"config": map[string]interface{}{
					"output_to": "console",
					"complete":  true,
					"position":  map[string]interface{}{"x": 900, "y": 200},
				},
			},
			{
				"id":   nComment,
				"type": "comment",
				"name": "About This Flow",
				"config": map[string]interface{}{
					"text":     "Heartbeat & System Monitor\n\nThis flow demonstrates a periodic health monitoring pattern.\n- Schedule node triggers every minute via cron\n- Function generates simulated system metrics\n- RBE (Report By Exception) only passes messages when the status changes\n- Template formats a readable status report\n\nThe RBE node prevents flooding: only the first message and status changes are reported.",
					"color":    "#f59e0b",
					"position": map[string]interface{}{"x": 100, "y": 400},
				},
			},
		},
		Connections: []map[string]interface{}{
			{"id": cSchFunc, "source": nSchedule, "target": nFunc},
			{"id": cFuncRBE, "source": nFunc, "target": nRBE},
			{"id": cRBETmpl, "source": nRBE, "target": nTemplate},
			{"id": cTmplDbg, "source": nTemplate, "target": nDebug},
		},
	}
}
