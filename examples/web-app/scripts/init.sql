-- 数据库初始化脚本

-- 创建用户表
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    email VARCHAR(100) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建文章表
CREATE TABLE IF NOT EXISTS articles (
    id SERIAL PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    content TEXT NOT NULL,
    excerpt VARCHAR(500),
    author_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_articles_author_id ON articles(author_id);
CREATE INDEX IF NOT EXISTS idx_articles_created_at ON articles(created_at DESC);

-- 插入示例数据
INSERT INTO users (username, email, password_hash) VALUES
    ('admin', 'admin@shode.com', 'admin123'),
    ('testuser', 'test@shode.com', 'test123')
ON CONFLICT (username) DO NOTHING;

INSERT INTO articles (title, content, excerpt, author_id) VALUES
    ('欢迎使用 Shode', '这是一个使用 Shode 框架构建的 Web 应用示例...', '示例文章摘要', 1),
    ('Shode 框架介绍', 'Shode 是一个现代化的应用开发框架...', '框架介绍', 1)
ON CONFLICT DO NOTHING;
