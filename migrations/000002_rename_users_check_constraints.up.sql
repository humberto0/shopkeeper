-- Align check constraint names with the <table>_<column>_<rule>_check convention.
-- RENAME CONSTRAINT is a catalog-only change: no table rewrite or revalidation,
-- just a brief ACCESS EXCLUSIVE lock.
ALTER TABLE users RENAME CONSTRAINT users_name_not_blank TO users_name_not_blank_check;
ALTER TABLE users RENAME CONSTRAINT users_role_valid TO users_role_valid_check;
