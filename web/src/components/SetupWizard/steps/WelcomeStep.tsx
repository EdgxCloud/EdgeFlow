/**
 * Welcome Step
 *
 * Introduction screen for the setup wizard
 */

import { Cpu, Workflow, Shield, Zap } from 'lucide-react'

export function WelcomeStep() {
  return (
    <div className="space-y-8">
      {/* Hero Section */}
      <div className="text-center space-y-4">
        <div className="w-20 h-20 mx-auto bg-gradient-to-br from-blue-500 to-purple-600 rounded-2xl flex items-center justify-center shadow-lg">
          <Workflow className="w-10 h-10 text-white" />
        </div>
        <h1 className="text-3xl font-bold text-gray-900 dark:text-white">
          Welcome to EdgeFlow
        </h1>
        <p className="text-lg text-muted-foreground max-w-md mx-auto">
          Let's configure your IoT edge device for optimal performance. This
          wizard will guide you through the setup process.
        </p>
      </div>

      {/* Features Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 gap-4 max-w-2xl mx-auto">
        <div className="flex items-start gap-3 p-4 bg-blue-50 dark:bg-blue-950/30 rounded-xl border border-blue-100 dark:border-blue-900">
          <div className="w-10 h-10 bg-blue-500 rounded-lg flex items-center justify-center flex-shrink-0">
            <Cpu className="w-5 h-5 text-white" />
          </div>
          <div>
            <h3 className="font-semibold text-blue-900 dark:text-blue-100">
              Board Detection
            </h3>
            <p className="text-sm text-blue-700 dark:text-blue-300">
              Auto-detect and configure for your specific hardware
            </p>
          </div>
        </div>

        <div className="flex items-start gap-3 p-4 bg-green-50 dark:bg-green-950/30 rounded-xl border border-green-100 dark:border-green-900">
          <div className="w-10 h-10 bg-green-500 rounded-lg flex items-center justify-center flex-shrink-0">
            <Shield className="w-5 h-5 text-white" />
          </div>
          <div>
            <h3 className="font-semibold text-green-900 dark:text-green-100">
              Secure Setup
            </h3>
            <p className="text-sm text-green-700 dark:text-green-300">
              Configure TLS, authentication, and permissions
            </p>
          </div>
        </div>

        <div className="flex items-start gap-3 p-4 bg-purple-50 dark:bg-purple-950/30 rounded-xl border border-purple-100 dark:border-purple-900">
          <div className="w-10 h-10 bg-purple-500 rounded-lg flex items-center justify-center flex-shrink-0">
            <Workflow className="w-5 h-5 text-white" />
          </div>
          <div>
            <h3 className="font-semibold text-purple-900 dark:text-purple-100">
              MQTT Ready
            </h3>
            <p className="text-sm text-purple-700 dark:text-purple-300">
              Built-in broker or connect to external services
            </p>
          </div>
        </div>

        <div className="flex items-start gap-3 p-4 bg-amber-50 dark:bg-amber-950/30 rounded-xl border border-amber-100 dark:border-amber-900">
          <div className="w-10 h-10 bg-amber-500 rounded-lg flex items-center justify-center flex-shrink-0">
            <Zap className="w-5 h-5 text-white" />
          </div>
          <div>
            <h3 className="font-semibold text-amber-900 dark:text-amber-100">
              GPIO Control
            </h3>
            <p className="text-sm text-amber-700 dark:text-amber-300">
              Full access to GPIO, I2C, SPI, and more
            </p>
          </div>
        </div>
      </div>

      {/* What to expect */}
      <div className="bg-gray-50 dark:bg-gray-800/50 rounded-xl p-6 max-w-2xl mx-auto">
        <h3 className="font-semibold mb-3 text-gray-900 dark:text-white">
          What we'll configure:
        </h3>
        <ol className="space-y-2 text-sm text-muted-foreground">
          <li className="flex items-center gap-2">
            <span className="w-6 h-6 bg-primary text-primary-foreground rounded-full flex items-center justify-center text-xs font-bold">
              1
            </span>
            <span>Select your IoT board model</span>
          </li>
          <li className="flex items-center gap-2">
            <span className="w-6 h-6 bg-primary text-primary-foreground rounded-full flex items-center justify-center text-xs font-bold">
              2
            </span>
            <span>Configure network settings (WiFi/Ethernet)</span>
          </li>
          <li className="flex items-center gap-2">
            <span className="w-6 h-6 bg-primary text-primary-foreground rounded-full flex items-center justify-center text-xs font-bold">
              3
            </span>
            <span>Set up MQTT broker connection</span>
          </li>
          <li className="flex items-center gap-2">
            <span className="w-6 h-6 bg-primary text-primary-foreground rounded-full flex items-center justify-center text-xs font-bold">
              4
            </span>
            <span>Configure GPIO and hardware permissions</span>
          </li>
          <li className="flex items-center gap-2">
            <span className="w-6 h-6 bg-primary text-primary-foreground rounded-full flex items-center justify-center text-xs font-bold">
              5
            </span>
            <span>Review and complete setup</span>
          </li>
        </ol>
      </div>
    </div>
  )
}
