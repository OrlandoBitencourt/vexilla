@echo off
REM Vexilla Benchmark Runner for Windows

echo.
echo Vexilla Performance Benchmark Suite
echo ========================================
echo.

REM Create results directory
if not exist "results" mkdir results

REM Generate timestamp
for /f "tokens=2 delims==" %%a in ('wmic OS Get localdatetime /value') do set "dt=%%a"
set "timestamp=%dt:~0,8%_%dt:~8,6%"

set "RESULTS_FILE=results\benchmark_%timestamp%.txt"
set "SUMMARY_FILE=results\summary_%timestamp%.md"

echo Results will be saved to: %RESULTS_FILE%
echo.

REM Check if benchstat is installed
where benchstat >nul 2>nul
if %ERRORLEVEL% NEQ 0 (
    echo Installing benchstat for statistical analysis...
    go install golang.org/x/perf/cmd/benchstat@latest
)

REM Run benchmarks
echo Running benchmarks...
echo This may take a few minutes...
echo.

go test -bench=. -benchmem -benchtime=3s -timeout=30m > "%RESULTS_FILE%" 2>&1

REM Display results
type "%RESULTS_FILE%"

echo.
echo.
echo ✅ Benchmarks complete!
echo.
echo Results saved to: %RESULTS_FILE%
echo.

REM Display key metrics
echo Key Performance Metrics:
echo.

echo Local Evaluation (Simple):
findstr /C:"BenchmarkLocalEvaluation_Simple-" "%RESULTS_FILE%"

echo.
echo Cache Hit:
findstr /C:"BenchmarkCacheHit-" "%RESULTS_FILE%"

echo.
echo Concurrent Evaluations:
findstr /C:"BenchmarkConcurrentEvaluations-" "%RESULTS_FILE%"

echo.
echo Memory Allocation:
findstr /C:"BenchmarkMemoryAllocation-" "%RESULTS_FILE%"

echo.
echo View full results: type %RESULTS_FILE%
echo.

pause
