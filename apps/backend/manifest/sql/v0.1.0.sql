-- sys_user table
CREATE TABLE IF NOT EXISTS sys_user (
    id          INT PRIMARY KEY AUTO_INCREMENT,
    username    VARCHAR(64)  NOT NULL,
    password    VARCHAR(256) NOT NULL,
    nickname    VARCHAR(64)  NOT NULL DEFAULT '',
    email       VARCHAR(128) NOT NULL DEFAULT '',
    phone       VARCHAR(20)  NOT NULL DEFAULT '',
    sex         TINYINT      NOT NULL DEFAULT 0,
    avatar      VARCHAR(512) NOT NULL DEFAULT '',
    status      TINYINT      NOT NULL DEFAULT 1,
    remark      VARCHAR(512) NOT NULL DEFAULT '',
    login_date  DATETIME,
    created_at  DATETIME,
    updated_at  DATETIME,
    deleted_at  DATETIME,
    UNIQUE(username)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- Default admin user (password: admin123, bcrypt hash)
INSERT IGNORE INTO sys_user (username, password, nickname, status, created_at, updated_at)
VALUES ('admin', '$2a$10$6u4IIEd63chleDWJIY6.NewSU7YrpBQ0Tbp.KfLiG71NQrRlL9qTe', '管理员', 1, NOW(), NOW());
