# EdgeFlow Frontend Phase 1 - Testing Guide

## ✅ What's Been Implemented

### Phase 1: Foundation & Infrastructure (COMPLETE)

1. **Type System** - Complete TypeScript interfaces
2. **API Client** - REST API integration for node registry
3. **State Management** - `useNodeConfig` hook with validation
4. **Core Components** - NodeConfigDialog, PropertyField
5. **Reusable Components** - ColorPicker, IconPicker, JSONEditor, CronBuilder
6. **Test Page** - Comprehensive component testing page

## 🚀 How to Test

### Step 1: Start the Backend Server

```bash
cd "c:\Users\Administrator\Desktop\Project\Farhotech IOT Edge"
.\edgeflow.exe
```

**Expected Output**:
```
╔═══════════════════════════════════════╗
║       EdgeFlow v0.1.0                ║
║   پلتفرم اتوماسیون سبک Edge و IoT    ║
╚═══════════════════════════════════════╝
✅ Core nodes registered successfully
✅ Dashboard widgets (14) registered successfully
✅ Node registration complete
Server starting on http://0.0.0.0:8080
```

### Step 2: Frontend is Already Running

The frontend dev server is already running on **http://localhost:3000**

### Step 3: Open the Test Page

Navigate to: **http://localhost:3000/test**

## 📋 Test Checklist

### ✅ Reusable Components Tab

#### 1. Color Picker
- [ ] Click "Select Color" button
- [ ] Select a preset color (should update immediately)
- [ ] Try custom hex input (`#FF5733`)
- [ ] Use native color picker (right side)
- [ ] Check recent colors appear after selection
- [ ] Verify selected color shows in preview box

#### 2. Icon Picker
- [ ] Click "Select Icon" button
- [ ] Search for an icon (try "home", "user", "settings")
- [ ] Browse different categories (All, UI, Devices)
- [ ] Click an icon to select it
- [ ] Check recent icons appear after selection
- [ ] Verify selected icon name shows

#### 3. JSON Editor
- [ ] Edit the JSON in Monaco editor
- [ ] Try invalid JSON (should show red error)
- [ ] Click "Format" button (should pretty-print)
- [ ] Click "Minify" button (should compact)
- [ ] Click expand button (should increase height)
- [ ] Verify parsed value shows below

#### 4. Cron Builder
- [ ] Try "Every minute" preset
- [ ] Switch to "Hourly" tab, set minute to 30
- [ ] Switch to "Daily" tab, set time to 09:00
- [ ] Switch to "Weekly" tab, select Monday + Friday
- [ ] Switch to "Custom" tab, enter custom cron
- [ ] Verify expression updates in footer

### ✅ Node Config Dialogs Tab

#### 1. Test Chart Node Dialog
- [ ] Click "Temperature Chart" button
- [ ] Dialog should open with gradient header
- [ ] Check "Configuration" tab shows properties
- [ ] Try changing "Chart Type" dropdown
- [ ] Try changing "Max Data Points" number
- [ ] Check "Info" tab shows node information
- [ ] Click "Cancel" (should close)
- [ ] Reopen and make changes
- [ ] Click "Save Changes" (should show success toast)

#### 2. Test Gauge Node Dialog
- [ ] Click "CPU Gauge" button
- [ ] Verify properties render (min, max, units)
- [ ] Change min/max values
- [ ] Test validation (try max < min)
- [ ] Save and verify

#### 3. Test Button Node Dialog
- [ ] Click "Trigger Button" button
- [ ] Verify color pickers work
- [ ] Test icon picker
- [ ] Save configuration

## 🐛 Expected Issues & Solutions

### Issue: API 404 Errors in Console

**Problem**: Frontend makes API calls but backend might not have full implementation

**Solution**: This is normal for Phase 1. The dialogs will show loading states. We're testing the UI components, not the full API integration.

### Issue: "Node type not found"

**Problem**: Some node types might not return full schema

**Solution**: The backend API at `/api/v1/node-types/:type` should return the full NodeInfo with properties array. Check backend console for errors.

### Issue: Validation doesn't work

**Problem**: Server-side validation endpoint doesn't exist yet

**Solution**: Client-side validation should still work (required fields, types, min/max). Server validation is optional.

## 🔍 What to Look For

### Good Signs ✅
1. All components render without console errors
2. Dropdowns, switches, inputs work smoothly
3. Color picker shows colors correctly
4. Icon picker shows Lucide icons
5. JSON editor has syntax highlighting
6. Cron builder tabs switch correctly
7. Node config dialogs open/close smoothly
8. Properties render dynamically based on node type

### Bad Signs ❌
1. Console errors about missing components
2. Blank screens or infinite loading
3. Buttons don't respond to clicks
4. Validation errors don't show
5. API calls fail with CORS errors

## 📊 Backend API Requirements

For full testing, the backend should implement:

### 1. Get All Node Types
```
GET /api/v1/node-types
Response: {
  "node_types": [
    {
      "type": "dashboard-chart",
      "name": "Dashboard Chart",
      "category": "dashboard",
      "description": "Display data as charts",
      "icon": "chart-line",
      "color": "#8b5cf6",
      "properties": [
        {
          "name": "id",
          "label": "Widget ID",
          "type": "string",
          "required": true
        },
        ...
      ]
    }
  ]
}
```

### 2. Get Specific Node Type
```
GET /api/v1/node-types/dashboard-chart
Response: {
  "type": "dashboard-chart",
  "name": "Dashboard Chart",
  "properties": [...],
  ...
}
```

### 3. Update Node Config (Optional for Phase 1)
```
PUT /api/v1/flows/:flowId/nodes/:nodeId
Body: {
  "config": {
    "label": "Temperature",
    "chartType": "line"
  }
}
```

## 📝 Test Results Template

Copy this and fill in your results:

```
## Test Results - [Date/Time]

### Reusable Components
- [ ] Color Picker: PASS / FAIL - [Notes]
- [ ] Icon Picker: PASS / FAIL - [Notes]
- [ ] JSON Editor: PASS / FAIL - [Notes]
- [ ] Cron Builder: PASS / FAIL - [Notes]

### Node Config Dialogs
- [ ] Chart Dialog: PASS / FAIL - [Notes]
- [ ] Gauge Dialog: PASS / FAIL - [Notes]
- [ ] Button Dialog: PASS / FAIL - [Notes]

### API Integration
- [ ] GET /api/v1/node-types: PASS / FAIL
- [ ] GET /api/v1/node-types/:type: PASS / FAIL
- [ ] Node schema loads correctly: PASS / FAIL

### Issues Found
1. [Issue description]
2. [Issue description]

### Screenshots
[Attach screenshots if possible]
```

## 🎯 Next Steps After Testing

Once Phase 1 testing is complete, we'll proceed to:

### Phase 2: Dashboard Widget Editors (Starting Now)

Creating specialized editors for:
1. ChartEditor - Series management, axis configuration
2. GaugeEditor - Sector editor with colors
3. FormBuilderEditor - Drag-to-reorder fields
4. TableEditor - Column editor
5. And 10 more widget editors...

These will use the reusable components we just tested!

---

**Test Page URL**: http://localhost:3000/test
**Backend Health**: http://localhost:8080/api/v1/health
**Frontend HMR**: Enabled (changes reload automatically)
