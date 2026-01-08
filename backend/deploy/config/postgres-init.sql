-- Create auth_db if it doesn't exist
SELECT 'CREATE DATABASE auth_db'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'auth_db')\gexec
