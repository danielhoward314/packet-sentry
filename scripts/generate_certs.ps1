$ErrorActionPreference = 'Stop'

$certsDir = "certs"
if (Test-Path $certsDir) {
    Remove-Item -Recurse -Force $certsDir
}
New-Item -ItemType Directory -Path $certsDir | Out-Null
Push-Location $certsDir

Write-Host "Generating CA key and certificate..."
& openssl genrsa -out ca.key.pem 4096
& openssl req -x509 -new -nodes -key ca.key.pem -sha256 -days 3650 `
    -subj "/CN=Packet Sentry Local Test CA" `
    -out ca.cert.pem

function Generate-Cert {
    param (
        [string]$name,
        [string]$commonName
    )

    Write-Host "Generating $name key and certificate signing request..."
    & openssl genrsa -out "$name.key.pem" 2048
    & openssl req -new -key "$name.key.pem" -subj "/CN=$commonName" -out "$name.csr.pem"

    Write-Host "Creating a SAN config for $name (dev hostname)..."
    $extFile = "$name`_ext.cnf"
    "subjectAltName = DNS:$commonName" | Out-File -Encoding ascii $extFile

    Write-Host "Signing the $name certificate using the CA and SAN config..."
    & openssl x509 -req -in "$name.csr.pem" -CA ca.cert.pem -CAkey ca.key.pem `
        -CAcreateserial -out "$name.cert.pem" -days 365 -sha256 `
        -extfile $extFile
}

Generate-Cert -name "agent_server" -commonName "agent-api.packet-sentry.local"
Generate-Cert -name "gateway_server" -commonName "gateway.packet-sentry.local"
Generate-Cert -name "web_api_server" -commonName "web-api.packet-sentry.local"

Write-Host ""
Write-Host "All certificates and keys generated:"
Write-Host "- CA cert/key:                    ca.cert.pem / ca.key.pem"
Write-Host "- Agent Server cert/key:          agent_server.cert.pem / agent_server.key.pem"
Write-Host "- Gateway Server cert/key:        gateway_server.cert.pem / gateway_server.key.pem"
Write-Host "- Web API Server cert/key:        web_api_server.cert.pem / web_api_server.key.pem"

Pop-Location
