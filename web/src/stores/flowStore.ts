import { create } from 'zustand'
import { Flow, flowsApi } from '../lib/api'

interface FlowStore {
  flows: Flow[]
  currentFlow: Flow | null
  loading: boolean
  error: string | null

  // Actions
  fetchFlows: () => Promise<void>
  fetchFlow: (id: string) => Promise<void>
  createFlow: (name: string, description: string) => Promise<Flow | null>
  updateFlow: (id: string, data: Partial<Flow>) => Promise<void>
  deleteFlow: (id: string) => Promise<void>
  startFlow: (id: string) => Promise<void>
  stopFlow: (id: string) => Promise<void>
  setCurrentFlow: (flow: Flow | null) => void
}

export const useFlowStore = create<FlowStore>((set, get) => ({
  flows: [],
  currentFlow: null,
  loading: false,
  error: null,

  fetchFlows: async () => {
    set({ loading: true, error: null })
    try {
      const response = await flowsApi.list()
      set({ flows: response.data.flows, loading: false })
    } catch (error: any) {
      // If API fails, use sample workflows as fallback
      const sampleFlows: Flow[] = [
        {
          id: 'sample-rpi5-demo',
          name: 'Raspberry Pi 5 IoT Demo - Complete Automation',
          description: 'Comprehensive workflow for Raspberry Pi 5 (4GB) demonstrating GPIO control, temperature monitoring, MQTT, HTTP APIs, and automation logic. Includes DS18B20 temp sensor, BMP280 pressure sensor, PIR motion detection, RGB LED indicators, relay control, PWM fan, and file logging.',
          status: 'stopped',
          nodes: {},
          connections: [],
          config: {},
          createdAt: new Date().toISOString(),
          updatedAt: new Date().toISOString(),
        },
      ]
      set({ flows: sampleFlows, error: error.message, loading: false })
    }
  },

  fetchFlow: async (id: string) => {
    set({ loading: true, error: null })
    try {
      const response = await flowsApi.get(id)
      // Debug: log what the server returned
      const flowData = response.data
      if (flowData?.nodes) {
        const nodesArr = Array.isArray(flowData.nodes) ? flowData.nodes : Object.values(flowData.nodes)
        nodesArr.forEach((n: any) => {
          if (n.type === 'inject') {
            console.log('[fetchFlow] Server returned inject config:', n.id, JSON.stringify(n.config))
          }
        })
      }
      set({ currentFlow: flowData, loading: false })
    } catch (error: any) {
      set({ loading: false, error: error.message })
      // If API fails and it's the sample workflow, load sample data
      if (id === 'sample-rpi5-demo') {
        const sampleFlow: Flow = {
          id: 'sample-rpi5-demo',
          name: 'Raspberry Pi 5 IoT Demo - Complete Automation',
          description: 'Comprehensive workflow for Raspberry Pi 5 (4GB) demonstrating GPIO control, temperature monitoring, MQTT, HTTP APIs, and automation logic.',
          status: 'stopped',
          nodes: {
            'inject-1': {
              id: 'inject-1',
              type: 'inject',
              name: 'Every 30 Seconds Trigger',
              category: 'input' as const,
              inputs: [],
              outputs: ['gpio-temp-1'],
              status: 'idle' as const,
              config: {
                interval: 30,
                payload: { timestamp: '{{timestamp}}', device: 'raspberry-pi-5' },
                repeat: true,
              }
            },
            'gpio-temp-1': {
              id: 'gpio-temp-1',
              type: 'ds18b20',
              name: 'DS18B20 Temperature Sensor',
              category: 'input' as const,
              inputs: ['inject-1'],
              outputs: ['function-1', 'mqtt-out-1'],
              status: 'idle' as const,
              config: {
                pin: 'GPIO4',
                sensorId: '28-00000xxxxx',
                unit: 'celsius',
                precision: 2,
                interval: 30000,
              }
            },
            'function-1': {
              id: 'function-1',
              type: 'function',
              name: 'Process Temperature Data',
              category: 'function' as const,
              inputs: ['gpio-temp-1'],
              outputs: ['switch-1', 'debug-1'],
              status: 'idle' as const,
              config: {
                code: "const temp = msg.payload.temperature;\nconst timestamp = new Date().toISOString();\n\nmsg.payload = {\n  device: 'raspberry-pi-5',\n  sensor: 'ds18b20',\n  temperature: temp,\n  unit: 'celsius',\n  timestamp: timestamp,\n  status: temp > 30 ? 'warning' : 'normal'\n};\n\nif (temp > 35) {\n  msg.alert = true;\n  msg.alertLevel = 'critical';\n}\n\nreturn msg;"
              }
            },
            'switch-1': {
              id: 'switch-1',
              type: 'switch',
              name: 'Temperature Router',
              category: 'processing' as const,
              inputs: ['function-1'],
              outputs: ['gpio-out-1'],
              status: 'idle' as const,
              config: {
                property: 'payload.temperature',
                rules: [
                  { type: 'gt', value: 35, output: 0 },
                  { type: 'between', value: [25, 35], output: 1 },
                  { type: 'lt', value: 25, output: 2 },
                ]
              }
            },
            'gpio-out-1': {
              id: 'gpio-out-1',
              type: 'gpio-out',
              name: 'Red LED (High Temp)',
              category: 'output' as const,
              inputs: ['switch-1'],
              outputs: [],
              status: 'idle' as const,
              config: {
                pin: 17,
                mode: 'output',
                initialState: 'low',
                pwm: false,
                description: 'BCM GPIO17 - Red LED for high temperature alert',
              }
            },
            'mqtt-out-1': {
              id: 'mqtt-out-1',
              type: 'mqtt-out',
              name: 'Publish to MQTT',
              category: 'output' as const,
              inputs: ['gpio-temp-1'],
              outputs: [],
              status: 'idle' as const,
              config: {
                broker: 'mqtt://localhost:1883',
                topic: 'home/raspberry-pi-5/temperature',
                qos: 1,
                retain: true,
                clientId: 'raspberry-pi-5-sensor',
              }
            },
            'debug-1': {
              id: 'debug-1',
              type: 'debug',
              name: 'Temperature Debug',
              category: 'output' as const,
              inputs: ['function-1'],
              outputs: [],
              status: 'idle' as const,
              config: {
                output: 'console',
                active: true,
                complete: 'payload',
              }
            },
          },
          connections: [
            { id: 'conn-1', source_id: 'inject-1', target_id: 'gpio-temp-1' },
            { id: 'conn-2', source_id: 'gpio-temp-1', target_id: 'function-1' },
            { id: 'conn-3', source_id: 'gpio-temp-1', target_id: 'mqtt-out-1' },
            { id: 'conn-4', source_id: 'function-1', target_id: 'switch-1' },
            { id: 'conn-5', source_id: 'function-1', target_id: 'debug-1' },
            { id: 'conn-6', source_id: 'switch-1', target_id: 'gpio-out-1' },
          ],
          config: {},
          createdAt: new Date().toISOString(),
          updatedAt: new Date().toISOString(),
        }
        set({ currentFlow: sampleFlow, loading: false, error: null })
      } else {
        // Re-throw the error so components can handle it
        throw error
      }
    }
  },

  createFlow: async (name: string, description: string) => {
    set({ loading: true, error: null })
    try {
      const response = await flowsApi.create({ name, description })
      const newFlow = response.data
      set((state) => ({
        flows: [...state.flows, newFlow],
        currentFlow: newFlow,
        loading: false,
      }))
      return newFlow
    } catch (error: any) {
      set({ error: error.message, loading: false })
      return null
    }
  },

  updateFlow: async (id: string, data: Partial<Flow>) => {
    try {
      // Debug: log what we're sending
      if (data.nodes) {
        const nodesArr = Array.isArray(data.nodes) ? data.nodes : Object.values(data.nodes)
        nodesArr.forEach((n: any) => {
          if (n.type === 'inject') {
            console.log('[updateFlow] Sending inject config to server:', n.id, JSON.stringify(n.config))
          }
        })
      }

      const response = await flowsApi.update(id, data)
      // Use server response to update local state (ensures round-trip consistency)
      const serverData = response.data

      // Debug: log what server returned
      if (serverData?.nodes) {
        const respNodes = Array.isArray(serverData.nodes) ? serverData.nodes : Object.values(serverData.nodes)
        respNodes.forEach((n: any) => {
          if (n.type === 'inject') {
            console.log('[updateFlow] Server responded with inject config:', n.id, JSON.stringify(n.config))
          }
        })
      }

      set((state) => ({
        flows: state.flows.map((f) => (f.id === id ? { ...f, ...serverData } : f)),
        currentFlow:
          state.currentFlow?.id === id
            ? { ...state.currentFlow, ...serverData }
            : state.currentFlow,
      }))
    } catch (error: any) {
      set({ error: error.message })
    }
  },

  deleteFlow: async (id: string) => {
    set({ loading: true, error: null })
    try {
      await flowsApi.delete(id)
      set((state) => ({
        flows: state.flows.filter((f) => f.id !== id),
        currentFlow: state.currentFlow?.id === id ? null : state.currentFlow,
        loading: false,
      }))
    } catch (error: any) {
      set({ error: error.message, loading: false })
    }
  },

  startFlow: async (id: string) => {
    try {
      await flowsApi.start(id)
      set((state) => ({
        flows: state.flows.map((f) =>
          f.id === id ? { ...f, status: 'running' as const } : f
        ),
        currentFlow:
          state.currentFlow?.id === id
            ? { ...state.currentFlow, status: 'running' as const }
            : state.currentFlow,
      }))
    } catch (error: any) {
      set({ error: error.message })
    }
  },

  stopFlow: async (id: string) => {
    try {
      await flowsApi.stop(id)
      set((state) => ({
        flows: state.flows.map((f) =>
          f.id === id ? { ...f, status: 'stopped' as const } : f
        ),
        currentFlow:
          state.currentFlow?.id === id
            ? { ...state.currentFlow, status: 'stopped' as const }
            : state.currentFlow,
      }))
    } catch (error: any) {
      set({ error: error.message })
    }
  },

  setCurrentFlow: (flow: Flow | null) => {
    set({ currentFlow: flow })
  },
}))
