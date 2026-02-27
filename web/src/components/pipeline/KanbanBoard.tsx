import { useState } from 'react';
import {
  DndContext,
  DragOverlay,
  pointerWithin,
  PointerSensor,
  useSensor,
  useSensors,
  useDroppable,
  type DragStartEvent,
  type DragEndEvent,
} from '@dnd-kit/core';
import KanbanColumn from './KanbanColumn';
import ApplicationCard from './ApplicationCard';
import type { PipelineApplication } from '../../api/client';

const STAGES = ['applied', 'hr_screen', 'hm_review', 'first_interview', 'final_interview', 'offer'] as const;

const STAGE_COLORS: Record<string, string> = {
  applied: 'border-t-blue-400',
  hr_screen: 'border-t-cyan-400',
  hm_review: 'border-t-yellow-400',
  first_interview: 'border-t-purple-400',
  final_interview: 'border-t-indigo-400',
  offer: 'border-t-green-400',
};

export const STAGE_LABELS: Record<string, string> = {
  applied: 'Applied',
  hr_screen: 'HR Screen',
  hm_review: 'HM Review',
  first_interview: '1st Interview',
  final_interview: 'Final Interview',
  offer: 'Offer',
  rejected: 'Rejected',
};

function DroppableRejectBar({ count, onClick, isDragging }: { count: number; onClick: () => void; isDragging: boolean }) {
  const { setNodeRef, isOver } = useDroppable({ id: 'rejected' });
  return (
    <div ref={setNodeRef} className={`transition-all ${isDragging ? 'pt-2' : ''}`}>
      <button
        onClick={onClick}
        className={`w-full flex items-center justify-between px-4 border rounded-lg transition-all ${
          isDragging
            ? isOver
              ? 'py-8 bg-red-200 border-red-400 ring-2 ring-red-400 border-dashed'
              : 'py-8 bg-red-50 border-red-300 border-dashed'
            : 'py-3 mt-4 bg-red-50 border-red-200 hover:bg-red-100'
        }`}
      >
        <span className={`font-medium text-red-700 ${isDragging ? 'text-base' : 'text-sm'}`}>
          {isDragging ? (isOver ? 'Drop to reject' : 'Drag here to reject') : 'Rejected'}
        </span>
        <span className="text-sm bg-red-200 text-red-800 px-2.5 py-0.5 rounded-full font-medium">
          {count}
        </span>
      </button>
    </div>
  );
}

interface Props {
  stages: Record<string, PipelineApplication[]>;
  onMoveStage: (appId: string, newStage: string) => void;
  onCardClick: (app: PipelineApplication) => void;
  onRejectClick: () => void;
}

export default function KanbanBoard({ stages, onMoveStage, onCardClick, onRejectClick }: Props) {
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

    if (STAGES.includes(targetStage as typeof STAGES[number]) || targetStage === 'rejected') {
      const app = findApp(appId);
      if (app && app.stage !== targetStage) {
        onMoveStage(appId, targetStage);
      }
    }
  };

  const rejectedCount = (stages.rejected || []).length;

  return (
    <DndContext
      sensors={sensors}
      collisionDetection={pointerWithin}
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
      <DroppableRejectBar count={rejectedCount} onClick={onRejectClick} isDragging={!!activeApp} />
      <DragOverlay>
        {activeApp ? <ApplicationCard app={activeApp} onClick={() => {}} isDragging /> : null}
      </DragOverlay>
    </DndContext>
  );
}
