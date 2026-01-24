import React, { useState, useRef, useEffect } from 'react';
import { Card } from '@/components/ui/card';
import { Button } from '@/components/ui/button';
import { ChevronUp, ChevronDown, X } from 'lucide-react';

export interface MobileBottomSheetProps {
  children: React.ReactNode;
  title?: string;
  defaultOpen?: boolean;
  onClose?: () => void;
  snapPoints?: number[];
}

export function MobileBottomSheet({
  children,
  title,
  defaultOpen = false,
  onClose,
  snapPoints = [0.3, 0.6, 0.9],
}: MobileBottomSheetProps) {
  const [isOpen, setIsOpen] = useState(defaultOpen);
  const [snapPoint, setSnapPoint] = useState(snapPoints[0]);
  const [isDragging, setIsDragging] = useState(false);
  const sheetRef = useRef<HTMLDivElement>(null);
  const startY = useRef(0);
  const startHeight = useRef(0);

  const height = `${snapPoint * 100}vh`;

  const handleTouchStart = (e: React.TouchEvent) => {
    setIsDragging(true);
    startY.current = e.touches[0].clientY;
    startHeight.current = snapPoint * window.innerHeight;
  };

  const handleTouchMove = (e: React.TouchEvent) => {
    if (!isDragging) return;

    const currentY = e.touches[0].clientY;
    const diff = startY.current - currentY;
    const newHeight = startHeight.current + diff;
    const newSnapPoint = Math.min(
      Math.max(newHeight / window.innerHeight, 0.1),
      0.95
    );

    setSnapPoint(newSnapPoint);
  };

  const handleTouchEnd = () => {
    setIsDragging(false);

    const closest = snapPoints.reduce((prev, curr) =>
      Math.abs(curr - snapPoint) < Math.abs(prev - snapPoint) ? curr : prev
    );

    setSnapPoint(closest);

    if (closest < 0.2) {
      handleClose();
    }
  };

  const handleClose = () => {
    setIsOpen(false);
    if (onClose) {
      onClose();
    }
  };

  const handleSnapToNext = () => {
    const currentIndex = snapPoints.indexOf(snapPoint);
    const nextIndex = Math.min(currentIndex + 1, snapPoints.length - 1);
    setSnapPoint(snapPoints[nextIndex]);
  };

  const handleSnapToPrev = () => {
    const currentIndex = snapPoints.indexOf(snapPoint);
    const prevIndex = Math.max(currentIndex - 1, 0);
    setSnapPoint(snapPoints[prevIndex]);
  };

  if (!isOpen) {
    return null;
  }

  return (
    <>
      <div
        className="fixed inset-0 bg-black/50 z-40 md:hidden"
        onClick={handleClose}
      />

      <Card
        ref={sheetRef}
        className="fixed bottom-0 left-0 right-0 z-50 rounded-t-2xl md:hidden transition-all"
        style={{
          height,
          touchAction: 'none',
        }}
      >
        <div
          className="h-12 flex items-center justify-between px-4 cursor-grab active:cursor-grabbing"
          onTouchStart={handleTouchStart}
          onTouchMove={handleTouchMove}
          onTouchEnd={handleTouchEnd}
        >
          <div className="flex-1 flex items-center justify-center">
            <div className="w-12 h-1 bg-muted-foreground/30 rounded-full" />
          </div>

          {title && (
            <h3 className="absolute left-1/2 transform -translate-x-1/2 font-semibold">
              {title}
            </h3>
          )}

          <Button
            variant="ghost"
            size="icon"
            onClick={handleClose}
            className="ml-auto"
          >
            <X className="h-4 w-4" />
          </Button>
        </div>

        <div className="h-[calc(100%-3rem)] overflow-y-auto p-4">{children}</div>

        <div className="absolute right-4 bottom-4 flex flex-col gap-2">
          <Button
            size="icon"
            variant="secondary"
            onClick={handleSnapToNext}
            disabled={snapPoints.indexOf(snapPoint) === snapPoints.length - 1}
          >
            <ChevronUp className="h-4 w-4" />
          </Button>
          <Button
            size="icon"
            variant="secondary"
            onClick={handleSnapToPrev}
            disabled={snapPoints.indexOf(snapPoint) === 0}
          >
            <ChevronDown className="h-4 w-4" />
          </Button>
        </div>
      </Card>
    </>
  );
}

export function useMobileBottomSheet() {
  const [isOpen, setIsOpen] = useState(false);
  const [content, setContent] = useState<React.ReactNode>(null);

  const open = (newContent: React.ReactNode) => {
    setContent(newContent);
    setIsOpen(true);
  };

  const close = () => {
    setIsOpen(false);
  };

  return {
    isOpen,
    content,
    open,
    close,
  };
}
