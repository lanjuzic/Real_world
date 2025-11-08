-- ================================================
--  RealWorld Example App - PostgreSQL 初始化脚本
--  Author: PQC / lanjuzi
--  Date: 2025-10-30
--  Description: 初始化数据库、表结构、索引和约束
-- ================================================

-- ========== 清理旧表（开发环境用） ==========
DROP TABLE IF EXISTS article_tags, tags, favorites, follows, comments, articles, users CASCADE;

-- ========== 创建数据库（如果还没创建） ==========
-- ⚠️ 如果你是直接执行在指定 db（如 realworld_db）中，可跳过此步
-- CREATE DATABASE realworld_db;
-- \c realworld_db;

-- ================================================
-- USERS 表 - 用户信息
-- ================================================
CREATE TABLE users (
    id              SERIAL PRIMARY KEY,
    username        VARCHAR(50) NOT NULL UNIQUE,
    email           VARCHAR(120) NOT NULL UNIQUE,
    password_hash   VARCHAR(255) NOT NULL,
    bio             TEXT,
    image           TEXT,
    created_at      TIMESTAMP DEFAULT NOW(),
    updated_at      TIMESTAMP DEFAULT NOW()
);

-- ================================================
-- ARTICLES 表 - 文章
-- ================================================
CREATE TABLE articles (
    id              SERIAL PRIMARY KEY,
    slug            VARCHAR(255) UNIQUE NOT NULL,
    title           VARCHAR(255) NOT NULL,
    description     TEXT,
    body            TEXT,
    author_id       INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at      TIMESTAMP DEFAULT NOW(),
    updated_at      TIMESTAMP DEFAULT NOW()
);
CREATE INDEX idx_articles_author_id ON articles(author_id);

-- ================================================
-- COMMENTS 表 - 评论
-- ================================================
CREATE TABLE comments (
    id              SERIAL PRIMARY KEY,
    body            TEXT NOT NULL,
    author_id       INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    article_id      INT NOT NULL REFERENCES articles(id) ON DELETE CASCADE,
    created_at      TIMESTAMP DEFAULT NOW(),
    updated_at      TIMESTAMP DEFAULT NOW()
);
CREATE INDEX idx_comments_article_id ON comments(article_id);
CREATE INDEX idx_comments_author_id  ON comments(author_id);

-- ================================================
-- FOLLOWS 表 - 用户关注关系
-- ================================================
CREATE TABLE follows (
    follower_id     INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    followee_id     INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at      TIMESTAMP DEFAULT NOW(),
    PRIMARY KEY (follower_id, followee_id)
);
CREATE INDEX idx_follows_follower_id ON follows(follower_id);
CREATE INDEX idx_follows_followee_id ON follows(followee_id);

-- ================================================
-- FAVORITES 表 - 收藏关系（用户 <-> 文章）
-- ================================================
CREATE TABLE favorites (
    user_id         INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    article_id      INT NOT NULL REFERENCES articles(id) ON DELETE CASCADE,
    created_at      TIMESTAMP DEFAULT NOW(),
    PRIMARY KEY (user_id, article_id)
);
CREATE INDEX idx_favorites_user_id    ON favorites(user_id);
CREATE INDEX idx_favorites_article_id ON favorites(article_id);

-- ================================================
-- TAGS 表 - 标签
-- ================================================
CREATE TABLE tags (
    id              SERIAL PRIMARY KEY,
    name            VARCHAR(50) NOT NULL UNIQUE
);

-- ================================================
-- ARTICLE_TAGS 表 - 文章与标签多对多关系
-- ================================================
CREATE TABLE article_tags (
    article_id      INT NOT NULL REFERENCES articles(id) ON DELETE CASCADE,
    tag_id          INT NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (article_id, tag_id)
);
CREATE INDEX idx_article_tags_article_id ON article_tags(article_id);
CREATE INDEX idx_article_tags_tag_id     ON article_tags(tag_id);

-- ================================================
-- 可选优化：未来分区/扩展建议
-- ================================================
-- 1. 热数据与冷数据分区（articles/comments 可按时间分区）
-- 2. 用户行为日志建议单独库 log_db 存放
-- 3. 若数据量大，article_tags 可单独 shard 化

-- ================================================
-- 初始化完成
-- ================================================
SELECT '✅ RealWorld 数据库初始化完成！' AS status;
