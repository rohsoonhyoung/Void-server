@echo off
:START
TIMEOUT 10
SETLOCAL EnableExtensions
set EXE=dragon-legend.exe
FOR /F %%x IN ('tasklist /NH /FI "IMAGENAME eq %EXE%"') DO IF %%x == %EXE% goto START

start dragon-legend.exe
goto START