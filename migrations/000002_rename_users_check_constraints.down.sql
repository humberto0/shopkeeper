ALTER TABLE users RENAME CONSTRAINT users_role_valid_check TO users_role_valid;
ALTER TABLE users RENAME CONSTRAINT users_name_not_blank_check TO users_name_not_blank;
