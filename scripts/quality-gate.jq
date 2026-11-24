def measure($key):
  [.conditions[] | select(.metricKey == $key and .status == "OK") | .actualValue | tonumber]
  | if length == 1 then .[0] else error("Condición ausente o fallida: " + $key) end;

.projectStatus
| .status == "OK" and (.ignoredConditions == false)
  and (measure("new_coverage") >= 60)
  and (measure("new_duplicated_lines_density") < 3)
  and (measure("new_bugs") == 0)
  and (measure("new_vulnerabilities") == 0)
  and (measure("new_security_hotspots_reviewed") == 100)
