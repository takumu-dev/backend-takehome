-- Remove seed data in reverse order due to foreign key constraints
DELETE FROM comments WHERE post_id IN (1, 2, 3, 4, 5);
DELETE FROM posts WHERE author_id IN (1, 2, 3);
DELETE FROM users WHERE email IN ('john@example.com', 'jane@example.com', 'bob@example.com');