BEGIN;

CREATE TYPE user_role_type AS ENUM (
  'root',
  'admin',
  'user'
);

CREATE TABLE user (
  id SERIAL PRIMARY KEY,
  name VARCHAR(100)
);

CREATE TABLE service (
  name VARCHAR(100) PRIMARY KEY
);

CREATE TABLE user_group (
  id SERIAL PRIMARY KEY,
  user_id INT NOT NULL,
  user_role user_role_type NOT NULL,
  service_name VARCHAR(100) NOT NULL,
  created_ts TIMESTAMP DEFAULT NOW(),

  CONSTRAINT fk_user_group_user_id
    FOREIGN KEY user_id REFERENCES user(id)
    ON DELETE CASCADE,

  CONSTRAINT fk_user_group_service_name
    FOREIGN KEY service_name REFERENCES service(name)
    ON DELETE CASCADE
);

COMMIT;
