BEGIN;

-- 15 candidates
INSERT INTO candidates (id, name, email, phone, organization_id) VALUES
('c0000001-0000-0000-0000-000000000001', 'Marcus Chen',       'marcus.chen@gmail.com',       '+1-415-555-0101', '7cebee99-abdd-4015-978f-0ff5bf421fe9'),
('c0000001-0000-0000-0000-000000000002', 'Priya Sharma',      'priya.sharma@outlook.com',    '+1-510-555-0102', '7cebee99-abdd-4015-978f-0ff5bf421fe9'),
('c0000001-0000-0000-0000-000000000003', 'Jordan Williams',   'jwilliams@proton.me',         '+1-628-555-0103', '7cebee99-abdd-4015-978f-0ff5bf421fe9'),
('c0000001-0000-0000-0000-000000000004', 'Sofia Andersson',   'sofia.a@hey.com',             '+46-70-555-0104', '7cebee99-abdd-4015-978f-0ff5bf421fe9'),
('c0000001-0000-0000-0000-000000000005', 'Tyler Okonkwo',     'tokonkwo@fastmail.com',       '+1-312-555-0105', '7cebee99-abdd-4015-978f-0ff5bf421fe9'),
('c0000001-0000-0000-0000-000000000006', 'Aisha Patel',       'aisha.patel@gmail.com',       '+44-20-5550-0106','7cebee99-abdd-4015-978f-0ff5bf421fe9'),
('c0000001-0000-0000-0000-000000000007', 'Liam Nakamura',     'liam.nak@yahoo.com',          '+1-206-555-0107', '7cebee99-abdd-4015-978f-0ff5bf421fe9'),
('c0000001-0000-0000-0000-000000000008', 'Elena Rodriguez',   'elena.r@lawmail.com',         '+1-305-555-0108', '7cebee99-abdd-4015-978f-0ff5bf421fe9'),
('c0000001-0000-0000-0000-000000000009', 'David Osei',        'david.osei@gmail.com',        '+1-646-555-0109', '7cebee99-abdd-4015-978f-0ff5bf421fe9'),
('c0000001-0000-0000-0000-000000000010', 'Hannah Fischer',    'h.fischer@outlook.de',        '+49-30-555-0110', '7cebee99-abdd-4015-978f-0ff5bf421fe9'),
('c0000001-0000-0000-0000-000000000011', 'Raj Krishnamurthy', 'raj.k@proton.me',             '+1-408-555-0111', '7cebee99-abdd-4015-978f-0ff5bf421fe9'),
('c0000001-0000-0000-0000-000000000012', 'Camille Dupont',    'c.dupont@gmail.com',          '+33-1-5550-0112', '7cebee99-abdd-4015-978f-0ff5bf421fe9'),
('c0000001-0000-0000-0000-000000000013', 'Alex Kim',          'alex.kim@icloud.com',         '+1-213-555-0113', '7cebee99-abdd-4015-978f-0ff5bf421fe9'),
('c0000001-0000-0000-0000-000000000014', 'Nadia Volkov',      'nvolkov@gmail.com',           '+1-718-555-0114', '7cebee99-abdd-4015-978f-0ff5bf421fe9'),
('c0000001-0000-0000-0000-000000000015', 'Ben Agyeman',       'ben.agyeman@outlook.com',     '+1-202-555-0115', '7cebee99-abdd-4015-978f-0ff5bf421fe9');

-- Marcus Chen: strong CC candidate, applied to both CC roles (Staff R + SR)
INSERT INTO applications (candidate_id, job_id, stage, created_at) VALUES
('c0000001-0000-0000-0000-000000000001', '22d86bac-ee52-43a4-81be-ee8c2f359b95', 'final_interview', now() - interval '12 days'),
('c0000001-0000-0000-0000-000000000001', '05b55a85-39d7-402d-b3cf-a80638c5256f', 'first_interview', now() - interval '8 days');

-- Priya Sharma: CC candidate, deep in pipeline for Staff R (offer stage)
INSERT INTO applications (candidate_id, job_id, stage, created_at) VALUES
('c0000001-0000-0000-0000-000000000002', '22d86bac-ee52-43a4-81be-ee8c2f359b95', 'offer', now() - interval '20 days');

-- Jordan Williams: applied to SR, still early
INSERT INTO applications (candidate_id, job_id, stage, created_at) VALUES
('c0000001-0000-0000-0000-000000000003', '05b55a85-39d7-402d-b3cf-a80638c5256f', 'applied', now() - interval '2 days');

-- Sofia Andersson: applied to Staff R, got rejected
INSERT INTO applications (candidate_id, job_id, stage, rejection_reason, rejection_notes, created_at) VALUES
('c0000001-0000-0000-0000-000000000004', '22d86bac-ee52-43a4-81be-ee8c2f359b95', 'rejected', 'not_qualified', 'Strong background but missing required distributed systems experience', now() - interval '15 days');

-- Tyler Okonkwo: CC candidate, applied to both CC roles (hm_review)
INSERT INTO applications (candidate_id, job_id, stage, created_at) VALUES
('c0000001-0000-0000-0000-000000000005', '22d86bac-ee52-43a4-81be-ee8c2f359b95', 'hm_review', now() - interval '5 days'),
('c0000001-0000-0000-0000-000000000005', '05b55a85-39d7-402d-b3cf-a80638c5256f', 'hm_review', now() - interval '4 days');

-- Aisha Patel: SR candidate, in approval stage
INSERT INTO applications (candidate_id, job_id, stage, created_at) VALUES
('c0000001-0000-0000-0000-000000000006', '05b55a85-39d7-402d-b3cf-a80638c5256f', 'approval', now() - interval '18 days');

-- Liam Nakamura: hired for SR
INSERT INTO applications (candidate_id, job_id, stage, created_at) VALUES
('c0000001-0000-0000-0000-000000000007', '05b55a85-39d7-402d-b3cf-a80638c5256f', 'hired', now() - interval '25 days');

-- Elena Rodriguez: legal candidate, final interview for LR 5
INSERT INTO applications (candidate_id, job_id, stage, created_at) VALUES
('c0000001-0000-0000-0000-000000000008', 'e09881f2-bf2c-4318-82b1-2d7164a47e34', 'final_interview', now() - interval '10 days');

-- David Osei: legal candidate, first interview for LR 5
INSERT INTO applications (candidate_id, job_id, stage, created_at) VALUES
('c0000001-0000-0000-0000-000000000009', 'e09881f2-bf2c-4318-82b1-2d7164a47e34', 'first_interview', now() - interval '7 days');

-- Hannah Fischer: legal candidate, just applied to LR 5
INSERT INTO applications (candidate_id, job_id, stage, created_at) VALUES
('c0000001-0000-0000-0000-000000000010', 'e09881f2-bf2c-4318-82b1-2d7164a47e34', 'applied', now() - interval '1 day');

-- Raj Krishnamurthy: legal candidate, rejected from LR 5
INSERT INTO applications (candidate_id, job_id, stage, rejection_reason, rejection_notes, created_at) VALUES
('c0000001-0000-0000-0000-000000000011', 'e09881f2-bf2c-4318-82b1-2d7164a47e34', 'rejected', 'culture_fit', 'Interview panel felt communication style mismatch with team', now() - interval '14 days');

-- Camille Dupont: legal candidate, offer stage for LR 5
INSERT INTO applications (candidate_id, job_id, stage, created_at) VALUES
('c0000001-0000-0000-0000-000000000012', 'e09881f2-bf2c-4318-82b1-2d7164a47e34', 'offer', now() - interval '22 days');

-- Alex Kim: cross-functional, applied to LR 5 AND Staff R
INSERT INTO applications (candidate_id, job_id, stage, created_at) VALUES
('c0000001-0000-0000-0000-000000000013', 'e09881f2-bf2c-4318-82b1-2d7164a47e34', 'hm_review', now() - interval '6 days'),
('c0000001-0000-0000-0000-000000000013', '22d86bac-ee52-43a4-81be-ee8c2f359b95', 'applied', now() - interval '3 days');

-- Nadia Volkov: applied to SR and LR 5
INSERT INTO applications (candidate_id, job_id, stage, created_at) VALUES
('c0000001-0000-0000-0000-000000000014', '05b55a85-39d7-402d-b3cf-a80638c5256f', 'first_interview', now() - interval '9 days'),
('c0000001-0000-0000-0000-000000000014', 'e09881f2-bf2c-4318-82b1-2d7164a47e34', 'applied', now() - interval '4 days');

-- Ben Agyeman: applied to all three roles
INSERT INTO applications (candidate_id, job_id, stage, created_at) VALUES
('c0000001-0000-0000-0000-000000000015', '22d86bac-ee52-43a4-81be-ee8c2f359b95', 'first_interview', now() - interval '11 days'),
('c0000001-0000-0000-0000-000000000015', '05b55a85-39d7-402d-b3cf-a80638c5256f', 'applied', now() - interval '3 days'),
('c0000001-0000-0000-0000-000000000015', 'e09881f2-bf2c-4318-82b1-2d7164a47e34', 'hm_review', now() - interval '6 days');

COMMIT;
