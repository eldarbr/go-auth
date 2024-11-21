BEGIN;

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- previous migration flaw.
ALTER TABLE "users_roles"
  RENAME CONSTRAINT "users_groups_pkey" TO "users_roles_pkey";


-- add users.id = uuid
ALTER TABLE "users"
  ADD COLUMN "id" UUID UNIQUE NOT NULL DEFAULT uuid_generate_v4();


-- START migrate the users_roles table to use users' uuid

ALTER TABLE "users_roles"
  ADD COLUMN "user_id" UUID NOT NULL DEFAULT uuid_nil();

UPDATE "users_roles"
SET "user_id" = "users"."id"
FROM "users"
WHERE "users"."username" = "users_roles"."username";

ALTER TABLE "users_roles"
  ALTER COLUMN "user_id"
    DROP DEFAULT;

ALTER TABLE "users_roles"
  ADD CONSTRAINT "fk_users_roles_user_id"
    FOREIGN KEY ("user_id") REFERENCES "users"("id")
    ON DELETE CASCADE;

ALTER TABLE "users_roles"
  DROP CONSTRAINT "uk_users_roles_username_service_name";

ALTER TABLE "users_roles"
  ADD CONSTRAINT "uk_users_roles_user_id_service_name"
    UNIQUE ("user_id", "service_name");

ALTER TABLE "users_roles"
  DROP COLUMN "username";

-- FINISH migrate the users_roles table to use users' uuid

COMMIT;
