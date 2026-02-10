import { useState, useEffect } from 'react'
import { BrowserRouter, Routes, Route, Navigate, useNavigate } from 'react-router-dom'
import { Toaster } from 'sonner'
import ErrorBoundary from './components/ErrorBoundary'
import Layout from './components/Layout'
import Dashboard from './pages/Dashboard'
import Workflows from './pages/Workflows'
import EditorFull from './pages/EditorFull'
import ExecutionsFull from './pages/ExecutionsFull'
import SettingsFull from './pages/SettingsFull'
import ModuleManager from './pages/ModuleManager'
import TestComponents from './pages/TestComponents'
import { SetupWizard } from './components/SetupWizard'
import type { SetupConfig } from './components/SetupWizard/types'

const SETUP_COMPLETE_KEY = 'edgeflow_setup_complete'

function FirstRunWizard() {
  const [showWizard, setShowWizard] = useState(false)
  const navigate = useNavigate()

  useEffect(() => {
    const setupDone = localStorage.getItem(SETUP_COMPLETE_KEY)
    if (!setupDone) {
      setShowWizard(true)
    }
  }, [])

  const handleComplete = (config: SetupConfig) => {
    localStorage.setItem(SETUP_COMPLETE_KEY, JSON.stringify({
      completedAt: new Date().toISOString(),
      board: config.board,
    }))
    setShowWizard(false)
  }

  const handleGoToEditor = () => {
    setShowWizard(false)
    navigate('/editor')
  }

  const handleClose = () => {
    localStorage.setItem(SETUP_COMPLETE_KEY, JSON.stringify({
      completedAt: new Date().toISOString(),
      skipped: true,
    }))
    setShowWizard(false)
  }

  if (!showWizard) return null

  return (
    <SetupWizard
      isOpen={showWizard}
      onClose={handleClose}
      onComplete={handleComplete}
      onGoToEditor={handleGoToEditor}
    />
  )
}

function App() {
  return (
    <ErrorBoundary>
      <Toaster position="top-right" richColors closeButton />
      <BrowserRouter future={{ v7_relativeSplatPath: true, v7_startTransition: true }}>
        <FirstRunWizard />
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
