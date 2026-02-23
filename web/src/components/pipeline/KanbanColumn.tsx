import { useDroppable } from '@dnd-kit/core';
import ApplicationCard from './ApplicationCard';
import type { PipelineApplication } from '../../api/client';

interface Props {
  stage: string;
  color: string;
  applications: PipelineApplication[];
  onCardClick: (app: PipelineApplication) => void;
}

export default function KanbanColumn({ stage, color, applications, onCardClick }: Props) {
  const { setNodeRef, isOver } = useDroppable({ id: stage });

  return (
    <div
      ref={setNodeRef}
      className={`flex-shrink-0 w-64 bg-gray-50 rounded-lg border-t-4 ${color} ${
        isOver ? 'ring-2 ring-blue-300 bg-blue-50' : ''
      }`}
    >
      <div className="p-3 border-b bg-white rounded-t-lg">
        <div className="flex items-center justify-between">
          <h3 className="text-sm font-semibold text-gray-700 capitalize">{stage}</h3>
          <span className="text-xs bg-gray-200 text-gray-600 px-2 py-0.5 rounded-full">
            {applications.length}
          </span>
        </div>
      </div>
      <div className="p-2 space-y-2 min-h-[200px]">
        {applications.map(app => (
          <ApplicationCard key={app.id} app={app} onClick={() => onCardClick(app)} />
        ))}
      </div>
    </div>
  );
}
