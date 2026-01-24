/**
 * Test Components Page
 *
 * Test page for Phase 1 and Phase 2 components
 */

import { useState } from 'react'
import { ColorPicker, IconPicker, JSONEditor, CronBuilder, CodeEditor } from '@/components/Common'
import { NodeConfigDialog } from '@/components/NodeConfig/NodeConfigDialog'
import {
  ChartEditor,
  GaugeEditor,
  FormBuilderEditor,
  TableEditor,
  TextEditor,
  ButtonEditor,
  SliderEditor,
  SwitchEditor,
} from '@/components/NodeConfig/Dashboard'
import {
  GPIOPinSelector,
  MQTTTopicBuilder,
  SwitchRuleBuilder,
  HTTPRequestBuilder,
  ChangeTransformBuilder,
} from '@/components/NodeConfig/Specialized'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import type { NodeConfig } from '@/types/node'

export default function TestComponents() {
  // Test states for reusable components
  const [color, setColor] = useState('#3b82f6')
  const [icon, setIcon] = useState('Zap')
  const [json, setJson] = useState({ test: true, value: 42 })
  const [cron, setCron] = useState('* * * * *')
  const [jsCode, setJsCode] = useState('return msg;')
  const [pyCode, setPyCode] = useState('return msg')

  // Test state for node config dialog
  const [selectedNode, setSelectedNode] = useState<NodeConfig | null>(null)

  // Test states for dashboard editors
  const [chartConfig, setChartConfig] = useState({
    chartType: 'line',
    maxDataSize: 50,
    legend: true,
    series: [
      { name: 'Temperature', color: '#ef4444', visible: true },
      { name: 'Humidity', color: '#3b82f6', visible: true },
    ],
  })

  const [gaugeConfig, setGaugeConfig] = useState({
    gaugeType: 'semi',
    min: 0,
    max: 100,
    units: '%',
    showValue: true,
    showSectors: true,
    needleColor: '#3b82f6',
    backgroundColor: '#e5e7eb',
    sectors: [
      { from: 0, to: 30, color: '#22c55e', label: 'Low' },
      { from: 30, to: 70, color: '#f59e0b', label: 'Medium' },
      { from: 70, to: 100, color: '#ef4444', label: 'High' },
    ],
  })

  const [formConfig, setFormConfig] = useState({
    submitButtonText: 'Submit',
    showResetButton: true,
    formLayout: 'vertical',
    fields: [
      {
        id: 'field1',
        type: 'text',
        label: 'Username',
        name: 'username',
        required: true,
      },
      {
        id: 'field2',
        type: 'email',
        label: 'Email',
        name: 'email',
        required: true,
      },
    ],
  })

  const [tableConfig, setTableConfig] = useState({
    maxRows: 100,
    pagination: true,
    pageSize: 10,
    striped: true,
    bordered: true,
    columns: [
      { id: 'col1', header: 'ID', key: 'id', type: 'number', align: 'right' },
      { id: 'col2', header: 'Name', key: 'name', type: 'string', align: 'left' },
      { id: 'col3', header: 'Status', key: 'status', type: 'badge', align: 'center' },
    ],
  })

  // Test states for specialized editors
  const [gpioConfig, setGpioConfig] = useState({
    pin: 17,
    mode: 'output',
    pullMode: 'off',
    initialValue: false,
    debounceMs: 0,
  })

  const [mqttConfig, setMqttConfig] = useState({
    topic: 'home/+/temperature',
    qos: 0,
    retain: false,
    wildcardMode: true,
  })

  const [switchConfig, setSwitchConfig] = useState({
    rules: [
      {
        id: 'rule1',
        property: 'payload',
        propertyType: 'msg.payload',
        operator: 'gt',
        value: '100',
        outputIndex: 0,
      },
      {
        id: 'rule2',
        property: 'payload',
        propertyType: 'msg.payload',
        operator: 'lte',
        value: '100',
        outputIndex: 1,
      },
    ],
    checkAll: false,
    outputCount: 2,
  })

  const [httpConfig, setHttpConfig] = useState({
    method: 'POST',
    url: 'https://api.example.com/data',
    headers: [
      { id: '1', key: 'Content-Type', value: 'application/json', enabled: true },
    ],
    authType: 'bearer',
    authConfig: { token: 'your-token-here' },
    bodyType: 'json',
    body: { temperature: 25, humidity: 60 },
    timeout: 30000,
    followRedirects: true,
    validateSSL: true,
  })

  const [changeConfig, setChangeConfig] = useState({
    rules: [
      {
        id: 'rule1',
        action: 'set',
        property: 'temperature',
        propertyType: 'msg',
        valueType: 'msg',
        value: 'payload.temp',
      },
      {
        id: 'rule2',
        action: 'delete',
        property: 'raw',
        propertyType: 'msg',
      },
    ],
  })

  // Sample nodes for testing
  const testNodes: NodeConfig[] = [
    {
      id: 'test-chart-1',
      type: 'dashboard-chart',
      name: 'Temperature Chart',
      config: {
        id: 'chart-1',
        label: 'Temperature',
        chartType: 'line',
        maxDataSize: 50,
        legend: true,
      },
    },
    {
      id: 'test-gauge-1',
      type: 'dashboard-gauge',
      name: 'CPU Gauge',
      config: {
        id: 'gauge-1',
        label: 'CPU Usage',
        min: 0,
        max: 100,
        units: '%',
      },
    },
    {
      id: 'test-button-1',
      type: 'dashboard-button',
      name: 'Trigger Button',
      config: {
        id: 'button-1',
        label: 'Trigger',
        buttonLabel: 'Click Me',
        bgColor: '#3b82f6',
        fgColor: '#ffffff',
      },
    },
  ]

  return (
    <div className="min-h-screen bg-background p-8">
      <div className="max-w-6xl mx-auto space-y-8">
        {/* Header */}
        <div className="space-y-2">
          <h1 className="text-4xl font-bold">Phase 1 Component Testing</h1>
          <p className="text-muted-foreground">
            Test all foundational components built in Phase 1
          </p>
        </div>

        <Tabs defaultValue="reusable" className="w-full">
          <TabsList className="grid w-full grid-cols-4">
            <TabsTrigger value="reusable">Reusable Components</TabsTrigger>
            <TabsTrigger value="dashboard">Dashboard Editors</TabsTrigger>
            <TabsTrigger value="specialized">Specialized Editors</TabsTrigger>
            <TabsTrigger value="dialogs">Node Config Dialogs</TabsTrigger>
          </TabsList>

          {/* Reusable Components Tab */}
          <TabsContent value="reusable" className="space-y-6 mt-6">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              {/* Color Picker */}
              <Card>
                <CardHeader>
                  <CardTitle>Color Picker</CardTitle>
                  <CardDescription>
                    Preset colors + custom hex input + recent colors
                  </CardDescription>
                </CardHeader>
                <CardContent className="space-y-4">
                  <ColorPicker
                    value={color}
                    onChange={setColor}
                    label="Select Color"
                    allowAlpha={false}
                  />
                  <div className="flex items-center gap-3 p-4 bg-muted rounded-lg">
                    <div
                      className="w-16 h-16 rounded-lg border-2 border-white shadow-md"
                      style={{ backgroundColor: color }}
                    />
                    <div>
                      <p className="text-sm font-semibold">Selected Color</p>
                      <p className="text-xs text-muted-foreground font-mono">{color}</p>
                    </div>
                  </div>
                </CardContent>
              </Card>

              {/* Icon Picker */}
              <Card>
                <CardHeader>
                  <CardTitle>Icon Picker</CardTitle>
                  <CardDescription>1000+ Lucide icons with search</CardDescription>
                </CardHeader>
                <CardContent className="space-y-4">
                  <IconPicker value={icon} onChange={setIcon} label="Select Icon" />
                  <div className="flex items-center gap-3 p-4 bg-muted rounded-lg">
                    <div className="w-16 h-16 rounded-lg bg-primary/10 flex items-center justify-center">
                      {/* Icon will render here */}
                      <span className="text-2xl">🔥</span>
                    </div>
                    <div>
                      <p className="text-sm font-semibold">Selected Icon</p>
                      <p className="text-xs text-muted-foreground font-mono">{icon}</p>
                    </div>
                  </div>
                </CardContent>
              </Card>

              {/* JSON Editor */}
              <Card className="md:col-span-2">
                <CardHeader>
                  <CardTitle>JSON Editor</CardTitle>
                  <CardDescription>Monaco-based editor with validation</CardDescription>
                </CardHeader>
                <CardContent>
                  <JSONEditor
                    value={json}
                    onChange={setJson}
                    label="Edit JSON"
                    height={200}
                    showValidation={true}
                  />
                  <div className="mt-4 p-4 bg-muted rounded-lg">
                    <p className="text-sm font-semibold mb-2">Parsed Value:</p>
                    <pre className="text-xs">{JSON.stringify(json, null, 2)}</pre>
                  </div>
                </CardContent>
              </Card>

              {/* Cron Builder */}
              <Card className="md:col-span-2">
                <CardHeader>
                  <CardTitle>Cron Builder</CardTitle>
                  <CardDescription>Visual cron expression builder</CardDescription>
                </CardHeader>
                <CardContent>
                  <CronBuilder
                    value={cron}
                    onChange={setCron}
                    label="Schedule"
                    enableSeconds={false}
                    showPreview={true}
                  />
                  <div className="mt-4 p-4 bg-muted rounded-lg">
                    <p className="text-sm font-semibold mb-2">Cron Expression:</p>
                    <code className="text-sm bg-background px-3 py-2 rounded border">
                      {cron}
                    </code>
                  </div>
                </CardContent>
              </Card>

              {/* JavaScript Code Editor */}
              <Card className="md:col-span-2">
                <CardHeader>
                  <CardTitle>JavaScript Code Editor</CardTitle>
                  <CardDescription>Monaco editor with snippets and autocomplete</CardDescription>
                </CardHeader>
                <CardContent>
                  <CodeEditor
                    value={jsCode}
                    onChange={setJsCode}
                    label="JavaScript Function"
                    language="javascript"
                    height={250}
                    showLanguageSelector={true}
                  />
                </CardContent>
              </Card>

              {/* Python Code Editor */}
              <Card className="md:col-span-2">
                <CardHeader>
                  <CardTitle>Python Code Editor</CardTitle>
                  <CardDescription>Monaco editor with Python support</CardDescription>
                </CardHeader>
                <CardContent>
                  <CodeEditor
                    value={pyCode}
                    onChange={setPyCode}
                    label="Python Function"
                    language="python"
                    height={250}
                    showLanguageSelector={true}
                  />
                </CardContent>
              </Card>
            </div>
          </TabsContent>

          {/* Dashboard Editors Tab */}
          <TabsContent value="dashboard" className="space-y-6 mt-6">
            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              {/* Chart Editor */}
              <Card className="md:col-span-2">
                <CardHeader>
                  <CardTitle>Chart Editor</CardTitle>
                  <CardDescription>
                    Series management with drag-to-reorder and color picker
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  <ChartEditor config={chartConfig} onChange={setChartConfig} />
                  <div className="mt-4 p-4 bg-muted rounded-lg">
                    <p className="text-sm font-semibold mb-2">Current Configuration:</p>
                    <pre className="text-xs overflow-x-auto">{JSON.stringify(chartConfig, null, 2)}</pre>
                  </div>
                </CardContent>
              </Card>

              {/* Gauge Editor */}
              <Card className="md:col-span-2">
                <CardHeader>
                  <CardTitle>Gauge Editor</CardTitle>
                  <CardDescription>
                    Sector color zones with drag-to-reorder and validation
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  <GaugeEditor config={gaugeConfig} onChange={setGaugeConfig} />
                  <div className="mt-4 p-4 bg-muted rounded-lg">
                    <p className="text-sm font-semibold mb-2">Current Configuration:</p>
                    <pre className="text-xs overflow-x-auto">{JSON.stringify(gaugeConfig, null, 2)}</pre>
                  </div>
                </CardContent>
              </Card>

              {/* Form Builder Editor */}
              <Card className="md:col-span-2">
                <CardHeader>
                  <CardTitle>Form Builder Editor</CardTitle>
                  <CardDescription>
                    Visual form builder with 14 field types and drag-to-reorder
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  <FormBuilderEditor config={formConfig} onChange={setFormConfig} />
                  <div className="mt-4 p-4 bg-muted rounded-lg">
                    <p className="text-sm font-semibold mb-2">Current Configuration:</p>
                    <pre className="text-xs overflow-x-auto">{JSON.stringify(formConfig, null, 2)}</pre>
                  </div>
                </CardContent>
              </Card>

              {/* Table Editor */}
              <Card className="md:col-span-2">
                <CardHeader>
                  <CardTitle>Table Editor</CardTitle>
                  <CardDescription>
                    Column management with 7 types and drag-to-reorder
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  <TableEditor config={tableConfig} onChange={setTableConfig} />
                  <div className="mt-4 p-4 bg-muted rounded-lg">
                    <p className="text-sm font-semibold mb-2">Current Configuration:</p>
                    <pre className="text-xs overflow-x-auto">{JSON.stringify(tableConfig, null, 2)}</pre>
                  </div>
                </CardContent>
              </Card>

              {/* Simple Widget Editors */}
              <Card>
                <CardHeader>
                  <CardTitle>Text Editor</CardTitle>
                  <CardDescription>Text formatting with markdown/HTML support</CardDescription>
                </CardHeader>
                <CardContent>
                  <TextEditor
                    config={{ text: 'Sample text', fontSize: 16, color: '#000000' }}
                    onChange={(config) => console.log('Text config:', config)}
                  />
                </CardContent>
              </Card>

              <Card>
                <CardHeader>
                  <CardTitle>Button Editor</CardTitle>
                  <CardDescription>Button with icon and color customization</CardDescription>
                </CardHeader>
                <CardContent>
                  <ButtonEditor
                    config={{ buttonLabel: 'Click Me', bgColor: '#3b82f6', fgColor: '#ffffff' }}
                    onChange={(config) => console.log('Button config:', config)}
                  />
                </CardContent>
              </Card>

              <Card>
                <CardHeader>
                  <CardTitle>Slider Editor</CardTitle>
                  <CardDescription>Range slider with min/max/step configuration</CardDescription>
                </CardHeader>
                <CardContent>
                  <SliderEditor
                    config={{ min: 0, max: 100, step: 1, defaultValue: 50 }}
                    onChange={(config) => console.log('Slider config:', config)}
                  />
                </CardContent>
              </Card>

              <Card>
                <CardHeader>
                  <CardTitle>Switch Editor</CardTitle>
                  <CardDescription>Toggle switch with custom on/off states</CardDescription>
                </CardHeader>
                <CardContent>
                  <SwitchEditor
                    config={{ onLabel: 'On', offLabel: 'Off', defaultValue: false }}
                    onChange={(config) => console.log('Switch config:', config)}
                  />
                </CardContent>
              </Card>
            </div>

            {/* Instructions */}
            <Card>
              <CardHeader>
                <CardTitle>Testing Instructions</CardTitle>
              </CardHeader>
              <CardContent className="space-y-4">
                <div>
                  <h4 className="text-sm font-semibold mb-2">1. Test Chart Editor</h4>
                  <ul className="text-sm text-muted-foreground space-y-1 ml-4 list-disc">
                    <li>Add/remove series using the "+ Add Series" button</li>
                    <li>Drag series to reorder (use the grip handle)</li>
                    <li>Change series colors using color picker</li>
                    <li>Toggle series visibility</li>
                    <li>Change chart type and observe series compatibility</li>
                  </ul>
                </div>

                <div>
                  <h4 className="text-sm font-semibold mb-2">2. Test Gauge Editor</h4>
                  <ul className="text-sm text-muted-foreground space-y-1 ml-4 list-disc">
                    <li>Add/remove color sectors</li>
                    <li>Drag sectors to reorder</li>
                    <li>Adjust sector ranges and observe validation</li>
                    <li>Try creating overlapping sectors (should show error)</li>
                    <li>Use "Auto Distribute" to evenly space sectors</li>
                  </ul>
                </div>

                <div>
                  <h4 className="text-sm font-semibold mb-2">3. Test Form Builder</h4>
                  <ul className="text-sm text-muted-foreground space-y-1 ml-4 list-disc">
                    <li>Add different field types from dropdown</li>
                    <li>Click fields to expand/collapse configuration</li>
                    <li>Drag fields to reorder</li>
                    <li>Duplicate fields using copy button</li>
                    <li>Configure field-specific options (options for select, rows for textarea)</li>
                  </ul>
                </div>

                <div>
                  <h4 className="text-sm font-semibold mb-2">4. Test Table Editor</h4>
                  <ul className="text-sm text-muted-foreground space-y-1 ml-4 list-disc">
                    <li>Add/remove columns</li>
                    <li>Drag columns to reorder</li>
                    <li>Change column types and alignments</li>
                    <li>Toggle column visibility</li>
                    <li>Configure table-level options (pagination, striped, etc.)</li>
                  </ul>
                </div>
              </CardContent>
            </Card>
          </TabsContent>

          {/* Specialized Editors Tab */}
          <TabsContent value="specialized" className="space-y-6 mt-6">
            <div className="grid grid-cols-1 gap-6">
              {/* GPIO Pin Selector */}
              <Card>
                <CardHeader>
                  <CardTitle>GPIO Pin Selector</CardTitle>
                  <CardDescription>
                    Visual Raspberry Pi 5 GPIO pinout with pin mode configuration
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  <GPIOPinSelector config={gpioConfig} onChange={setGpioConfig} />
                  <div className="mt-4 p-4 bg-muted rounded-lg">
                    <p className="text-sm font-semibold mb-2">Current Configuration:</p>
                    <pre className="text-xs overflow-x-auto">{JSON.stringify(gpioConfig, null, 2)}</pre>
                  </div>
                </CardContent>
              </Card>

              {/* MQTT Topic Builder */}
              <Card>
                <CardHeader>
                  <CardTitle>MQTT Topic Builder</CardTitle>
                  <CardDescription>
                    Visual MQTT topic builder with wildcard support
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  <MQTTTopicBuilder
                    value={mqttConfig}
                    onChange={setMqttConfig}
                    mode="subscribe"
                  />
                  <div className="mt-4 p-4 bg-muted rounded-lg">
                    <p className="text-sm font-semibold mb-2">Current Configuration:</p>
                    <pre className="text-xs overflow-x-auto">{JSON.stringify(mqttConfig, null, 2)}</pre>
                  </div>
                </CardContent>
              </Card>

              {/* Switch Rule Builder */}
              <Card>
                <CardHeader>
                  <CardTitle>Switch Rule Builder</CardTitle>
                  <CardDescription>
                    Visual rule builder for conditional message routing
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  <SwitchRuleBuilder value={switchConfig} onChange={setSwitchConfig} />
                  <div className="mt-4 p-4 bg-muted rounded-lg">
                    <p className="text-sm font-semibold mb-2">Current Configuration:</p>
                    <pre className="text-xs overflow-x-auto">{JSON.stringify(switchConfig, null, 2)}</pre>
                  </div>
                </CardContent>
              </Card>

              {/* HTTP Request Builder */}
              <Card>
                <CardHeader>
                  <CardTitle>HTTP Request Builder</CardTitle>
                  <CardDescription>
                    Complete HTTP request configuration with auth and body
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  <HTTPRequestBuilder value={httpConfig} onChange={setHttpConfig} />
                  <div className="mt-4 p-4 bg-muted rounded-lg">
                    <p className="text-sm font-semibold mb-2">Current Configuration:</p>
                    <pre className="text-xs overflow-x-auto">{JSON.stringify(httpConfig, null, 2)}</pre>
                  </div>
                </CardContent>
              </Card>

              {/* Change Transform Builder */}
              <Card>
                <CardHeader>
                  <CardTitle>Change Transform Builder</CardTitle>
                  <CardDescription>
                    Message transformation rules (set, change, delete, move)
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  <ChangeTransformBuilder value={changeConfig} onChange={setChangeConfig} />
                  <div className="mt-4 p-4 bg-muted rounded-lg">
                    <p className="text-sm font-semibold mb-2">Current Configuration:</p>
                    <pre className="text-xs overflow-x-auto">{JSON.stringify(changeConfig, null, 2)}</pre>
                  </div>
                </CardContent>
              </Card>
            </div>

            {/* Instructions */}
            <Card>
              <CardHeader>
                <CardTitle>Testing Instructions</CardTitle>
              </CardHeader>
              <CardContent className="space-y-4">
                <div>
                  <h4 className="text-sm font-semibold mb-2">1. Test GPIO Pin Selector</h4>
                  <ul className="text-sm text-muted-foreground space-y-1 ml-4 list-disc">
                    <li>Click on GPIO pins in the visual pinout diagram</li>
                    <li>Hover over pins to see detailed information</li>
                    <li>Select pin mode (input, output, PWM, I2C, SPI, UART)</li>
                    <li>Configure mode-specific settings (pull resistor, initial value, debounce)</li>
                    <li>Note: Power, ground, and reserved pins are not selectable</li>
                  </ul>
                </div>

                <div>
                  <h4 className="text-sm font-semibold mb-2">2. Test MQTT Topic Builder</h4>
                  <ul className="text-sm text-muted-foreground space-y-1 ml-4 list-disc">
                    <li>Build topics using the level-based builder</li>
                    <li>Add/remove topic levels</li>
                    <li>Toggle wildcards (+ for single-level, # for multi-level)</li>
                    <li>Try common topic patterns</li>
                    <li>Configure QoS and retain settings</li>
                  </ul>
                </div>

                <div>
                  <h4 className="text-sm font-semibold mb-2">3. Test Switch Rule Builder</h4>
                  <ul className="text-sm text-muted-foreground space-y-1 ml-4 list-disc">
                    <li>Add/remove rules</li>
                    <li>Drag rules to reorder priority</li>
                    <li>Select property types (msg.payload, msg.topic, etc.)</li>
                    <li>Choose operators (equals, greater than, contains, etc.)</li>
                    <li>Assign rules to different outputs</li>
                    <li>Toggle "check all rules" mode</li>
                  </ul>
                </div>

                <div>
                  <h4 className="text-sm font-semibold mb-2">4. Test HTTP Request Builder</h4>
                  <ul className="text-sm text-muted-foreground space-y-1 ml-4 list-disc">
                    <li>Select HTTP method (GET, POST, PUT, etc.)</li>
                    <li>Add/remove custom headers</li>
                    <li>Configure authentication (Basic, Bearer, API Key)</li>
                    <li>Set request body (JSON, form data, raw text)</li>
                    <li>Adjust timeout and SSL validation settings</li>
                  </ul>
                </div>

                <div>
                  <h4 className="text-sm font-semibold mb-2">5. Test Change Transform Builder</h4>
                  <ul className="text-sm text-muted-foreground space-y-1 ml-4 list-disc">
                    <li>Add transformation rules (set, change, delete, move)</li>
                    <li>Drag rules to reorder execution order</li>
                    <li>Select property types (msg, flow, global)</li>
                    <li>Choose value types (string, number, JSON, timestamp, etc.)</li>
                    <li>Test move operations to rename properties</li>
                  </ul>
                </div>
              </CardContent>
            </Card>
          </TabsContent>

          {/* Node Config Dialogs Tab */}
          <TabsContent value="dialogs" className="space-y-6 mt-6">
            <Card>
              <CardHeader>
                <CardTitle>Node Configuration Dialogs</CardTitle>
                <CardDescription>
                  Test the dynamic node configuration dialog with different node types
                </CardDescription>
              </CardHeader>
              <CardContent className="space-y-4">
                <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                  {testNodes.map((node) => (
                    <Button
                      key={node.id}
                      variant="outline"
                      className="h-auto flex-col items-start p-4 gap-2"
                      onClick={() => setSelectedNode(node)}
                    >
                      <div className="text-sm font-semibold">{node.name}</div>
                      <div className="text-xs text-muted-foreground">{node.type}</div>
                    </Button>
                  ))}
                </div>

                <div className="p-4 bg-muted rounded-lg">
                  <p className="text-sm text-muted-foreground">
                    Click a button above to open the configuration dialog for that node type.
                    The dialog will dynamically render properties based on the node's schema
                    from the registry API.
                  </p>
                </div>

                <div className="p-4 bg-yellow-50 dark:bg-yellow-950/20 border border-yellow-200 dark:border-yellow-900 rounded-lg">
                  <p className="text-sm font-semibold text-yellow-900 dark:text-yellow-100 mb-1">
                    ⚠️ Note: API Required
                  </p>
                  <p className="text-xs text-yellow-700 dark:text-yellow-300">
                    The backend API endpoint{' '}
                    <code className="bg-yellow-100 dark:bg-yellow-900 px-1 rounded">
                      GET /api/v1/registry/nodes/:type
                    </code>{' '}
                    must be running to fetch node schemas. If the API is not available, the
                    dialog will show a loading state.
                  </p>
                </div>
              </CardContent>
            </Card>
          </TabsContent>
        </Tabs>

        {/* Instructions */}
        <Card>
          <CardHeader>
            <CardTitle>Testing Instructions</CardTitle>
          </CardHeader>
          <CardContent className="space-y-4">
            <div>
              <h4 className="text-sm font-semibold mb-2">1. Test Reusable Components</h4>
              <ul className="text-sm text-muted-foreground space-y-1 ml-4 list-disc">
                <li>ColorPicker: Select preset colors, try custom hex, check recent colors</li>
                <li>IconPicker: Search for icons, browse categories, check recent icons</li>
                <li>JSONEditor: Edit JSON, test validation, try format/minify buttons</li>
                <li>CronBuilder: Try different tabs, use presets, test custom expressions</li>
              </ul>
            </div>

            <div>
              <h4 className="text-sm font-semibold mb-2">2. Test Node Config Dialogs</h4>
              <ul className="text-sm text-muted-foreground space-y-1 ml-4 list-disc">
                <li>Click on node buttons to open configuration dialogs</li>
                <li>Verify properties render correctly from schema</li>
                <li>Test validation (try leaving required fields empty)</li>
                <li>Test save/cancel buttons</li>
                <li>Check unsaved changes warning</li>
              </ul>
            </div>

            <div>
              <h4 className="text-sm font-semibold mb-2">3. Backend Requirements</h4>
              <ul className="text-sm text-muted-foreground space-y-1 ml-4 list-disc">
                <li>Ensure EdgeFlow server is running on http://localhost:8080</li>
                <li>
                  API endpoint <code>/api/v1/registry/nodes</code> should return node list
                </li>
                <li>
                  API endpoint <code>/api/v1/registry/nodes/:type</code> should return node
                  schema
                </li>
              </ul>
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Node Config Dialog */}
      {selectedNode && (
        <NodeConfigDialog
          node={selectedNode}
          flowId="test-flow"
          onClose={() => setSelectedNode(null)}
          onSave={(nodeId, config) => {
            console.log('Saved node config:', nodeId, config)
            setSelectedNode(null)
          }}
        />
      )}
    </div>
  )
}
