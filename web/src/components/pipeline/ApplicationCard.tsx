import { useDraggable } from '@dnd-kit/core';
import type { PipelineApplication } from '../../api/client';

function pgText(val: { String: string; Valid: boolean } | null): string | null {
  return val?.Valid ? val.String : null;
}

interface Props {
  app: PipelineApplication;
  onClick: () => void;
  isDragging?: boolean;
}

export default function ApplicationCard({ app, onClick, isDragging }: Props) {
  const { attributes, listeners, setNodeRef, transform } = useDraggable({ id: app.id });

  const style = transform
    ? { transform: `translate(${transform.x}px, ${transform.y}px)` }
    : undefined;

  return (
    <div
      ref={setNodeRef}
      style={style}
      {...listeners}
      {...attributes}
      onClick={onClick}
      className={`bg-white rounded border p-3 cursor-pointer hover:shadow-sm transition-shadow ${
        isDragging ? 'shadow-lg opacity-90' : ''
      }`}
    >
      <p className="text-sm font-medium text-gray-900">{app.candidate_name}</p>
      <p className="text-xs text-gray-500 mt-0.5">{app.candidate_email}</p>
      {pgText(app.rejection_reason) && (
        <p className="text-xs text-red-500 mt-1">{pgText(app.rejection_reason)}</p>
      )}
    </div>
  );
}
