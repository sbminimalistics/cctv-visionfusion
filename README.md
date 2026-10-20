# CCTV VisionFusion

CCTV VisionFusion is an event-driven security platform that combines inputs from CCTV cameras and physical motion sensors to eliminate false alarms. The system uses configurable event combinations to produce reliable alert triggers. By utilizing this smart event overlay strategy, the platform drastically minimizes false positives before forwarding verified alerts to the UI dashboard web app. The dashboard serves as the central control hub, presenting users with connected device statuses, historical data, real-time alert feeds, and system configurations.

## Prerequisites

### System Requirements (Production Host)
* **Operating System:** Ubuntu 24.04
* **Container Runtime:** Docker Engine (Standalone)
* **User Identity:** Non-root user with a fixed UID (e.g., UID 1000) for secure volume bind-mounting

### Infrastructure & Deployment Tooling
* **Infrastructure as Code (IaC):** Terraform v1.8+ (Primary deployment engine for production container states)

### Development & Compilation Tooling (Local Workstation)
To compile the Protocol Buffer schemas or run services locally, you require:

* **Container Infrastructure:** Docker Engine
* **Go Ecosystem:** Go 1.26.2+ and the `protoc-gen-go` binary plugin
* **Python Ecosystem:** Python 3.15+ and `uv` or `pip` package manager
* **Node.js Ecosystem:** Node.js 24.16.0+ (LTS)
* **Schema Compiler:** Protocol Buffers Compiler (`protoc` v25+) mapped to your system PATH

## Installation

```powershell
PS C:\cctv-visionfusion> powershell scripts/install.ps1
```

## Protobuf schemas generation

```powershell
PS C:\cctv-visionfusion> powershell scripts/proto.ps1
```

## Development

before you start development script: install dependencies for each sub-project

```powershell
PS C:\cctv-visionfusion> powershell scripts/dev.ps1
```

## Deployment

before you deploy: create secret.tfvars and fill it with values for sensitive terraform variables

```powershell
PS C:\cctv-visionfusion\deployments\terraform> terraform init
PS C:\cctv-visionfusion\deployments\terraform> terraform apply -var-file="secret.tfvars"
```
