CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username VARCHAR(100) NOT NULL,
    password VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    role VARCHAR(50) NOT NULL,
    enableTOTP BOOLEAN DEFAULT FALSE,
    secretTOTP VARCHAR(255)
);

CREATE TABLE IF NOT EXISTS files (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID, -- Cho phép file anonymous upload.
    name VARCHAR(255) NOT NULL,
    password VARCHAR(255),
    type TEXT,
    size BIGINT,
    is_public BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT now(),
    available_from TIMESTAMPTZ,
    available_to TIMESTAMPTZ,
    enable_totp BOOLEAN DEFAULT false,
    share_token TEXT,
    CONSTRAINT files_password_check CHECK (length(password) >= 6),
    CONSTRAINT files_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS filestat (
    file_id UUID NOT NULL,
    download_count BIGINT DEFAULT 0,
    user_download_count BIGINT DEFAULT 0,
    CONSTRAINT filestat_file_id_fkey FOREIGN KEY (file_id) REFERENCES files(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS shared (
    user_id UUID NOT NULL,
    file_id UUID NOT NULL,
    PRIMARY KEY (user_id, file_id),
    CONSTRAINT shared_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT shared_file_id_fkey FOREIGN KEY (file_id) REFERENCES files(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS download (
    download_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    time TIMESTAMPTZ DEFAULT now(),
    user_id UUID,
    file_id UUID NOT NULL,
    CONSTRAINT download_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT download_file_id_fkey FOREIGN KEY (file_id) REFERENCES files(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS jwt_blacklist (
    id SERIAL PRIMARY KEY,
    token TEXT NOT NULL,
    expired_at TIMESTAMP NOT NULL
);

CREATE TABLE IF NOT EXISTS user_totp (
    user_id UUID PRIMARY KEY,
    secret TEXT NOT NULL,
    CONSTRAINT user_totp_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS usersLoginSession (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    cid UUID NOT NULL
);

CREATE PROCEDURE proc_download(f_id UUID, u_id UUID)
LANGUAGE SQL
AS $$
    UPDATE filestat 
    SET 
        download_count = download_count + 1
    WHERE file_id = f_id;

    UPDATE filestat
    SET
        user_download_count = user_download_count + 1
    WHERE file_id = f_id AND NOT EXISTS (SELECT 1 FROM download WHERE user_id = u_id AND file_id = f_id);

    INSERT INTO download (file_id, user_id) VALUES (f_id, u_id);
$$;
