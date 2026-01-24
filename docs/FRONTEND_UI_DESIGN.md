# EdgeFlow Frontend UI Design - Modern Best Practices (2025)

## Executive Summary

Based on research of leading workflow automation platforms ([n8n](https://n8n.io/features/), [Node-RED Dashboard 2.0](https://dashboard.flowfuse.com/), and modern UI/UX trends), this document outlines a **lightweight, beautiful, and production-ready** frontend design for EdgeFlow.

**Design Philosophy**: *"The power of code with the ease of a visual interface"*

---

## Research Findings: Industry Best Practices

### n8n UI Excellence (2025)
- **Drag-and-drop canvas** with visual data flow at every step
- **AI-ready interface** with dedicated nodes for OpenAI, Hugging Face, etc.
- **Hybrid approach**: Visual nodes + embedded JavaScript/Python code editors
- **Debugging features**: Re-run individual steps, mock data, execution logs
- **Branch merging**: Visual merge of workflow branches
- **Embedded chat interface** for user interaction

Source: [n8n Features 2025](https://n8n.io/features/) | [n8n Guide 2026](https://hatchworks.com/blog/ai-agents/n8n-guide/)

### Node-RED Dashboard 2.0 (2025)
- **Grid-based layout** (1 unit = 48px)
- **Responsive design** with collapsing/fixed navigation
- **Dark/Light themes** with customization
- **Tab-based organization** for large workflows
- **Single Source of Truth** vs **Multi-Tenancy** patterns
- **Deprecated Angular 1.0** → Modern Vue.js/React stack

Source: [Node-RED Dashboard 2.0](https://dashboard.flowfuse.com/getting-started.html) | [Comparing Dashboards](https://flowfuse.com/blog/2023/03/comparing-node-red-dashboards/)

### Modern UI/UX Trends (2025)
- **AI-powered workflows** with AI as creative collaborator
- **Adaptive interfaces** with smart hiding of non-essential features
- **Seamless Figma-to-code** integration (Webflow pattern)
- **Rapid prototyping** from sketches to high-fidelity
- **Automation focus**: Streamline repetitive tasks
- **Responsive, intuitive, visually efficient**

Source: [Future of UI/UX Design 2025](https://motiongility.com/future-of-ui-ux-design/) | [UX Design Trends 2025](https://www.fullstack.com/labs/resources/blog/top-5-ux-ui-design-trends-in-2025-the-future-of-user-experiences)

### React Flow (Canvas Library)
- **Used by Stripe & Typeform**
- **Highly customizable** nodes and edges
- **Dark/Light mode** toggle
- **Advanced features**: Auto-layout, copy-paste, undo-redo
- **Mobile-friendly** touchpad interactions
- **Plugin system** for extensions

Source: [React Flow](https://reactflow.dev) | [Workflow Builder SDK](https://gitnation.com/contents/building-ai-workflow-editor-ui-in-react-with-workflow-builder-sdk)

---

## EdgeFlow Frontend Architecture

### Technology Stack

```yaml
Core Framework: React 18+
Canvas Library: React Flow (xyflow)
UI Components: shadcn/ui + Radix UI
Styling: TailwindCSS v4
State Management: Zustand (lightweight)
Code Editor: Monaco Editor (VS Code engine)
Build Tool: Vite
Language: TypeScript
RTL Support: Built-in for Persian/Arabic
```

**Why This Stack?**
- **React Flow**: Industry-standard for workflow editors (Stripe, Typeform)
- **shadcn/ui**: Customizable, accessible, copy-paste components
- **Zustand**: 3x smaller than Redux, perfect for edge devices
- **Monaco**: Same editor as VS Code, syntax highlighting
- **Vite**: 10x faster than Webpack, ideal for Raspberry Pi

---

## UI Structure & Layout

### 1. Main Application Layout

```
┌─────────────────────────────────────────────────────────┐
│ Top Bar (48px)                                          │
│ [Logo] [Deploy ▼] [Debug] [Settings]       [User] [⚙] │
├──────┬──────────────────────────────────────────────────┤
│      │                                                  │
│      │            Canvas Workspace                      │
│  L   │                                                  │
│  e   │         [Node-based Flow Editor]                │
│  f   │                                                  │
│  t   │                                                  │
│      │                                                  │
│  S   │                                                  │
│  i   ├──────────────────────────────────────────────────┤
│  d   │ Bottom Panel (Collapsible)                      │
│  e   │ [Debug Log] [Context Data] [Execution History]  │
│  b   │                                                  │
│  a   │                                                  │
│  r   │                                                  │
│      │                                                  │
│(280) │                                                  │
└──────┴──────────────────────────────────────────────────┘
```

### 2. Left Sidebar (Collapsible)

**Node Palette** with categorized nodes:
```
📥 Input
  • Inject
  • HTTP In
  • MQTT In
  • GPIO In

🔄 Function
  • Function (JS)
  • Python
  • Exec (Shell)
  • Template
  • Switch
  • Change
  • Range

📤 Output
  • Debug
  • HTTP Response
  • MQTT Out
  • GPIO Out

⚡ Hardware
  • PWM
  • Servo
  • I2C
  • SPI
  • Serial

🌐 Network
  • HTTP Request
  • WebSocket
  • TCP
  • UDP

💾 Storage
  • File
  • Database
  • Context Store
```

**Features**:
- Search/filter nodes (Cmd+K)
- Drag-and-drop to canvas
- Categorized accordion
- Icon + description preview
- Recently used section

### 3. Canvas Workspace (React Flow)

**Grid-based canvas** with:
- **Infinite scrolling** (pan with middle mouse or space+drag)
- **Zoom controls** (10% - 400%)
- **Mini-map** (bottom-right corner)
- **Snap to grid** (optional, 10px grid)
- **Multi-select** (Shift+click or drag box)
- **Keyboard shortcuts**:
  - `Ctrl+C/V`: Copy/Paste
  - `Ctrl+Z/Y`: Undo/Redo
  - `Del`: Delete
  - `Ctrl+A`: Select all
  - `Ctrl+F`: Find node

**Connection Rules**:
- Bezier curves for edges
- Animated flow direction (dots moving)
- Color-coded by message type:
  - Blue: Data
  - Red: Error
  - Green: Success
  - Orange: Warning
- Click edge to add breakpoint/debug

### 4. Node Design (Custom Components)

#### Standard Node Layout
```
┌────────────────────────────┐
│ [Icon] Node Name      [⚙] │ ← Header (drag handle)
├────────────────────────────┤
│ • Input                    │ ← Ports (left)
│                            │
│   [Mini Preview]           │ ← Content area
│   Config summary           │
│                            │
│                     Output │ ← Ports (right)
└────────────────────────────┘
```

#### Node States
- **Idle**: Gray border
- **Running**: Blue pulsing border
- **Success**: Green checkmark
- **Error**: Red border + icon
- **Disabled**: Dashed border, opacity 50%

#### Node Variants

**Compact Node** (for GPIO/simple nodes):
```
┌──────────────┐
│ [⚡] GPIO 18 │
│ •           •│
└──────────────┘
```

**Expanded Node** (for code editors):
```
┌──────────────────────────────┐
│ [🐍] Python Script      [×] │
├──────────────────────────────┤
│ import json                  │
│ msg['temp'] = 25.5          │
│ return msg                   │
│                              │
│ [Edit Code ↗]               │
└──────────────────────────────┘
```

### 5. Node Configuration Panel (Right Sidebar)

**Slides in from right** when node is selected (400px width):

```
┌──────────────────────────────────┐
│ [×] Python Script                │
│                                  │
│ ┌─ Properties ─────────────────┐ │
│ │                              │ │
│ │ Python Code *                │ │
│ │ ┌──────────────────────────┐ │ │
│ │ │ # Python code editor     │ │ │
│ │ │ import json              │ │ │
│ │ │                          │ │ │
│ │ └──────────────────────────┘ │ │
│ │                              │ │
│ │ Python Path                  │ │
│ │ [python3          ▼]         │ │
│ │                              │ │
│ │ Timeout (seconds)            │ │
│ │ [30                    ]     │ │
│ │                              │ │
│ │ ☐ Use Virtual Environment    │ │
│ │                              │ │
│ └──────────────────────────────┘ │
│                                  │
│ ┌─ Advanced ────────────────────┐│
│ │ Virtual Env Path             ││
│ │ Template Variables           ││
│ └──────────────────────────────┘│
│                                  │
│        [Cancel]  [Apply]         │
└──────────────────────────────────┘
```

**Features**:
- Live validation
- Required fields marked with *
- Help tooltips (? icon)
- Expandable advanced section
- Code editor with syntax highlighting
- Template variable autocomplete

### 6. Top Bar Features

```
┌──────────────────────────────────────────────────────┐
│ [Logo] EdgeFlow                                      │
│                                                      │
│ [Deploy ▼] [Debug] [+Import] [Export]               │
│                                                      │
│                           [Search] [👤 User] [⚙]    │
└──────────────────────────────────────────────────────┘
```

**Deploy Button** (Primary action):
- Dropdown options:
  - Full Deploy (restart all)
  - Modified Nodes Only
  - Current Flow Only
- Shows deployment status
- Confirmation dialog for full deploy

**Debug Toggle**:
- Enable/disable debug messages
- Shows message count badge
- Opens debug panel

**User Menu**:
- Account settings
- Theme toggle
- Language (EN/FA)
- Logout

### 7. Bottom Panel (Resizable, Collapsible)

**Tabs**:
1. **Debug Log**
   ```
   [Clear] [Filter ▼] [⏸ Pause]

   14:32:15 [GPIO In] Pin 18: HIGH
   14:32:16 [Python] Temp: 25.5°C
   14:32:16 [MQTT Out] Published to /sensor/temp
   ```

2. **Context Data** (Live view)
   ```
   Node Context | Flow Context | Global Context

   node-123: {
     "counter": 42,
     "lastValue": 25.5
   }
   ```

3. **Execution History**
   ```
   [Date Filter ▼] [Export CSV]

   12:00:00  Flow-A  Success  2.3s  [View]
   11:55:00  Flow-B  Error    0.8s  [Debug]
   ```

---

## Color Palette & Theming

### Light Theme (Default)
```css
--background: #FFFFFF
--foreground: #0F172A
--card: #F8FAFC
--border: #E2E8F0
--primary: #3B82F6    /* Blue */
--success: #10B981    /* Green */
--warning: #F59E0B    /* Orange */
--error: #EF4444      /* Red */
--accent: #8B5CF6     /* Purple for GPIO */
```

### Dark Theme
```css
--background: #0F172A
--foreground: #F1F5F9
--card: #1E293B
--border: #334155
--primary: #60A5FA
--success: #34D399
--warning: #FBBF24
--error: #F87171
--accent: #A78BFA
```

### Node Category Colors
```
Input:    #3B82F6  (Blue)
Output:   #10B981  (Green)
Function: #8B5CF6  (Purple)
Hardware: #F59E0B  (Orange)
Network:  #06B6D4  (Cyan)
Storage:  #6366F1  (Indigo)
```

---

## Component Library (shadcn/ui)

### Core Components
```
✓ Button (primary, secondary, ghost, outline)
✓ Input (text, number, password)
✓ Select / Combobox
✓ Checkbox / Radio
✓ Switch (toggle)
✓ Slider
✓ Textarea
✓ Dialog / Modal
✓ Dropdown Menu
✓ Tooltip
✓ Tabs
✓ Accordion
✓ Badge
✓ Alert
✓ Toast (notifications)
✓ Command (Cmd+K palette)
✓ Popover
✓ Sheet (slide-in panel)
```

### Custom Components
```
✓ NodeCard (workflow node)
✓ CodeEditor (Monaco wrapper)
✓ FlowCanvas (React Flow wrapper)
✓ MiniMap
✓ DebugPanel
✓ ContextViewer
✓ PropertyForm (dynamic node config)
✓ ConnectionHandle
```

---

## Responsive Design

### Breakpoints
```
sm: 640px   (Tablets)
md: 768px   (Small laptops)
lg: 1024px  (Desktop)
xl: 1280px  (Large desktop)
2xl: 1536px (Ultra-wide)
```

### Mobile Adaptation
- **< 768px**: Hide sidebar by default
- **Touch gestures**: Pinch to zoom, two-finger pan
- **Bottom sheet**: Replace right sidebar with bottom sheet
- **Simplified nodes**: Show compact view only
- **Hamburger menu**: Collapse top bar items

---

## Performance Optimization

### Bundle Size Targets
```
Initial Load:  < 300 KB (gzipped)
React Flow:    ~50 KB
Monaco Editor: Lazy-loaded (~500 KB)
shadcn/ui:     Tree-shaken, ~40 KB
TailwindCSS:   Purged, ~20 KB
Total Target:  < 1 MB (with code editor)
```

### Optimization Strategies
1. **Code splitting**: Lazy load Monaco, heavy components
2. **Virtualization**: Node palette uses virtual scrolling
3. **Debouncing**: Canvas interactions debounced (100ms)
4. **Web Workers**: JSON parsing for large flows
5. **Service Worker**: Offline-first PWA
6. **Asset optimization**: SVG icons, WebP images

### Target Performance (Raspberry Pi 4)
```
Initial Load:     < 2 seconds
Canvas Response:  < 16ms (60 FPS)
Node Add:         < 50ms
Deploy:           < 500ms
Memory Usage:     < 200 MB
```

---

## Accessibility (WCAG 2.1 Level AA)

### Requirements
- ✓ Keyboard navigation (Tab, Arrow keys, Enter, Esc)
- ✓ Screen reader support (ARIA labels)
- ✓ Focus indicators (2px outline)
- ✓ Color contrast 4.5:1 minimum
- ✓ Text resizing up to 200%
- ✓ No flashing content
- ✓ Skip to main content link
- ✓ Alt text for all icons

### Keyboard Shortcuts
```
Global:
  Ctrl+K       Command palette
  Ctrl+S       Save flow
  Ctrl+D       Deploy
  Ctrl+/       Toggle debug panel

Canvas:
  Space+Drag   Pan canvas
  Ctrl+Scroll  Zoom
  Ctrl+A       Select all
  Ctrl+C/V     Copy/Paste
  Ctrl+Z/Y     Undo/Redo
  Del          Delete

Node:
  Enter        Edit node
  Tab          Next field
  Esc          Close panel
```

---

## RTL (Right-to-Left) Support

### Persian/Arabic Layout
```
✓ Automatic text direction detection
✓ Mirrored sidebar (right to left)
✓ Reversed flow direction (nodes flow RTL)
✓ RTL-aware grid layout
✓ Bidirectional text support
✓ Locale-aware date/time formatting
```

### Implementation
```typescript
// TailwindCSS RTL plugin
import rtlPlugin from 'tailwindcss-rtl';

// Auto-detect language
const isRTL = document.dir === 'rtl' ||
              navigator.language.startsWith('fa') ||
              navigator.language.startsWith('ar');
```

---

## User Experience Enhancements

### Onboarding Flow
1. **Welcome modal** with video tutorial
2. **Interactive tour** (first-time users)
3. **Sample flows** library (one-click import)
4. **Tooltips** on first interaction
5. **Contextual help** (? icons)

### Smart Features
- **Auto-save** (every 30 seconds)
- **Version history** (restore previous versions)
- **Undo/Redo** (unlimited history)
- **Node search** (Cmd+K)
- **Quick actions** (right-click context menu)
- **Drag-and-drop** import (JSON files)
- **Copy-paste** between flows
- **Bulk operations** (multi-select + action)

### Feedback & Notifications
```
Toast Notifications (top-right):
  ✓ Success: "Flow deployed successfully"
  ⚠ Warning: "GPIO pin already in use"
  ✗ Error: "Failed to connect to MQTT broker"
  ℹ Info: "Auto-save enabled"
```

### Loading States
- **Skeleton screens** (avoid spinners)
- **Progress bars** for long operations
- **Optimistic updates** (assume success)
- **Graceful degradation** (show cached data)

---

## File Structure

```
frontend/
├── public/
│   ├── icons/           # Node icons (SVG)
│   └── samples/         # Sample flows
├── src/
│   ├── components/
│   │   ├── ui/         # shadcn components
│   │   ├── nodes/      # Custom node components
│   │   ├── canvas/     # React Flow wrapper
│   │   ├── panels/     # Sidebars, sheets
│   │   └── layout/     # Layout components
│   ├── lib/
│   │   ├── api.ts      # Backend API client
│   │   ├── store.ts    # Zustand store
│   │   └── utils.ts    # Utilities
│   ├── hooks/
│   │   ├── useFlow.ts
│   │   ├── useNodes.ts
│   │   └── useDebug.ts
│   ├── types/
│   │   ├── node.ts
│   │   ├── flow.ts
│   │   └── message.ts
│   ├── styles/
│   │   └── globals.css
│   ├── App.tsx
│   └── main.tsx
├── package.json
├── vite.config.ts
├── tailwind.config.ts
└── tsconfig.json
```

---

## Implementation Roadmap

### Phase 1: Core Infrastructure (Week 1-2)
- [ ] Setup Vite + React + TypeScript
- [ ] Install React Flow + shadcn/ui
- [ ] Create base layout (top bar, sidebar, canvas)
- [ ] Setup Zustand store
- [ ] Configure TailwindCSS with custom theme
- [ ] Implement dark/light mode toggle

### Phase 2: Canvas & Nodes (Week 3-4)
- [ ] Integrate React Flow canvas
- [ ] Create base NodeCard component
- [ ] Implement drag-and-drop from palette
- [ ] Add connection logic
- [ ] Implement zoom/pan controls
- [ ] Add mini-map

### Phase 3: Node Configuration (Week 5-6)
- [ ] Build property panel (right sidebar)
- [ ] Integrate Monaco Editor for code nodes
- [ ] Create dynamic form generator
- [ ] Add validation logic
- [ ] Implement Apply/Cancel actions
- [ ] Add template variable autocomplete

### Phase 4: Node Types (Week 7-8)
- [ ] GPIO In/Out nodes
- [ ] Python/Exec nodes (code editors)
- [ ] Hardware nodes (PWM, Servo, SPI, Serial)
- [ ] Function nodes (JS, Switch, Change)
- [ ] Input/Output nodes (Inject, Debug, HTTP)
- [ ] Test all 30+ node types

### Phase 5: Features (Week 9-10)
- [ ] Debug panel with live logs
- [ ] Context data viewer
- [ ] Execution history
- [ ] Deploy mechanism (full/partial)
- [ ] Import/Export flows
- [ ] Copy/Paste/Undo/Redo
- [ ] Node search (Cmd+K)

### Phase 6: Polish & Optimization (Week 11-12)
- [ ] Performance optimization (lazy loading)
- [ ] Accessibility testing (WCAG 2.1 AA)
- [ ] RTL support for Persian
- [ ] Mobile responsive design
- [ ] Onboarding tour
- [ ] Sample flows library
- [ ] User testing & bug fixes

### Phase 7: Testing & Documentation (Week 13-14)
- [ ] Unit tests (Jest + React Testing Library)
- [ ] E2E tests (Playwright)
- [ ] Performance benchmarks (Lighthouse)
- [ ] User documentation
- [ ] Video tutorials
- [ ] Deploy to production

---

## API Integration

### REST API Endpoints
```typescript
// Flow Management
GET    /api/flows           // List all flows
GET    /api/flows/:id       // Get flow by ID
POST   /api/flows           // Create new flow
PUT    /api/flows/:id       // Update flow
DELETE /api/flows/:id       // Delete flow

// Deployment
POST   /api/deploy          // Deploy flows
GET    /api/deploy/status   // Deployment status

// Nodes
GET    /api/nodes           // Get node registry
GET    /api/nodes/:type     // Get node schema

// Debug
GET    /api/debug/logs      // Get debug logs
POST   /api/debug/clear     // Clear logs

// Context
GET    /api/context/:scope  // Get context data
PUT    /api/context/:scope  // Set context data

// Execution
GET    /api/executions      // Get execution history
GET    /api/executions/:id  // Get execution details
```

### WebSocket Events
```typescript
// Real-time updates
ws://localhost:3000/ws

Events:
  - node:status        // Node execution status
  - debug:message      // Debug log message
  - context:update     // Context data changed
  - deploy:progress    // Deployment progress
  - error:occurred     // Runtime error
```

---

## Comparison with Competitors

| Feature | EdgeFlow | n8n | Node-RED | Make.com |
|---------|----------|-----|----------|----------|
| **Open Source** | ✅ MIT | ✅ Apache 2.0 | ✅ Apache 2.0 | ❌ Proprietary |
| **Self-Hosted** | ✅ | ✅ | ✅ | ❌ |
| **IoT Focus** | ✅ Raspberry Pi | ❌ | ✅ | ❌ |
| **Python Support** | ✅ Native | ✅ via exec | ✅ via exec | ❌ |
| **GPIO Control** | ✅ Native | ❌ | ✅ | ❌ |
| **Modern UI** | ✅ React Flow | ✅ Vue.js | ⚠️ jQuery | ✅ Custom |
| **Dark Mode** | ✅ | ✅ | ⚠️ Limited | ✅ |
| **RTL Support** | ✅ Persian | ❌ | ❌ | ❌ |
| **Lightweight** | ✅ < 1 MB | ⚠️ 5 MB | ✅ < 1 MB | N/A |
| **Offline-First** | ✅ PWA | ❌ | ⚠️ Limited | ❌ |
| **Mobile Support** | ✅ Responsive | ⚠️ Limited | ❌ | ✅ |

**EdgeFlow Advantages**:
- ✅ Built for edge devices (Raspberry Pi optimization)
- ✅ Native hardware control (GPIO, SPI, I2C, Serial)
- ✅ Persian language support (RTL)
- ✅ Lightweight bundle (< 1 MB vs 5 MB)
- ✅ Modern React stack (vs jQuery/Angular 1)
- ✅ Offline-first PWA
- ✅ True open source (MIT license)

---

## Conclusion

This design leverages the best practices from industry leaders (n8n, Node-RED, React Flow) while optimizing for:
1. **Performance**: < 1 MB bundle, 60 FPS on Raspberry Pi
2. **Developer Experience**: Modern React + TypeScript + shadcn/ui
3. **User Experience**: Intuitive drag-and-drop, smart features
4. **Accessibility**: WCAG 2.1 AA compliant
5. **Localization**: RTL support for Persian/Arabic
6. **Hardware Integration**: Native GPIO, Python, Shell support

**Next Steps**: Begin Phase 1 implementation with core infrastructure setup.

---

## Sources & References

1. [n8n Workflow Automation Features](https://n8n.io/features/)
2. [n8n Guide 2026: AI Workflow Deep Dive](https://hatchworks.com/blog/ai-agents/n8n-guide/)
3. [Node-RED Dashboard 2.0 Getting Started](https://dashboard.flowfuse.com/getting-started.html)
4. [Comparing Node-RED Dashboards Solutions](https://flowfuse.com/blog/2023/03/comparing-node-red-dashboards/)
5. [Future of UI/UX Design: 2026 Trends](https://motiongility.com/future-of-ui-ux-design/)
6. [UX Design Trends 2025 - FullStack](https://www.fullstack.com/labs/resources/blog/top-5-ux-ui-design-trends-in-2025-the-future-of-user-experiences)
7. [React Flow - Node-Based UIs](https://reactflow.dev)
8. [Building AI Workflow Editor UI with React](https://gitnation.com/contents/building-ai-workflow-editor-ui-in-react-with-workflow-builder-sdk)
9. [Modern UI/UX Workflow in 2025: AI Design](https://www.zignuts.com/blog/modern-ui-ux-workflow-ai-design)
10. [shadcn/ui Component Library](https://ui.shadcn.com/)

---

**Document Version**: 1.0
**Last Updated**: 2026-01-20
**Author**: EdgeFlow Development Team
**Status**: ✅ Ready for Implementation
