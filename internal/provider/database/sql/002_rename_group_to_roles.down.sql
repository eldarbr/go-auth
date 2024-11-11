BEGIN;

ALTER TABLE "users_roles"
  RENAME CONSTRAINT "fk_users_roles_username"
    TO "fk_users_groups_username";

ALTER TABLE "users_roles"
  RENAME CONSTRAINT "fk_users_roles_service_name"
    TO "fk_users_groups_service_name";

ALTER TABLE "users_roles"
  RENAME CONSTRAINT "uk_users_roles_username_service_name"
    TO "uk_users_groups_username_service_name";

ALTER TABLE "users_roles"
  RENAME TO "users_groups";

COMMIT;
