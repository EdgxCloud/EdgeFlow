import React, { useState } from 'react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Dialog, DialogContent, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Download, Eye, Tag } from 'lucide-react';
import { Flow } from '@/types/flow';

export interface SampleFlow {
  id: string;
  name: string;
  description: string;
  category: string;
  tags: string[];
  difficulty: 'beginner' | 'intermediate' | 'advanced';
  flow: Flow;
}

// Sample flows using n8n-style [x, y] array positions
const sampleFlows: SampleFlow[] = [
  {
    id: 'hello-world',
    name: 'Hello World',
    description: 'Simple flow that sends a "Hello World" message',
    category: 'Basic',
    tags: ['beginner', 'inject', 'debug'],
    difficulty: 'beginner',
    flow: {
      id: 'sample-hello-world',
      name: 'Hello World',
      nodes: [
        {
          id: 'inject-1',
          type: 'inject',
          name: 'Trigger',
          position: [100, 100],
          config: { payload: 'Hello World' },
        },
        {
          id: 'debug-1',
          type: 'debug',
          name: 'Debug',
          position: [300, 100],
          config: { console: true },
        },
      ],
      connections: [{ id: 'e1', source: 'inject-1', target: 'debug-1' }],
    },
  },
  {
    id: 'http-api',
    name: 'HTTP API Consumer',
    description: 'Fetch data from an HTTP API and log the response',
    category: 'Network',
    tags: ['http', 'api', 'json'],
    difficulty: 'beginner',
    flow: {
      id: 'sample-http-api',
      name: 'HTTP API Consumer',
      nodes: [
        {
          id: 'inject-1',
          type: 'inject',
          name: 'Timer',
          position: [100, 100],
          config: { interval: 60 },
        },
        {
          id: 'http-1',
          type: 'http-request',
          name: 'HTTP Request',
          position: [300, 100],
          config: { method: 'GET', url: 'https://api.example.com/data' },
        },
        {
          id: 'json-1',
          type: 'json-parser',
          name: 'Parse JSON',
          position: [500, 100],
          config: { action: 'parse' },
        },
        {
          id: 'debug-1',
          type: 'debug',
          name: 'Debug',
          position: [700, 100],
          config: { console: true },
        },
      ],
      connections: [
        { id: 'e1', source: 'inject-1', target: 'http-1' },
        { id: 'e2', source: 'http-1', target: 'json-1' },
        { id: 'e3', source: 'json-1', target: 'debug-1' },
      ],
    },
  },
  {
    id: 'mqtt-broker',
    name: 'MQTT Data Logger',
    description: 'Subscribe to MQTT topic and log messages',
    category: 'IoT',
    tags: ['mqtt', 'iot', 'logging'],
    difficulty: 'intermediate',
    flow: {
      id: 'sample-mqtt-logger',
      name: 'MQTT Data Logger',
      nodes: [
        {
          id: 'mqtt-in-1',
          type: 'mqtt-in',
          name: 'MQTT In',
          position: [100, 100],
          config: { broker: 'tcp://localhost:1883', topic: 'sensors/#' },
        },
        {
          id: 'json-1',
          type: 'json-parser',
          name: 'Parse JSON',
          position: [300, 100],
          config: { action: 'parse' },
        },
        {
          id: 'function-1',
          type: 'function',
          name: 'Add Timestamp',
          position: [500, 100],
          config: { code: 'msg.payload.timestamp = Date.now(); return msg;' },
        },
        {
          id: 'debug-1',
          type: 'debug',
          name: 'Debug',
          position: [700, 100],
          config: { console: true },
        },
      ],
      connections: [
        { id: 'e1', source: 'mqtt-in-1', target: 'json-1' },
        { id: 'e2', source: 'json-1', target: 'function-1' },
        { id: 'e3', source: 'function-1', target: 'debug-1' },
      ],
    },
  },
  {
    id: 'temperature-monitor',
    name: 'Temperature Monitor',
    description: 'Monitor temperature sensor and send alerts',
    category: 'Sensors',
    tags: ['sensor', 'temperature', 'alert'],
    difficulty: 'intermediate',
    flow: {
      id: 'sample-temp-monitor',
      name: 'Temperature Monitor',
      nodes: [
        {
          id: 'inject-1',
          type: 'inject',
          name: 'Timer',
          position: [100, 100],
          config: { interval: 10 },
        },
        {
          id: 'dht-1',
          type: 'dht',
          name: 'DHT22 Sensor',
          position: [300, 100],
          config: { pin: 4, type: 'dht22' },
        },
        {
          id: 'switch-1',
          type: 'switch',
          name: 'Temp Check',
          position: [500, 50],
          config: {
            property: 'payload.temperature',
            rules: [{ t: 'gt', v: 30 }],
          },
        },
        {
          id: 'template-1',
          type: 'template',
          name: 'Alert Message',
          position: [700, 50],
          config: {
            template: 'Alert: Temperature is {{temperature}}°C',
            syntax: 'mustache',
          },
        },
        {
          id: 'debug-1',
          type: 'debug',
          name: 'Debug',
          position: [500, 150],
          config: { console: true },
        },
      ],
      connections: [
        { id: 'e1', source: 'inject-1', target: 'dht-1' },
        { id: 'e2', source: 'dht-1', target: 'switch-1' },
        { id: 'e3', source: 'switch-1', target: 'template-1', sourceOutput: '0' },
        { id: 'e4', source: 'dht-1', target: 'debug-1' },
      ],
    },
  },
  {
    id: 'gpio-relay',
    name: 'GPIO Relay Control',
    description: 'Control relay based on button input',
    category: 'GPIO',
    tags: ['gpio', 'relay', 'button'],
    difficulty: 'beginner',
    flow: {
      id: 'sample-gpio-relay',
      name: 'GPIO Relay Control',
      nodes: [
        {
          id: 'gpio-in-1',
          type: 'gpio-in',
          name: 'Button Input',
          position: [100, 100],
          config: { pin: 17, pullMode: 'up', edgeMode: 'falling' },
        },
        {
          id: 'gpio-out-1',
          type: 'gpio-out',
          name: 'Relay Output',
          position: [300, 100],
          config: { pin: 27 },
        },
      ],
      connections: [{ id: 'e1', source: 'gpio-in-1', target: 'gpio-out-1' }],
    },
  },
];

export interface SampleFlowsLibraryProps {
  onImport: (flow: Flow) => void;
}

export function SampleFlowsLibrary({ onImport }: SampleFlowsLibraryProps) {
  const [selectedCategory, setSelectedCategory] = useState<string>('all');
  const [selectedFlow, setSelectedFlow] = useState<SampleFlow | null>(null);

  const categories = ['all', ...new Set(sampleFlows.map((f) => f.category))];

  const filteredFlows =
    selectedCategory === 'all'
      ? sampleFlows
      : sampleFlows.filter((f) => f.category === selectedCategory);

  const getDifficultyColor = (difficulty: string) => {
    switch (difficulty) {
      case 'beginner':
        return 'bg-green-500';
      case 'intermediate':
        return 'bg-yellow-500';
      case 'advanced':
        return 'bg-red-500';
      default:
        return 'bg-gray-500';
    }
  };

  return (
    <div className="space-y-4">
      <div className="flex gap-2 flex-wrap">
        {categories.map((category) => (
          <Button
            key={category}
            variant={selectedCategory === category ? 'default' : 'outline'}
            onClick={() => setSelectedCategory(category)}
            className="capitalize"
          >
            {category}
          </Button>
        ))}
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
        {filteredFlows.map((sample) => (
          <Card key={sample.id} className="hover:shadow-lg transition-shadow">
            <CardHeader>
              <CardTitle className="flex items-center justify-between">
                <span>{sample.name}</span>
                <Badge className={getDifficultyColor(sample.difficulty)}>
                  {sample.difficulty}
                </Badge>
              </CardTitle>
              <CardDescription>{sample.description}</CardDescription>
            </CardHeader>
            <CardContent>
              <div className="flex flex-wrap gap-1 mb-4">
                {sample.tags.map((tag) => (
                  <Badge key={tag} variant="secondary" className="text-xs">
                    <Tag className="w-3 h-3 mr-1" />
                    {tag}
                  </Badge>
                ))}
              </div>
              <div className="flex gap-2">
                <Button
                  size="sm"
                  variant="outline"
                  onClick={() => setSelectedFlow(sample)}
                  className="flex-1"
                >
                  <Eye className="w-4 h-4 mr-1" />
                  Preview
                </Button>
                <Button
                  size="sm"
                  onClick={() => onImport(sample.flow)}
                  className="flex-1"
                >
                  <Download className="w-4 h-4 mr-1" />
                  Import
                </Button>
              </div>
            </CardContent>
          </Card>
        ))}
      </div>

      <Dialog open={!!selectedFlow} onOpenChange={() => setSelectedFlow(null)}>
        <DialogContent className="max-w-3xl">
          <DialogHeader>
            <DialogTitle>{selectedFlow?.name}</DialogTitle>
          </DialogHeader>
          <div className="space-y-4">
            <p className="text-muted-foreground">{selectedFlow?.description}</p>
            <div>
              <h4 className="font-semibold mb-2">Flow Details:</h4>
              <ul className="space-y-1">
                <li>Nodes: {selectedFlow?.flow.nodes.length}</li>
                <li>Edges: {selectedFlow?.flow.edges.length}</li>
                <li>Category: {selectedFlow?.category}</li>
              </ul>
            </div>
            <pre className="bg-muted p-4 rounded text-xs overflow-auto max-h-96">
              {JSON.stringify(selectedFlow?.flow, null, 2)}
            </pre>
            <Button onClick={() => selectedFlow && onImport(selectedFlow.flow)}>
              Import This Flow
            </Button>
          </div>
        </DialogContent>
      </Dialog>
    </div>
  );
}
