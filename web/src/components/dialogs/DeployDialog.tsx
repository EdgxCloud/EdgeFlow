/**
 * Deploy Dialog Component
 * Allows users to deploy flows with different modes
 */

import { useState } from 'react'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { Label } from '@/components/ui/label'
import { RadioGroup, RadioGroupItem } from '@/components/ui/radio-group'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Loader2, Rocket, AlertCircle, CheckCircle2 } from 'lucide-react'
import { flowsApi, DeployRequest } from '@/services/flows'
import { toast } from 'sonner'

interface DeployDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  flowIds?: string[]
}

export default function DeployDialog({ open, onOpenChange, flowIds }: DeployDialogProps) {
  const [deployMode, setDeployMode] = useState<'full' | 'modified' | 'flow'>('full')
  const [isDeploying, setIsDeploying] = useState(false)
  const [deployResult, setDeployResult] = useState<{
    success: boolean
    message: string
    errors?: Record<string, string>
  } | null>(null)

  const handleDeploy = async () => {
    setIsDeploying(true)
    setDeployResult(null)

    try {
      const request: DeployRequest = {
        mode: deployMode,
        ...(deployMode === 'flow' && flowIds ? { flow_ids: flowIds } : {}),
      }

      const response = await flowsApi.deploy(request)

      setDeployResult({
        success: response.success,
        message: response.message,
        errors: response.errors,
      })

      if (response.success) {
        toast.success('Deployment successful', {
          description: response.message,
        })
      } else {
        toast.error('Deployment failed', {
          description: response.message,
        })
      }
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Unknown error'
      setDeployResult({
        success: false,
        message: errorMessage,
      })
      toast.error('Deployment failed', {
        description: errorMessage,
      })
    } finally {
      setIsDeploying(false)
    }
  }

  const handleClose = () => {
    onOpenChange(false)
    // Reset state after a short delay to avoid flickering
    setTimeout(() => {
      setDeployMode('full')
      setDeployResult(null)
    }, 200)
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-[500px]">
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Rocket className="h-5 w-5" />
            Deploy Flows
          </DialogTitle>
          <DialogDescription>
            Choose a deployment mode and click deploy to update your flows
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4 py-4">
          {/* Deploy Mode Selection */}
          <div className="space-y-3">
            <Label>Deployment Mode</Label>
            <RadioGroup
              value={deployMode}
              onValueChange={(value) => setDeployMode(value as 'full' | 'modified' | 'flow')}
            >
              <div className="flex items-start space-x-3 p-3 rounded-lg border hover:bg-accent cursor-pointer">
                <RadioGroupItem value="full" id="full" className="mt-1" />
                <div className="flex-1">
                  <Label htmlFor="full" className="cursor-pointer font-medium">
                    Full Deployment
                  </Label>
                  <p className="text-sm text-muted-foreground mt-1">
                    Stop all flows, deploy all changes, and restart. This is the safest option but causes
                    a brief interruption.
                  </p>
                </div>
              </div>

              <div className="flex items-start space-x-3 p-3 rounded-lg border hover:bg-accent cursor-pointer">
                <RadioGroupItem value="modified" id="modified" className="mt-1" />
                <div className="flex-1">
                  <Label htmlFor="modified" className="cursor-pointer font-medium">
                    Modified Flows
                  </Label>
                  <p className="text-sm text-muted-foreground mt-1">
                    Only deploy flows that have been modified since the last deployment. Faster with minimal
                    disruption.
                  </p>
                </div>
              </div>

              <div className="flex items-start space-x-3 p-3 rounded-lg border hover:bg-accent cursor-pointer">
                <RadioGroupItem value="flow" id="flow" className="mt-1" disabled={!flowIds} />
                <div className="flex-1">
                  <Label
                    htmlFor="flow"
                    className={`cursor-pointer font-medium ${!flowIds ? 'opacity-50' : ''}`}
                  >
                    Selected Flows Only
                  </Label>
                  <p className="text-sm text-muted-foreground mt-1">
                    Deploy only the currently selected flow(s). Other flows remain unchanged.
                    {!flowIds && ' (No flows selected)'}
                  </p>
                </div>
              </div>
            </RadioGroup>
          </div>

          {/* Deploy Result */}
          {deployResult && (
            <Alert variant={deployResult.success ? 'default' : 'destructive'}>
              {deployResult.success ? (
                <CheckCircle2 className="h-4 w-4" />
              ) : (
                <AlertCircle className="h-4 w-4" />
              )}
              <AlertDescription>
                <p className="font-medium">{deployResult.message}</p>
                {deployResult.errors && Object.keys(deployResult.errors).length > 0 && (
                  <ul className="mt-2 space-y-1 text-sm">
                    {Object.entries(deployResult.errors).map(([flowId, error]) => (
                      <li key={flowId}>
                        <strong>{flowId}:</strong> {error}
                      </li>
                    ))}
                  </ul>
                )}
              </AlertDescription>
            </Alert>
          )}

          {/* Warning */}
          {deployMode === 'full' && !deployResult && (
            <Alert>
              <AlertCircle className="h-4 w-4" />
              <AlertDescription className="text-sm">
                <strong>Warning:</strong> Full deployment will briefly interrupt all running flows.
              </AlertDescription>
            </Alert>
          )}
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={handleClose} disabled={isDeploying}>
            Cancel
          </Button>
          <Button onClick={handleDeploy} disabled={isDeploying}>
            {isDeploying ? (
              <>
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                Deploying...
              </>
            ) : (
              <>
                <Rocket className="mr-2 h-4 w-4" />
                Deploy
              </>
            )}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
