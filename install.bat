@echo off
echo Checking for Go installation...
where go >nul 2>nul
if %errorlevel% neq 0 (
    echo Go is not installed. Please install Go and try again.
    exit /b 1
)

echo Downloading dependencies...
go mod download
if %errorlevel% neq 0 (
    echo Error downloading dependencies.
    exit /b 1
)

echo Building PhantomWP...
go build -o phantomwp.exe
if %errorlevel% neq 0 (
    echo Error building PhantomWP.
    exit /b 1
)

echo PhantomWP has been successfully installed!
echo You can now use it by running: phantomwp.exe
