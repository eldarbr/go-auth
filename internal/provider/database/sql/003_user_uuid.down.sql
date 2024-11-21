BEGIN;

ALTER TABLE "users_roles"
  ADD COLUMN "username" VARCHAR(20) NOT NULL DEFAULT '';

UPDATE "users_roles"
SET "username" = "users"."username"
FROM "users"
WHERE "users"."id" = "users_roles"."user_id";

ALTER TABLE "users_roles"
  ALTER COLUMN "username"
    DROP DEFAULT;

ALTER TABLE "users_roles"
  ADD CONSTRAINT "fk_users_roles_username"
    FOREIGN KEY ("username") REFERENCES "users"("username")
    ON DELETE CASCADE;

ALTER TABLE "users_roles"
  DROP CONSTRAINT "uk_users_roles_user_id_service_name";

ALTER TABLE "users_roles"
  ADD CONSTRAINT "uk_users_roles_username_service_name"
    UNIQUE ("username", "service_name");

ALTER TABLE "users_roles"
  DROP COLUMN "user_id";

ALTER TABLE "users"
  DROP COLUMN "id";

ALTER TABLE "users_roles"
  RENAME CONSTRAINT "users_roles_pkey" TO "users_groups_pkey";

COMMIT;
