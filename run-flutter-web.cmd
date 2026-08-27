@echo off
cd /d "%~dp0app"
flutter run -d web-server --web-port 5555 --web-hostname 127.0.0.1
