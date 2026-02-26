# Simple rate limit test
$loginData = @{
    username = "test"
    password = "test"
} | ConvertTo-Json

for ($i = 1; $i -le 6; $i++) {
    try {
        $response = Invoke-RestMethod `
            -Method POST `
            -Uri "http://localhost:1010/auth/login" `
            -Body $loginData `
            -ContentType "application/json" `
            -ErrorAction Stop

        Write-Host "Request ${i}: SUCCESS"
    } catch {
        if ($_.Exception.Response) {
            $statusCode = $_.Exception.Response.StatusCode.value__
            Write-Host "Request ${i}: ERROR ${statusCode}"

            if ($statusCode -eq 429) {
                Write-Host "Rate limit reached "
                break
            }
        } else {
            Write-Host "Request ${i}: NETWORK / SERVER ERROR"
        }
    }

    Start-Sleep -Milliseconds 100
}
