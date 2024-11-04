BEGIN;

CREATE TYPE user_role_type AS ENUM (
  'root',
  'admin',
  'user'
);

CREATE TABLE users (
  username VARCHAR(20) PRIMARY KEY
);

CREATE TABLE services (
  name VARCHAR(20) PRIMARY KEY
);

CREATE TABLE users_groups (
  id SERIAL PRIMARY KEY,
  username VARCHAR(20) NOT NULL,
  user_role user_role_type NOT NULL,
  service_name VARCHAR(100) NOT NULL,
  created_ts TIMESTAMP DEFAULT NOW(),

  CONSTRAINT fk_users_groups_username
    FOREIGN KEY (username) REFERENCES users(username)
    ON DELETE CASCADE,

  CONSTRAINT fk_users_groups_service_name
    FOREIGN KEY (service_name) REFERENCES services(name)
    ON DELETE CASCADE,

  CONSTRAINT uk_users_groups_username_service_name
    UNIQUE (username, service_name)
);

COMMIT;
