-- name: UpsertCalendarConnection :one
INSERT INTO user_calendar_connections (user_id, provider, access_token, refresh_token, token_expiry, calendar_email)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (user_id, provider) DO UPDATE
SET access_token = EXCLUDED.access_token, refresh_token = EXCLUDED.refresh_token,
    token_expiry = EXCLUDED.token_expiry, calendar_email = EXCLUDED.calendar_email, updated_at = now()
RETURNING *;

-- name: GetCalendarConnection :one
SELECT * FROM user_calendar_connections WHERE user_id = $1 AND provider = $2;

-- name: GetCalendarConnectionByUser :one
SELECT * FROM user_calendar_connections WHERE user_id = $1 LIMIT 1;

-- name: DeleteCalendarConnection :exec
DELETE FROM user_calendar_connections WHERE user_id = $1;

-- name: CreateInterviewSchedule :one
INSERT INTO interview_schedules (application_id, token, duration_minutes, location, meeting_url, notes, created_by)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetScheduleByToken :one
SELECT s.*, j.title AS job_title, c.name AS candidate_name, c.email AS candidate_email, o.name AS org_name
FROM interview_schedules s
JOIN applications a ON s.application_id = a.id
JOIN jobs j ON a.job_id = j.id
JOIN candidates c ON a.candidate_id = c.id
JOIN jobs j2 ON a.job_id = j2.id
JOIN organizations o ON j2.organization_id = o.id
WHERE s.token = $1;

-- name: ListSchedulesByApplication :many
SELECT s.*, u.name AS created_by_name
FROM interview_schedules s
JOIN users u ON s.created_by = u.id
WHERE s.application_id = $1
ORDER BY s.created_at DESC;

-- name: UpdateScheduleStatus :exec
UPDATE interview_schedules SET status = $2, confirmed_at = CASE WHEN $2 = 'confirmed' THEN now() ELSE confirmed_at END
WHERE id = $1;

-- name: CreateInterviewSlot :one
INSERT INTO interview_slots (schedule_id, start_time, end_time)
VALUES ($1, $2, $3) RETURNING *;

-- name: ListSlotsBySchedule :many
SELECT * FROM interview_slots WHERE schedule_id = $1 ORDER BY start_time;

-- name: SelectSlot :exec
UPDATE interview_slots SET selected = (id = $2) WHERE schedule_id = $1;

-- name: AddScheduleInterviewer :exec
INSERT INTO interview_schedule_interviewers (schedule_id, user_id) VALUES ($1, $2);

-- name: ListScheduleInterviewers :many
SELECT si.schedule_id, si.user_id, si.calendar_event_id, u.name, u.email
FROM interview_schedule_interviewers si
JOIN users u ON si.user_id = u.id
WHERE si.schedule_id = $1;

-- name: SetCalendarEventID :exec
UPDATE interview_schedule_interviewers SET calendar_event_id = $3 WHERE schedule_id = $1 AND user_id = $2;
