#!/bin/bash
set -e

# DeCube Capacity Reporting Script
# Generates capacity usage report

OUTPUT_FILE="${OUTPUT_FILE:-capacity-report-$(date +%Y%m%d).json}"

echo "DeCube Capacity Report"
echo "======================"
echo "Date: $(date)"
echo ""

# Collect metrics
METRICS=$(cat <<EOF
{
  "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
  "cluster": {
    "nodes": $(docker ps --filter "name=decube" --format "{{.Names}}" | wc -l),
    "total_cpu": $(docker stats --no-stream --format "{{.CPUPerc}}" $(docker ps -q --filter "name=decube") | awk '{sum+=$1} END {print sum}'),
    "total_memory": $(docker stats --no-stream --format "{{.MemUsage}}" $(docker ps -q --filter "name=decube") | awk '{sum+=$1} END {print sum}')
  },
  "storage": {
    "used": $(df -h /var/lib/decube 2>/dev/null | awk 'NR==2 {print $3}' || echo "0"),
    "available": $(df -h /var/lib/decube 2>/dev/null | awk 'NR==2 {print $4}' || echo "0")
  },
  "network": {
    "bandwidth": "N/A"
  }
}
EOF
)

echo "$METRICS" | jq '.' > "$OUTPUT_FILE"

echo "Capacity report generated: $OUTPUT_FILE"
echo ""
echo "Summary:"
cat "$OUTPUT_FILE" | jq '.'

