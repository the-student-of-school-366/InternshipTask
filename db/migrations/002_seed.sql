
INSERT INTO teams (team_name) VALUES
                                  ('backend'),
                                  ('frontend')
    ON CONFLICT (team_name) DO NOTHING;

INSERT INTO users (user_id, username, team_name, is_active) VALUES
                                                                ('u1', 'Alice',   'backend',  true),
                                                                ('u2', 'Bob',     'backend',  true),
                                                                ('u3', 'Charlie', 'backend',  false),
                                                                ('u4', 'Dave',    'frontend', true),
                                                                ('u5', 'Eve',     'frontend', true),
                                                                ('u6', 'Frank',   'frontend', false)
    ON CONFLICT (user_id) DO NOTHING;

INSERT INTO pull_requests (
    pull_request_id,
    pull_request_name,
    author_id,
    status,
    assigned_reviewers,
    created_at,
    merged_at
) VALUES (
             'pr-1001',
             'Add search feature',
             'u1',
             'OPEN',
             ARRAY['u2','u4'],
             NOW(),
             NULL
         )
    ON CONFLICT (pull_request_id) DO NOTHING;
INSERT INTO pull_requests (
    pull_request_id,
    pull_request_name,
    author_id,
    status,
    assigned_reviewers,
    created_at,
    merged_at
) VALUES (
             'pr-1002',
             'Refactor payments',
             'u1',
             'MERGED',
             ARRAY['u2','u5'],
             NOW() - INTERVAL '2 days',
             NOW() - INTERVAL '1 day'
         )
    ON CONFLICT (pull_request_id) DO NOTHING;


INSERT INTO pull_requests (
    pull_request_id,
    pull_request_name,
    author_id,
    status,
    assigned_reviewers,
    created_at,
    merged_at
) VALUES (
             'pr-1003',
             'Docs update',
             'u3',
             'OPEN',
             ARRAY[]::TEXT[],
             NOW() - INTERVAL '3 days',
             NULL
         )
    ON CONFLICT (pull_request_id) DO NOTHING;