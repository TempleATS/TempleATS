import { useDraggable } from '@dnd-kit/core';
import { useNavigate } from 'react-router-dom';
import type { PipelineApplication } from '../../api/client';

function pgText(val: unknown): string | null {
  if (val === null || val === undefined) return null;
  if (typeof val === 'string') return val || null;
  if (typeof val === 'object' && val !== null && 'Valid' in val) {
    const t = val as { String: string; Valid: boolean };
    return t.Valid ? t.String : null;
  }
  return null;
}

interface Props {
  app: PipelineApplication;
  onClick: () => void;
  isDragging?: boolean;
}

export default function ApplicationCard({ app, isDragging }: Props) {
  const navigate = useNavigate();
  const { attributes, listeners, setNodeRef, transform } = useDraggable({ id: app.id });

  const style = transform
    ? { transform: `translate(${transform.x}px, ${transform.y}px)` }
    : undefined;

  const company = pgText(app.candidate_company);

  const handleClick = () => {
    navigate(`/applications/${app.id}`);
  };

  return (
    <div
      ref={setNodeRef}
      style={style}
      {...listeners}
      {...attributes}
      onClick={handleClick}
      className={`bg-white rounded border p-3 cursor-pointer hover:shadow-sm transition-shadow ${
        isDragging ? 'shadow-lg opacity-90' : ''
      }`}
    >
      <p className="text-sm font-medium text-gray-900">{app.candidate_name}</p>
      <p className="text-xs text-gray-500 mt-0.5">{company || app.candidate_email}</p>
      {pgText(app.rejection_reason) && (
        <p className="text-xs text-red-500 mt-1">{pgText(app.rejection_reason)}</p>
      )}
    </div>
  );
}
