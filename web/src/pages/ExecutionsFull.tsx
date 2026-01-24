import { useState } from 'react'
import { PlayCircle, CheckCircle, XCircle, Clock } from 'lucide-react'
import { Badge } from '@/components/ui/badge'

interface Execution {
  id: string
  flowId: string
  flowName: string
  status: 'running' | 'completed' | 'error' | 'pending'
  startTime: string
  endTime?: string
  duration?: number
  nodeCount: number
  completedNodes: number
  errorMessage?: string
}

export default function ExecutionsFull() {
  const [executions] = useState<Execution[]>([
    // Mock data - will be replaced with API calls
    {
      id: '1',
      flowId: 'flow-1',
      flowName: 'Test Flow 1',
      status: 'completed',
      startTime: '2024/01/20 14:30',
      endTime: '2024/01/20 14:31',
      duration: 65,
      nodeCount: 5,
      completedNodes: 5,
    },
    {
      id: '2',
      flowId: 'flow-1',
      flowName: 'Test Flow 1',
      status: 'running',
      startTime: '2024/01/20 15:00',
      nodeCount: 5,
      completedNodes: 3,
    },
    {
      id: '3',
      flowId: 'flow-2',
      flowName: 'Error Flow',
      status: 'error',
      startTime: '2024/01/20 14:00',
      endTime: '2024/01/20 14:02',
      duration: 120,
      nodeCount: 8,
      completedNodes: 4,
      errorMessage: 'Error executing HTTP Request node',
    },
  ])

  const [filter, setFilter] = useState<string>('all')

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'running':
        return <PlayCircle className="w-5 h-5 text-blue-500" />
      case 'completed':
        return <CheckCircle className="w-5 h-5 text-green-500" />
      case 'error':
        return <XCircle className="w-5 h-5 text-red-500" />
      default:
        return <Clock className="w-5 h-5 text-gray-500" />
    }
  }

  const getStatusBadge = (status: string) => {
    const variants: Record<string, any> = {
      running: 'info',
      completed: 'success',
      error: 'danger',
      pending: 'default',
    }
    return variants[status] || 'default'
  }

  const filteredExecutions =
    filter === 'all'
      ? executions
      : executions.filter((e) => e.status === filter)

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold text-gray-900 dark:text-white">
            Executions
          </h1>
          <p className="text-gray-600 dark:text-gray-400 mt-1">
            Flow execution history and status
          </p>
        </div>
      </div>

      {/* Filters */}
      <div className="flex items-center space-x-2">
        {['all', 'running', 'completed', 'error'].map((status) => (
          <button
            key={status}
            onClick={() => setFilter(status)}
            className={`px-4 py-2 rounded-lg text-sm font-medium transition-colors ${
              filter === status
                ? 'bg-blue-600 text-white'
                : 'bg-gray-100 dark:bg-gray-700 text-gray-700 dark:text-gray-300 hover:bg-gray-200 dark:hover:bg-gray-600'
            }`}
          >
            {status === 'all'
              ? 'All'
              : status === 'running'
              ? 'Running'
              : status === 'completed'
              ? 'Completed'
              : 'Error'}
          </button>
        ))}
      </div>

      {/* Executions table */}
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow overflow-hidden">
        <table className="min-w-full divide-y divide-gray-200 dark:divide-gray-700">
          <thead className="bg-gray-50 dark:bg-gray-900">
            <tr>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                ID
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                Flow Name
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                Status
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                Start Time
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                Duration
              </th>
              <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 dark:text-gray-400 uppercase tracking-wider">
                Progress
              </th>
            </tr>
          </thead>
          <tbody className="bg-white dark:bg-gray-800 divide-y divide-gray-200 dark:divide-gray-700">
            {filteredExecutions.map((execution) => (
              <tr
                key={execution.id}
                className="hover:bg-gray-50 dark:hover:bg-gray-700/50 cursor-pointer"
              >
                <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-900 dark:text-white">
                  {execution.id}
                </td>
                <td className="px-6 py-4 whitespace-nowrap">
                  <div className="flex items-center">
                    {getStatusIcon(execution.status)}
                    <span className="ml-2 text-sm font-medium text-gray-900 dark:text-white">
                      {execution.flowName}
                    </span>
                  </div>
                </td>
                <td className="px-6 py-4 whitespace-nowrap">
                  <Badge variant={getStatusBadge(execution.status)}>
                    {execution.status === 'running'
                      ? 'Running'
                      : execution.status === 'completed'
                      ? 'Completed'
                      : execution.status === 'error'
                      ? 'Error'
                      : 'Pending'}
                  </Badge>
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-600 dark:text-gray-400">
                  {execution.startTime}
                </td>
                <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-600 dark:text-gray-400">
                  {execution.duration
                    ? `${execution.duration}s`
                    : '-'}
                </td>
                <td className="px-6 py-4 whitespace-nowrap">
                  <div className="flex items-center">
                    <div className="w-full bg-gray-200 dark:bg-gray-700 rounded-full h-2 mr-3">
                      <div
                        className={`h-2 rounded-full ${
                          execution.status === 'completed'
                            ? 'bg-green-500'
                            : execution.status === 'error'
                            ? 'bg-red-500'
                            : 'bg-blue-500'
                        }`}
                        style={{
                          width: `${
                            (execution.completedNodes / execution.nodeCount) *
                            100
                          }%`,
                        }}
                      ></div>
                    </div>
                    <span className="text-sm text-gray-600 dark:text-gray-400">
                      {execution.completedNodes}/{execution.nodeCount}
                    </span>
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>

        {filteredExecutions.length === 0 && (
          <div className="text-center py-12">
            <Clock className="w-12 h-12 text-gray-400 mx-auto mb-4" />
            <p className="text-gray-600 dark:text-gray-400">
              No executions found
            </p>
          </div>
        )}
      </div>
    </div>
  )
}
