@echo off

rem Init reqs
cls

rem Init env
set SCRIPT_DIR=%~dp0
set PATH=%PATH%;%GOPATH%\bin;%SCRIPT_DIR%

echo ==============================
echo GO Environment
echo ==============================
go env
echo ------------------------------
