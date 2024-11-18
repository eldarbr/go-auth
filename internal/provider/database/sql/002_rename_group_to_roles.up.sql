BEGIN;

ALTER TABLE "users_groups"
  RENAME TO "users_roles";

ALTER TABLE "users_roles"
  RENAME CONSTRAINT "uk_users_groups_username_service_name"
    TO "uk_users_roles_username_service_name";

ALTER TABLE "users_roles"
  RENAME CONSTRAINT "fk_users_groups_service_name"
    TO "fk_users_roles_service_name";

ALTER TABLE "users_roles"
  RENAME CONSTRAINT "fk_users_groups_username"
    TO "fk_users_roles_username";

COMMIT;
