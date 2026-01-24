import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import ErrorBoundary from './components/ErrorBoundary'
import Layout from './components/Layout'
import Dashboard from './pages/Dashboard'
import Workflows from './pages/Workflows'
import EditorFull from './pages/EditorFull'
import ExecutionsFull from './pages/ExecutionsFull'
import SettingsFull from './pages/SettingsFull'
import ModuleManager from './pages/ModuleManager'
import TestComponents from './pages/TestComponents'

function App() {
  return (
    <ErrorBoundary>
      <BrowserRouter future={{ v7_relativeSplatPath: true, v7_startTransition: true }}>
        <Routes>
          <Route path="/" element={<Layout />}>
            <Route index element={<Navigate to="/dashboard" replace />} />
            <Route path="dashboard" element={<Dashboard />} />
            <Route path="workflows" element={<Workflows />} />
            <Route path="executions" element={<ExecutionsFull />} />
            <Route path="modules" element={<ModuleManager />} />
            <Route path="settings" element={<SettingsFull />} />
            <Route path="test" element={<TestComponents />} />
          </Route>
          {/* Editor route outside Layout to remove navbar */}
          <Route path="editor/:id?" element={<EditorFull />} />
        </Routes>
      </BrowserRouter>
    </ErrorBoundary>
  )
}

export default App
