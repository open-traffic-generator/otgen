# OTGen: Open Traffic Generator CLI Tool

## How to use

```Shell
otgen run 
  --api https://otg-api-endpoint      # OTG server API endpoint. Required
  [--yaml otg.yml | --json otg.json]  # OTG model file. If not provided, will use stdin
  [--xeta 2]                          # How long to wait before forcing traffic to stop. In multiples of ETA. Example: 1.5 (default 2)
  [--timeout 60]                      # Timeout for API communication with OTG server
  [--insecure]                        # Ignore X.509 certificate validation
````