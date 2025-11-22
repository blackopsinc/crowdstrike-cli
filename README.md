# CrowdStrike RTR Batch Service CLI

A command-line tool for executing scripts across multiple CrowdStrike-protected hosts using the Real-Time Response (RTR) batch API. This tool is designed for incident response teams to quickly execute remediation scripts, gather forensic data, or perform security operations across large numbers of endpoints simultaneously.

## Purpose in Incident Response

The CrowdStrike RTR Batch Service CLI is a critical tool for incident response operations, enabling security teams to:

- **Rapid Response**: Execute commands and scripts across hundreds or thousands of endpoints simultaneously, dramatically reducing response time during security incidents
- **Parallel Execution**: Process up to 32 hosts concurrently using goroutines, ensuring efficient use of time during critical incidents
- **Host Discovery**: Automatically search and identify hosts by hostname pattern, allowing quick targeting of affected systems
- **Batch Operations**: Leverage CrowdStrike's batch RTR API to coordinate commands across multiple endpoints without managing individual sessions
- **Forensic Data Collection**: Quickly gather system information, logs, or artifacts from multiple systems in parallel
- **Remediation Automation**: Execute remediation scripts across all affected hosts simultaneously, ensuring consistent response actions
- **Scalability**: Handle large-scale incidents by processing thousands of hosts efficiently through concurrent execution

This tool is particularly valuable during:
- Malware outbreaks requiring rapid containment
- Security incidents affecting multiple systems
- Compliance audits requiring data collection from numerous endpoints
- Proactive security operations across the enterprise

## Installation Guide

### Prerequisites

- **Go 1.21 or later** - [Download Go](https://golang.org/dl/)
- **CrowdStrike API Credentials** - You'll need:
  - `CLIENT_ID` - Your CrowdStrike API Client ID
  - `CLIENT_SECRET` - Your CrowdStrike API Client Secret
- **Network Access** - Access to CrowdStrike API endpoints (typically `https://api.crowdstrike.com`)

### Installation Methods

#### Option 1: Build from Source (Recommended)

1. **Clone or navigate to the project directory:**
   ```bash
   cd crowdstrike
   ```

2. **Build for your current platform:**
   ```bash
   make build
   ```
   The binary will be created in the `bin/` directory as `crowdstrike-cli`.

3. **Or build for a specific platform:**
   ```bash
   # Linux AMD64
   make linux-amd64
   
   # macOS (Apple Silicon)
   make darwin-arm64
   
   # Windows AMD64
   make windows-amd64
   ```

4. **Or build for all platforms:**
   ```bash
   make all
   ```

#### Option 2: Manual Go Build

1. **Navigate to the project directory:**
   ```bash
   cd crowdstrike
   ```

2. **Build the binary:**
   ```bash
   go build -o crowdstrike-cli crowdstrike-cli.go
   ```

3. **Make it executable (Linux/macOS):**
   ```bash
   chmod +x crowdstrike-cli
   ```

### Configuration

1. **Create a `.env` file in the project directory:**
   ```bash
   touch .env
   ```

2. **Add your CrowdStrike API credentials:**
   ```env
   CLIENT_ID=your_client_id_here
   CLIENT_SECRET=your_client_secret_here
   ```

   **Note:** Keep your `.env` file secure and never commit it to version control. Add `.env` to your `.gitignore` file.

3. **Alternative: Set environment variables directly:**
   ```bash
   # Linux/macOS
   export CLIENT_ID="your_client_id_here"
   export CLIENT_SECRET="your_client_secret_here"
   
   # Windows PowerShell
   $env:CLIENT_ID="your_client_id_here"
   $env:CLIENT_SECRET="your_client_secret_here"
   ```

## How-To Guide

### Basic Usage

The tool requires two arguments:
1. **Hostname pattern** - The hostname or pattern to search for
2. **Script/Command** - The script or command to execute on matching hosts

```bash
./crowdstrike-cli <hostname> <script>
```

### Examples

#### Example 1: Execute a PowerShell Script on Windows Hosts

```bash
./crowdstrike-cli "WIN-*" "Get-Process | Select-Object Name, Id, CPU | Format-Table"
```

This command:
- Searches for all hosts with hostnames starting with "WIN-"
- Executes the PowerShell command to list running processes
- Displays the output from each host

#### Example 2: Run a Bash Script on Linux Hosts

```bash
./crowdstrike-cli "linux-server*" "ps aux | grep -i suspicious"
```

This command:
- Finds all hosts with hostnames matching "linux-server*"
- Runs a process search command
- Shows results from all matching hosts

#### Example 3: Collect System Information

```bash
./crowdstrike-cli "prod-*" "systeminfo"
```

This command:
- Targets all production hosts (hostnames starting with "prod-")
- Collects system information
- Outputs results in parallel

#### Example 4: Execute a Multi-line Script

For complex scripts, you can pass them as a single string:

```bash
./crowdstrike-cli "web-*" "Get-EventLog -LogName Security -Newest 100 | Where-Object {$_.EntryType -eq 'FailureAudit'}"
```

### Understanding the Output

The tool executes commands in parallel across all matching hosts (up to 32 concurrent executions). Output from each host is displayed as it completes. The tool:

1. Authenticates with CrowdStrike API using your credentials
2. Searches for hosts matching your hostname pattern
3. Initializes RTR batch sessions for each host
4. Executes your script/command on each host concurrently
5. Displays the stdout output from each host

### Best Practices

1. **Test on a Small Group First**: Before running commands on hundreds of hosts, test with a specific hostname or small pattern:
   ```bash
   ./crowdstrike-cli "test-server-01" "whoami"
   ```

2. **Use Specific Hostname Patterns**: Be precise with hostname patterns to avoid unintended hosts:
   ```bash
   # Good - specific
   ./crowdstrike-cli "prod-web-*" "command"
   
   # Risky - too broad
   ./crowdstrike-cli "*" "command"
   ```

3. **Script Timeout Considerations**: The tool uses default timeouts (30 seconds for session, 10 minutes for command execution). For long-running scripts, ensure they complete within these limits.

4. **Monitor Output**: The parallel execution means output may appear interleaved. Review carefully to identify which output corresponds to which host.

5. **Error Handling**: If a host is offline or unreachable, the tool will continue processing other hosts. Check the output for error messages.

### Troubleshooting

#### Authentication Errors

If you see authentication errors:
- Verify your `CLIENT_ID` and `CLIENT_SECRET` are correct
- Ensure your `.env` file is in the same directory as the binary
- Check that environment variables are set correctly

#### No Hosts Found

If no hosts are found:
- Verify the hostname pattern matches your CrowdStrike environment
- Check that hosts are online and reporting to CrowdStrike
- Try a more specific or different hostname pattern

#### Command Execution Failures

If commands fail:
- Verify the command syntax is correct for the target OS
- Check that the command is available on the target systems
- Ensure you have appropriate permissions via CrowdStrike RTR

## Architecture

The tool uses:
- **CrowdStrike RTR Batch API** for coordinated command execution
- **Go goroutines** for parallel execution (up to 32 concurrent workers)
- **Semaphore pattern** to limit concurrent operations and prevent API rate limiting
- **Environment variable loading** from `.env` files for secure credential management

## Security Considerations

- **Never commit `.env` files** containing API credentials to version control
- **Use least-privilege API credentials** with only RTR permissions
- **Review scripts before execution** to ensure they don't cause unintended side effects
- **Monitor API rate limits** when executing on very large numbers of hosts
- **Log all operations** for audit and compliance purposes

## License

GPL-3.0 license

## Contributing
