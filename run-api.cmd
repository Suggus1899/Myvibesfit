@echo off
cd /d "%~dp0api"
set DATABASE_URL=postgres://postgres:1234@localhost:5432/myvibesfit?sslmode=disable
set JWT_ACCESS_SECRET=dev-local-secret-do-not-use-in-production
go run ./cmd/api
