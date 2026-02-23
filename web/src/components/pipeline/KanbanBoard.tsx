import { useState } from 'react';
import {
  DndContext,
  DragOverlay,
  closestCorners,
  PointerSensor,
  useSensor,
  useSensors,
  type DragStartEvent,
  type DragEndEvent,
} from '@dnd-kit/core';
import KanbanColumn from './KanbanColumn';
import ApplicationCard from './ApplicationCard';
import type { PipelineApplication } from '../../api/client';

const STAGES = ['applied', 'screening', 'interview', 'offer', 'hired', 'rejected'] as const;

const STAGE_COLORS: Record<string, string> = {
  applied: 'border-t-blue-400',
  screening: 'border-t-yellow-400',
  interview: 'border-t-purple-400',
  offer: 'border-t-green-400',
  hired: 'border-t-emerald-600',
  rejected: 'border-t-red-400',
};

interface Props {
  stages: Record<string, PipelineApplication[]>;
  onMoveStage: (appId: string, newStage: string) => void;
  onCardClick: (app: PipelineApplication) => void;
}

export default function KanbanBoard({ stages, onMoveStage, onCardClick }: Props) {
  const [activeApp, setActiveApp] = useState<PipelineApplication | null>(null);

  const sensors = useSensors(
    useSensor(PointerSensor, { activationConstraint: { distance: 8 } })
  );

  const findApp = (id: string): PipelineApplication | undefined => {
    for (const stage of STAGES) {
      const found = (stages[stage] || []).find(a => a.id === id);
      if (found) return found;
    }
    return undefined;
  };

  const handleDragStart = (event: DragStartEvent) => {
    const app = findApp(event.active.id as string);
    setActiveApp(app || null);
  };

  const handleDragEnd = (event: DragEndEvent) => {
    setActiveApp(null);
    const { active, over } = event;
    if (!over) return;

    const appId = active.id as string;
    const targetStage = over.id as string;

    // Only act if dropped on a valid stage column
    if (STAGES.includes(targetStage as typeof STAGES[number])) {
      const app = findApp(appId);
      if (app && app.stage !== targetStage) {
        onMoveStage(appId, targetStage);
      }
    }
  };

  return (
    <DndContext
      sensors={sensors}
      collisionDetection={closestCorners}
      onDragStart={handleDragStart}
      onDragEnd={handleDragEnd}
    >
      <div className="flex gap-4 overflow-x-auto pb-4">
        {STAGES.map(stage => (
          <KanbanColumn
            key={stage}
            stage={stage}
            color={STAGE_COLORS[stage]}
            applications={stages[stage] || []}
            onCardClick={onCardClick}
          />
        ))}
      </div>
      <DragOverlay>
        {activeApp ? <ApplicationCard app={activeApp} onClick={() => {}} isDragging /> : null}
      </DragOverlay>
    </DndContext>
  );
}
