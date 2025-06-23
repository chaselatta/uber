#!/bin/sh

echo "=========================================="
echo "UBER EXECUTION REPORT"
echo "=========================================="
echo ""

echo "📋 EXECUTION DETAILS:"
echo "  Command executed: $UBER_EXECUTED_COMMAND"
echo "  Tool found in: $UBER_EXECUTED_TOOL_PATH"
echo "  Arguments: $UBER_ARGS"
echo ""

echo "⏱️  TIMING BREAKDOWN:"
echo "  Tool discovery: ${UBER_TIMING_FIND_TOOL_MS}ms"
echo "  Environment setup: ${UBER_TIMING_ENV_SETUP_MS}ms"
echo "  Tool execution: ${UBER_TIMING_EXECUTION_MS}ms"
echo "  Total time: ${UBER_TOTAL_TIME_MS}ms"
echo ""

echo "🔧 UBER CONTEXT:"
echo "  Uber binary: $UBER_BIN_PATH"
echo "  Project root: $UBER_PROJECT_ROOT"
if [ "$UBER_VERBOSE" = "1" ]; then
  echo "  Verbose mode: enabled"
else
  echo "  Verbose mode: disabled"
fi
if [ -n "$UBER_GLOBAL_COMMAND_ARGS" ]; then
  echo "  Global args: $UBER_GLOBAL_COMMAND_ARGS"
fi
echo ""

echo "📊 PERFORMANCE SUMMARY:"
if [ "$UBER_TIMING_FIND_TOOL_MS" -gt 100 ]; then
  echo "  ⚠️  Tool discovery took longer than expected (${UBER_TIMING_FIND_TOOL_MS}ms)"
else
  echo "  ✅ Tool discovery was fast (${UBER_TIMING_FIND_TOOL_MS}ms)"
fi

if [ "$UBER_TIMING_ENV_SETUP_MS" -gt 50 ]; then
  echo "  ⚠️  Environment setup took longer than expected (${UBER_TIMING_ENV_SETUP_MS}ms)"
else
  echo "  ✅ Environment setup was fast (${UBER_TIMING_ENV_SETUP_MS}ms)"
fi

if [ "$UBER_TIMING_EXECUTION_MS" -gt 1000 ]; then
  echo "  ⚠️  Tool execution took longer than expected (${UBER_TIMING_EXECUTION_MS}ms)"
else
  echo "  ✅ Tool execution was reasonable (${UBER_TIMING_EXECUTION_MS}ms)"
fi
echo ""

echo "=========================================="
echo "END REPORT"
echo "=========================================="
