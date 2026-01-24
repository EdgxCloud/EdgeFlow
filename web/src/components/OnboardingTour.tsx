import React, { useState, useEffect } from 'react';
import { Card, CardContent } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { X, ChevronLeft, ChevronRight } from 'lucide-react';

export interface TourStep {
  target: string;
  title: string;
  content: string;
  placement?: 'top' | 'bottom' | 'left' | 'right';
  action?: () => void;
}

const defaultSteps: TourStep[] = [
  {
    target: '[data-tour="node-palette"]',
    title: 'Node Palette',
    content: 'Drag nodes from here to the canvas to build your flow. Nodes are organized by category.',
    placement: 'right',
  },
  {
    target: '[data-tour="canvas"]',
    title: 'Flow Canvas',
    content: 'This is where you build your flows. Drag nodes here and connect them to create logic.',
    placement: 'top',
  },
  {
    target: '[data-tour="deploy-button"]',
    title: 'Deploy Button',
    content: 'Click here to deploy your flows and make them active. You can choose full, modified, or single flow deployment.',
    placement: 'bottom',
  },
  {
    target: '[data-tour="debug-panel"]',
    title: 'Debug Panel',
    content: 'View debug messages and flow execution logs here. Use debug nodes to send messages to this panel.',
    placement: 'left',
  },
  {
    target: '[data-tour="settings"]',
    title: 'Settings',
    content: 'Configure global settings, context storage, and advanced options here.',
    placement: 'bottom',
  },
];

export interface OnboardingTourProps {
  steps?: TourStep[];
  onComplete?: () => void;
  onSkip?: () => void;
}

export function OnboardingTour({
  steps = defaultSteps,
  onComplete,
  onSkip,
}: OnboardingTourProps) {
  const [currentStep, setCurrentStep] = useState(0);
  const [position, setPosition] = useState({ top: 0, left: 0 });
  const [isVisible, setIsVisible] = useState(false);

  const step = steps[currentStep];

  useEffect(() => {
    const hasCompletedTour = localStorage.getItem('onboarding-completed');
    if (!hasCompletedTour) {
      setTimeout(() => setIsVisible(true), 1000);
    }
  }, []);

  useEffect(() => {
    if (!isVisible || !step) return;

    const updatePosition = () => {
      const target = document.querySelector(step.target);
      if (!target) return;

      const rect = target.getBoundingClientRect();
      const placement = step.placement || 'bottom';

      let top = 0;
      let left = 0;

      switch (placement) {
        case 'top':
          top = rect.top - 200;
          left = rect.left + rect.width / 2 - 150;
          break;
        case 'bottom':
          top = rect.bottom + 10;
          left = rect.left + rect.width / 2 - 150;
          break;
        case 'left':
          top = rect.top + rect.height / 2 - 100;
          left = rect.left - 320;
          break;
        case 'right':
          top = rect.top + rect.height / 2 - 100;
          left = rect.right + 10;
          break;
      }

      setPosition({ top, left });

      target.scrollIntoView({ behavior: 'smooth', block: 'center' });
      target.classList.add('tour-highlight');
    };

    updatePosition();
    window.addEventListener('resize', updatePosition);

    return () => {
      const target = document.querySelector(step.target);
      if (target) {
        target.classList.remove('tour-highlight');
      }
      window.removeEventListener('resize', updatePosition);
    };
  }, [currentStep, step, isVisible]);

  const handleNext = () => {
    if (step.action) {
      step.action();
    }

    if (currentStep < steps.length - 1) {
      setCurrentStep(currentStep + 1);
    } else {
      handleComplete();
    }
  };

  const handlePrevious = () => {
    if (currentStep > 0) {
      setCurrentStep(currentStep - 1);
    }
  };

  const handleSkip = () => {
    setIsVisible(false);
    localStorage.setItem('onboarding-completed', 'true');
    if (onSkip) {
      onSkip();
    }
  };

  const handleComplete = () => {
    setIsVisible(false);
    localStorage.setItem('onboarding-completed', 'true');
    if (onComplete) {
      onComplete();
    }
  };

  if (!isVisible || !step) {
    return null;
  }

  return (
    <>
      <div className="fixed inset-0 bg-black/50 z-40" onClick={handleSkip} />

      <Card
        className="fixed z-50 w-80 shadow-2xl"
        style={{
          top: `${position.top}px`,
          left: `${position.left}px`,
        }}
      >
        <CardContent className="p-6">
          <div className="flex justify-between items-start mb-4">
            <div className="flex-1">
              <h3 className="text-lg font-semibold">{step.title}</h3>
              <p className="text-sm text-muted-foreground mt-2">{step.content}</p>
            </div>
            <Button
              variant="ghost"
              size="icon"
              className="h-6 w-6"
              onClick={handleSkip}
            >
              <X className="h-4 w-4" />
            </Button>
          </div>

          <div className="flex items-center justify-between mt-4">
            <div className="text-sm text-muted-foreground">
              {currentStep + 1} / {steps.length}
            </div>

            <div className="flex gap-2">
              <Button
                variant="outline"
                size="sm"
                onClick={handlePrevious}
                disabled={currentStep === 0}
              >
                <ChevronLeft className="h-4 w-4 mr-1" />
                Back
              </Button>

              <Button size="sm" onClick={handleNext}>
                {currentStep === steps.length - 1 ? (
                  'Finish'
                ) : (
                  <>
                    Next
                    <ChevronRight className="h-4 w-4 ml-1" />
                  </>
                )}
              </Button>
            </div>
          </div>

          <div className="flex gap-1 mt-4">
            {steps.map((_, index) => (
              <div
                key={index}
                className={`h-1 flex-1 rounded ${
                  index === currentStep
                    ? 'bg-primary'
                    : index < currentStep
                    ? 'bg-primary/50'
                    : 'bg-muted'
                }`}
              />
            ))}
          </div>
        </CardContent>
      </Card>

      <style>{`
        .tour-highlight {
          position: relative;
          z-index: 45;
          box-shadow: 0 0 0 4px rgba(59, 130, 246, 0.5);
          border-radius: 4px;
        }
      `}</style>
    </>
  );
}

export function useOnboardingTour() {
  const [shouldShow, setShouldShow] = useState(false);

  useEffect(() => {
    const hasCompleted = localStorage.getItem('onboarding-completed');
    setShouldShow(!hasCompleted);
  }, []);

  const reset = () => {
    localStorage.removeItem('onboarding-completed');
    setShouldShow(true);
  };

  return {
    shouldShow,
    reset,
  };
}
