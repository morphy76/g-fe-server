CREATE USER feat_user WITH ENCRYPTED PASSWORD 'feat_password';
CREATE USER sonar_user WITH ENCRYPTED PASSWORD 'sonar_password';

CREATE DATABASE feat_db;
CREATE DATABASE sonar_db;

GRANT ALL PRIVILEGES ON DATABASE feat_db TO feat_user;
GRANT ALL PRIVILEGES ON DATABASE sonar_db TO sonar_user;

\connect feat_db;

CREATE SCHEMA IF NOT EXISTS feat_schema
  AUTHORIZATION feat_user;

COMMENT ON SCHEMA feat_schema
  IS 'schema for feat_user';

GRANT USAGE ON SCHEMA feat_schema TO feat_user;
GRANT ALL ON SCHEMA feat_schema TO feat_user;

\connect sonar_db;

CREATE SCHEMA IF NOT EXISTS sonar_schema
  AUTHORIZATION sonar_user;

COMMENT ON SCHEMA sonar_schema
  IS 'schema for sonar_user';

GRANT USAGE ON SCHEMA sonar_schema TO sonar_user;
GRANT ALL ON SCHEMA sonar_schema TO sonar_user;
