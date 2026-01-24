import React, { useState } from 'react';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { RadioGroup, RadioGroupItem } from '@/components/ui/radio-group';
import { Label } from '@/components/ui/label';
import { Progress } from '@/components/ui/progress';
import { Alert, AlertDescription } from '@/components/ui/alert';
import { CheckCircle2, XCircle, Loader2 } from 'lucide-react';

export type DeployMode = 'full' | 'modified' | 'flow';

export interface DeployDialogProps {
  open: boolean;
  onClose: () => void;
  onDeploy: (mode: DeployMode) => Promise<void>;
  flows: Array<{ id: string; name: string; modified?: boolean }>;
}

export function DeployDialog({
  open,
  onClose,
  onDeploy,
  flows,
}: DeployDialogProps) {
  const [mode, setMode] = useState<DeployMode>('modified');
  const [selectedFlowId, setSelectedFlowId] = useState<string>('');
  const [isDeploying, setIsDeploying] = useState(false);
  const [progress, setProgress] = useState(0);
  const [result, setResult] = useState<{
    success: boolean;
    message: string;
  } | null>(null);

  const modifiedFlows = flows.filter((f) => f.modified);

  const handleDeploy = async () => {
    setIsDeploying(true);
    setProgress(0);
    setResult(null);

    const progressInterval = setInterval(() => {
      setProgress((prev) => Math.min(prev + 10, 90));
    }, 100);

    try {
      await onDeploy(mode);
      setProgress(100);
      setResult({
        success: true,
        message: 'Deployment completed successfully',
      });
    } catch (error) {
      setProgress(100);
      setResult({
        success: false,
        message: error instanceof Error ? error.message : 'Deployment failed',
      });
    } finally {
      clearInterval(progressInterval);
      setIsDeploying(false);
    }
  };

  const handleClose = () => {
    if (!isDeploying) {
      setResult(null);
      setProgress(0);
      onClose();
    }
  };

  return (
    <Dialog open={open} onOpenChange={handleClose}>
      <DialogContent className="sm:max-w-[500px]">
        <DialogHeader>
          <DialogTitle>Deploy Flows</DialogTitle>
          <DialogDescription>
            Choose how you want to deploy your flows
          </DialogDescription>
        </DialogHeader>

        <div className="space-y-4 py-4">
          <RadioGroup
            value={mode}
            onValueChange={(value) => setMode(value as DeployMode)}
            disabled={isDeploying}
          >
            <div className="flex items-start space-x-2">
              <RadioGroupItem value="full" id="full" />
              <div className="flex-1">
                <Label htmlFor="full" className="font-semibold">
                  Full Deployment
                </Label>
                <p className="text-sm text-muted-foreground">
                  Stop all flows and deploy everything from scratch. All flows: {flows.length}
                </p>
              </div>
            </div>

            <div className="flex items-start space-x-2">
              <RadioGroupItem value="modified" id="modified" />
              <div className="flex-1">
                <Label htmlFor="modified" className="font-semibold">
                  Modified Flows Only
                </Label>
                <p className="text-sm text-muted-foreground">
                  Deploy only flows that have been modified. Modified flows: {modifiedFlows.length}
                </p>
              </div>
            </div>

            <div className="flex items-start space-x-2">
              <RadioGroupItem value="flow" id="flow" />
              <div className="flex-1">
                <Label htmlFor="flow" className="font-semibold">
                  Single Flow
                </Label>
                <p className="text-sm text-muted-foreground">
                  Deploy a specific flow
                </p>
                {mode === 'flow' && (
                  <select
                    className="mt-2 w-full rounded border p-2"
                    value={selectedFlowId}
                    onChange={(e) => setSelectedFlowId(e.target.value)}
                    disabled={isDeploying}
                  >
                    <option value="">Select a flow...</option>
                    {flows.map((flow) => (
                      <option key={flow.id} value={flow.id}>
                        {flow.name} {flow.modified && '(modified)'}
                      </option>
                    ))}
                  </select>
                )}
              </div>
            </div>
          </RadioGroup>

          {isDeploying && (
            <div className="space-y-2">
              <Progress value={progress} />
              <p className="text-sm text-center text-muted-foreground">
                Deploying... {progress}%
              </p>
            </div>
          )}

          {result && (
            <Alert variant={result.success ? 'default' : 'destructive'}>
              <div className="flex items-center gap-2">
                {result.success ? (
                  <CheckCircle2 className="h-4 w-4" />
                ) : (
                  <XCircle className="h-4 w-4" />
                )}
                <AlertDescription>{result.message}</AlertDescription>
              </div>
            </Alert>
          )}
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={handleClose} disabled={isDeploying}>
            {result ? 'Close' : 'Cancel'}
          </Button>
          {!result && (
            <Button
              onClick={handleDeploy}
              disabled={
                isDeploying ||
                (mode === 'flow' && !selectedFlowId) ||
                (mode === 'modified' && modifiedFlows.length === 0)
              }
            >
              {isDeploying && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
              Deploy
            </Button>
          )}
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
