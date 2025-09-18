-- Insert sample users
INSERT INTO users (name, email, password_hash) VALUES
('John Doe', 'john@example.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi'), -- password: password
('Jane Smith', 'jane@example.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi'), -- password: password
('Bob Johnson', 'bob@example.com', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi'); -- password: password

-- Insert sample posts
INSERT INTO posts (title, content, author_id) VALUES
('Welcome to Our Blog Platform', 'This is the first post on our new blog platform. We are excited to share our thoughts and ideas with you!', 1),
('Getting Started with Go', 'Go is an amazing programming language for building web applications. In this post, we will explore the basics of Go programming.', 1),
('Database Design Best Practices', 'When designing databases, it is important to consider normalization, indexing, and performance optimization.', 2),
('Building RESTful APIs', 'REST APIs are the backbone of modern web applications. Here are some best practices for building robust APIs.', 2),
('Introduction to Docker', 'Docker containers make it easy to deploy and scale applications. Learn how to get started with containerization.', 3);

-- Insert sample comments
INSERT INTO comments (post_id, author_name, content) VALUES
(1, 'Alice Cooper', 'Great post! Looking forward to more content.'),
(1, 'Charlie Brown', 'Thanks for sharing this. Very informative.'),
(2, 'Diana Prince', 'Go is indeed a fantastic language. I love its simplicity.'),
(2, 'Eve Adams', 'Could you write more about Go concurrency patterns?'),
(3, 'Frank Miller', 'Database design is crucial for application performance.'),
(4, 'Grace Hopper', 'REST APIs are essential for modern web development.'),
(4, 'Henry Ford', 'What about GraphQL vs REST? Would love to see a comparison.'),
(5, 'Ivy League', 'Docker has revolutionized how we deploy applications.');