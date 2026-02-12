// This PowerShell script adds the Super Admin Management routes to server.go

$filePath = "C:\Users\rajan\.gemini\antigravity\scratch\PeopleOS\Backend\internal\server\server.go"

# Read the file content
$content = Get-Content $filePath -Raw

# Define the new routes to add
$newRoutes = @"

		// Super Admin Management
		r.Route("/admins", func(r chi.Router) {
			r.Post("/", s.superAdminHandler.CreateSuperAdmin)
			r.Get("/", s.superAdminHandler.GetAllSuperAdmins)
		})
"@

# Find the position after the usage route and before System Management
$searchPattern = '(\t\t\t\}\))\r?\n(\r?\n\t\t\t// System Management)'
$replacement = "`$1$newRoutes`$2"

# Replace
$newContent = $content -replace $searchPattern, $replacement

# Write back
Set-Content $filePath -Value $newContent -NoNewline

Write-Host "Routes added successfully!"
