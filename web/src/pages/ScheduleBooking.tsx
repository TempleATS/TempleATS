import { useState, useEffect } from 'react';
import { useParams } from 'react-router-dom';
import { api, type PublicSchedule, type InterviewSlot } from '../api/client';

function formatDateTime(ts: { Time: string; Valid: boolean } | null): string {
  if (!ts || !ts.Valid) return '';
  const d = new Date(ts.Time);
  return d.toLocaleDateString('en-US', {
    weekday: 'long',
    year: 'numeric',
    month: 'long',
    day: 'numeric',
  }) + ' at ' + d.toLocaleTimeString('en-US', {
    hour: 'numeric',
    minute: '2-digit',
  });
}

function formatTime(ts: { Time: string; Valid: boolean } | null): string {
  if (!ts || !ts.Valid) return '';
  return new Date(ts.Time).toLocaleTimeString('en-US', { hour: 'numeric', minute: '2-digit' });
}

function formatDate(ts: { Time: string; Valid: boolean } | null): string {
  if (!ts || !ts.Valid) return '';
  return new Date(ts.Time).toLocaleDateString('en-US', {
    weekday: 'short',
    month: 'short',
    day: 'numeric',
  });
}

function pgText(val: unknown): string | null {
  if (val === null || val === undefined) return null;
  if (typeof val === 'string') return val || null;
  if (typeof val === 'object' && val !== null && 'String' in val) {
    const t = val as { String: string; Valid: boolean };
    return t.Valid ? t.String : null;
  }
  return null;
}

export default function ScheduleBooking() {
  const { token } = useParams<{ token: string }>();
  const [schedule, setSchedule] = useState<PublicSchedule | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [selectedSlot, setSelectedSlot] = useState<string | null>(null);
  const [confirming, setConfirming] = useState(false);
  const [confirmed, setConfirmed] = useState(false);

  useEffect(() => {
    if (!token) return;
    api.publicSchedule.get(token)
      .then(data => {
        setSchedule(data);
        if (data.status === 'confirmed') {
          setConfirmed(true);
          const sel = data.slots.find((s: InterviewSlot) => s.selected);
          if (sel) setSelectedSlot(sel.id);
        }
      })
      .catch(() => setError('Schedule not found or has expired.'))
      .finally(() => setLoading(false));
  }, [token]);

  const handleConfirm = async () => {
    if (!token || !selectedSlot) return;
    setConfirming(true);
    setError('');
    try {
      await api.publicSchedule.confirm(token, selectedSlot);
      setConfirmed(true);
    } catch {
      setError('Failed to confirm. Please try again.');
    } finally {
      setConfirming(false);
    }
  };

  if (loading) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <p className="text-gray-500">Loading...</p>
      </div>
    );
  }

  if (error && !schedule) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="bg-white rounded-lg shadow-lg p-8 max-w-md text-center">
          <h1 className="text-xl font-bold text-gray-900 mb-2">Oops!</h1>
          <p className="text-gray-600">{error}</p>
        </div>
      </div>
    );
  }

  if (!schedule) return null;

  const location = pgText(schedule.location);
  const meetingUrl = pgText(schedule.meeting_url);
  const notes = pgText(schedule.notes);

  // Group slots by date
  const slotsByDate: Record<string, InterviewSlot[]> = {};
  for (const slot of schedule.slots) {
    const dateKey = formatDate(slot.start_time);
    if (!slotsByDate[dateKey]) slotsByDate[dateKey] = [];
    slotsByDate[dateKey].push(slot);
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-indigo-50 to-white">
      <div className="max-w-lg mx-auto pt-12 px-4 pb-20">
        {/* Header */}
        <div className="text-center mb-8">
          <h1 className="text-2xl font-bold text-gray-900">{schedule.org_name}</h1>
          <p className="text-gray-600 mt-1">Interview Scheduling</p>
        </div>

        <div className="bg-white rounded-xl shadow-lg overflow-hidden">
          {/* Schedule Info */}
          <div className="p-6 border-b">
            <h2 className="text-lg font-semibold text-gray-900">
              {schedule.job_title}
            </h2>
            <p className="text-sm text-gray-600 mt-1">
              Hi {schedule.candidate_name}, please select a time for your interview.
            </p>
            <div className="mt-3 flex flex-wrap gap-3 text-xs text-gray-500">
              <span>{schedule.duration_minutes} min</span>
              {location && <span>{location}</span>}
              {meetingUrl && <span>Video call</span>}
            </div>
            {notes && (
              <p className="mt-3 text-sm text-gray-600 bg-gray-50 p-3 rounded">{notes}</p>
            )}
          </div>

          {/* Confirmed State */}
          {confirmed ? (
            <div className="p-6 text-center">
              <div className="w-16 h-16 bg-green-100 rounded-full flex items-center justify-center mx-auto mb-4">
                <svg className="w-8 h-8 text-green-600" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
                </svg>
              </div>
              <h3 className="text-lg font-semibold text-green-800">Interview Confirmed!</h3>
              {selectedSlot && (
                <p className="mt-2 text-gray-700 font-medium">
                  {formatDateTime(schedule.slots.find(s => s.id === selectedSlot)?.start_time || null)}
                </p>
              )}
              <p className="mt-3 text-sm text-gray-500">
                You'll receive a confirmation email with all the details.
              </p>
            </div>
          ) : schedule.status === 'cancelled' ? (
            <div className="p-6 text-center">
              <h3 className="text-lg font-semibold text-gray-600">This schedule has been cancelled.</h3>
            </div>
          ) : (
            <>
              {/* Slot Selection */}
              <div className="p-6">
                {error && (
                  <div className="mb-4 p-3 bg-red-50 text-red-700 rounded text-sm">{error}</div>
                )}

                <p className="text-sm font-medium text-gray-700 mb-3">Choose a time:</p>

                <div className="space-y-4">
                  {Object.entries(slotsByDate).map(([date, slots]) => (
                    <div key={date}>
                      <p className="text-xs font-semibold text-gray-500 uppercase tracking-wider mb-2">{date}</p>
                      <div className="space-y-2">
                        {slots.map(slot => (
                          <button
                            key={slot.id}
                            onClick={() => setSelectedSlot(slot.id)}
                            className={`w-full text-left px-4 py-3 rounded-lg border-2 transition-all ${
                              selectedSlot === slot.id
                                ? 'border-indigo-500 bg-indigo-50 text-indigo-900'
                                : 'border-gray-200 hover:border-indigo-300 hover:bg-gray-50'
                            }`}
                          >
                            <span className="font-medium">
                              {formatTime(slot.start_time)} - {formatTime(slot.end_time)}
                            </span>
                          </button>
                        ))}
                      </div>
                    </div>
                  ))}
                </div>
              </div>

              {/* Confirm Button */}
              <div className="p-6 border-t bg-gray-50">
                <button
                  onClick={handleConfirm}
                  disabled={!selectedSlot || confirming}
                  className="w-full py-3 bg-indigo-600 text-white font-medium rounded-lg hover:bg-indigo-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
                >
                  {confirming ? 'Confirming...' : 'Confirm Interview'}
                </button>
              </div>
            </>
          )}
        </div>
      </div>
    </div>
  );
}
