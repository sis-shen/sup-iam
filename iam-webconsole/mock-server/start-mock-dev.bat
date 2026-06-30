@echo off
set IAM_MOCK=true
start "MockServer" /B node "%~dp0\server.js"
start "ViteDev" /B npx.cmd vite --host 0.0.0.0 --port 5173
echo Mock server and Vite dev server started.
echo Mock API: http://localhost:3001/api/v1
echo Frontend: http://localhost:5173
