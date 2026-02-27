BEGIN;

-- Realistic candidates with companies, linkedin, and resume filenames
INSERT INTO candidates (id, name, email, phone, company, linkedin_url, resume_url, resume_filename, organization_id) VALUES
('cand-0001-0000-0000-000000000001', 'Sarah Chen', 'sarah.chen@gmail.com', '+1-415-555-0101', 'Google', 'https://linkedin.com/in/sarahchen', '/uploads/sarah_chen_resume.pdf', 'Sarah_Chen_Resume.pdf', '7cebee99-abdd-4015-978f-0ff5bf421fe9'),
('cand-0002-0000-0000-000000000002', 'James Rodriguez', 'james.rodriguez@outlook.com', '+1-212-555-0102', 'Meta', 'https://linkedin.com/in/jamesrodriguez', '/uploads/james_rodriguez_cv.pdf', 'James_Rodriguez_CV.pdf', '7cebee99-abdd-4015-978f-0ff5bf421fe9'),
('cand-0003-0000-0000-000000000003', 'Emily Watson', 'emily.watson@protonmail.com', '+1-650-555-0103', 'Stripe', 'https://linkedin.com/in/emilywatson', '/uploads/emily_watson_resume.pdf', 'Emily_Watson_Resume.pdf', '7cebee99-abdd-4015-978f-0ff5bf421fe9'),
('cand-0004-0000-0000-000000000004', 'Raj Patel', 'raj.patel@gmail.com', '+1-408-555-0104', 'Amazon', 'https://linkedin.com/in/rajpatel', '/uploads/raj_patel_resume.pdf', 'Raj_Patel_Resume.pdf', '7cebee99-abdd-4015-978f-0ff5bf421fe9'),
('cand-0005-0000-0000-000000000005', 'Maria Gonzalez', 'maria.gonzalez@yahoo.com', '+1-310-555-0105', 'Netflix', 'https://linkedin.com/in/mariagonzalez', '/uploads/maria_gonzalez_cv.pdf', 'Maria_Gonzalez_CV.pdf', '7cebee99-abdd-4015-978f-0ff5bf421fe9'),
('cand-0006-0000-0000-000000000006', 'Michael Park', 'michael.park@gmail.com', '+1-206-555-0106', 'Microsoft', 'https://linkedin.com/in/michaelpark', '/uploads/michael_park_resume.pdf', 'Michael_Park_Resume.pdf', '7cebee99-abdd-4015-978f-0ff5bf421fe9'),
('cand-0007-0000-0000-000000000007', 'Priya Sharma', 'priya.sharma@gmail.com', '+1-512-555-0107', 'Apple', 'https://linkedin.com/in/priyasharma', '/uploads/priya_sharma_resume.pdf', 'Priya_Sharma_Resume.pdf', '7cebee99-abdd-4015-978f-0ff5bf421fe9'),
('cand-0008-0000-0000-000000000008', 'Daniel Kim', 'daniel.kim@icloud.com', '+1-503-555-0108', 'Databricks', 'https://linkedin.com/in/danielkim', '/uploads/daniel_kim_cv.pdf', 'Daniel_Kim_CV.pdf', '7cebee99-abdd-4015-978f-0ff5bf421fe9'),
('cand-0009-0000-0000-000000000009', 'Lauren Mitchell', 'lauren.mitchell@gmail.com', '+1-617-555-0109', 'Notion', 'https://linkedin.com/in/laurenmitchell', '/uploads/lauren_mitchell_resume.pdf', 'Lauren_Mitchell_Resume.pdf', '7cebee99-abdd-4015-978f-0ff5bf421fe9'),
('cand-0010-0000-0000-000000000010', 'Ahmed Hassan', 'ahmed.hassan@outlook.com', '+1-312-555-0110', 'Airbnb', 'https://linkedin.com/in/ahmedhassan', '/uploads/ahmed_hassan_resume.pdf', 'Ahmed_Hassan_Resume.pdf', '7cebee99-abdd-4015-978f-0ff5bf421fe9'),
('cand-0011-0000-0000-000000000011', 'Jessica Liu', 'jessica.liu@gmail.com', '+1-415-555-0111', 'Figma', 'https://linkedin.com/in/jessicaliu', '/uploads/jessica_liu_resume.pdf', 'Jessica_Liu_Resume.pdf', '7cebee99-abdd-4015-978f-0ff5bf421fe9'),
('cand-0012-0000-0000-000000000012', 'Tom Anderson', 'tom.anderson@gmail.com', '+1-720-555-0112', 'Snowflake', 'https://linkedin.com/in/tomanderson', '/uploads/tom_anderson_cv.pdf', 'Tom_Anderson_CV.pdf', '7cebee99-abdd-4015-978f-0ff5bf421fe9'),
('cand-0013-0000-0000-000000000013', 'Nina Petrov', 'nina.petrov@protonmail.com', '+1-202-555-0113', 'Palantir', 'https://linkedin.com/in/ninapetrov', '/uploads/nina_petrov_resume.pdf', 'Nina_Petrov_Resume.pdf', '7cebee99-abdd-4015-978f-0ff5bf421fe9'),
('cand-0014-0000-0000-000000000014', 'Kevin O''Brien', 'kevin.obrien@gmail.com', '+1-858-555-0114', 'Coinbase', 'https://linkedin.com/in/kevinobrien', '/uploads/kevin_obrien_resume.pdf', 'Kevin_OBrien_Resume.pdf', '7cebee99-abdd-4015-978f-0ff5bf421fe9'),
('cand-0015-0000-0000-000000000015', 'Fatima Al-Rashid', 'fatima.alrashid@gmail.com', '+1-469-555-0115', 'Datadog', 'https://linkedin.com/in/fatimaalrashid', '/uploads/fatima_alrashid_cv.pdf', 'Fatima_AlRashid_CV.pdf', '7cebee99-abdd-4015-978f-0ff5bf421fe9')
ON CONFLICT (email, organization_id) DO NOTHING;

-- Applications to the 3 existing jobs (SR, LR 5, Staff R)
-- Mix of stages for realism
INSERT INTO applications (id, candidate_id, job_id, stage) VALUES
-- SR job applications
('app-0001-0000-0000-000000000001', 'cand-0001-0000-0000-000000000001', '05b55a85-39d7-402d-b3cf-a80638c5256f', 'final_interview'),
('app-0002-0000-0000-000000000002', 'cand-0002-0000-0000-000000000002', '05b55a85-39d7-402d-b3cf-a80638c5256f', 'hr_screen'),
('app-0003-0000-0000-000000000003', 'cand-0003-0000-0000-000000000003', '05b55a85-39d7-402d-b3cf-a80638c5256f', 'applied'),
('app-0004-0000-0000-000000000004', 'cand-0004-0000-0000-000000000004', '05b55a85-39d7-402d-b3cf-a80638c5256f', 'offer'),
('app-0005-0000-0000-000000000005', 'cand-0005-0000-0000-000000000005', '05b55a85-39d7-402d-b3cf-a80638c5256f', 'first_interview'),
-- LR 5 job applications
('app-0006-0000-0000-000000000006', 'cand-0006-0000-0000-000000000006', 'e09881f2-bf2c-4318-82b1-2d7164a47e34', 'hm_review'),
('app-0007-0000-0000-000000000007', 'cand-0007-0000-0000-000000000007', 'e09881f2-bf2c-4318-82b1-2d7164a47e34', 'applied'),
('app-0008-0000-0000-000000000008', 'cand-0008-0000-0000-000000000008', 'e09881f2-bf2c-4318-82b1-2d7164a47e34', 'first_interview'),
('app-0009-0000-0000-000000000009', 'cand-0009-0000-0000-000000000009', 'e09881f2-bf2c-4318-82b1-2d7164a47e34', 'rejected'),
('app-0010-0000-0000-000000000010', 'cand-0010-0000-0000-000000000010', 'e09881f2-bf2c-4318-82b1-2d7164a47e34', 'hr_screen'),
-- Staff R job applications
('app-0011-0000-0000-000000000011', 'cand-0011-0000-0000-000000000011', '22d86bac-ee52-43a4-81be-ee8c2f359b95', 'final_interview'),
('app-0012-0000-0000-000000000012', 'cand-0012-0000-0000-000000000012', '22d86bac-ee52-43a4-81be-ee8c2f359b95', 'applied'),
('app-0013-0000-0000-000000000013', 'cand-0013-0000-0000-000000000013', '22d86bac-ee52-43a4-81be-ee8c2f359b95', 'offer'),
('app-0014-0000-0000-000000000014', 'cand-0014-0000-0000-000000000014', '22d86bac-ee52-43a4-81be-ee8c2f359b95', 'hm_review'),
('app-0015-0000-0000-000000000015', 'cand-0015-0000-0000-000000000015', '22d86bac-ee52-43a4-81be-ee8c2f359b95', 'first_interview'),
-- Cross-applications (some candidates applied to multiple jobs)
('app-0016-0000-0000-000000000016', 'cand-0001-0000-0000-000000000001', 'e09881f2-bf2c-4318-82b1-2d7164a47e34', 'applied'),
('app-0017-0000-0000-000000000017', 'cand-0004-0000-0000-000000000004', '22d86bac-ee52-43a4-81be-ee8c2f359b95', 'hr_screen')
ON CONFLICT DO NOTHING;

COMMIT;
