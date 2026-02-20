import { useState, useEffect, lazy, Suspense } from 'react'
import { BrowserRouter, Routes, Route, Navigate, useNavigate } from 'react-router-dom'
import { Toaster } from 'sonner'
import ErrorBoundary from './components/ErrorBoundary'
import Layout from './components/Layout'
import type { SetupConfig } from './components/SetupWizard/types'

const Dashboard = lazy(() => import('./pages/Dashboard'))
const Workflows = lazy(() => import('./pages/Workflows'))
const EditorFull = lazy(() => import('./pages/EditorFull'))
const ExecutionsFull = lazy(() => import('./pages/ExecutionsFull'))
const SettingsFull = lazy(() => import('./pages/SettingsFull'))
const SaaSSettings = lazy(() => import('./pages/SaaSSettings'))
const ModuleManager = lazy(() => import('./pages/ModuleManager'))
const TestComponents = lazy(() => import('./pages/TestComponents'))
const SetupWizard = lazy(() => import('./components/SetupWizard').then(m => ({ default: m.SetupWizard })))

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
        <Suspense fallback={<div className="flex items-center justify-center h-screen text-muted-foreground">Loading...</div>}>
          <Routes>
            <Route path="/" element={<Layout />}>
              <Route index element={<Navigate to="/dashboard" replace />} />
              <Route path="dashboard" element={<Dashboard />} />
              <Route path="workflows" element={<Workflows />} />
              <Route path="executions" element={<ExecutionsFull />} />
              <Route path="modules" element={<ModuleManager />} />
              <Route path="settings" element={<SettingsFull />} />
              <Route path="settings/saas" element={<SaaSSettings />} />
              <Route path="test" element={<TestComponents />} />
            </Route>
            {/* Editor route outside Layout to remove navbar */}
            <Route path="editor/:id?" element={<EditorFull />} />
          </Routes>
        </Suspense>
      </BrowserRouter>
    </ErrorBoundary>
  )
}

export default App
