import { EmployeeTable } from "@/components/dashboard/employee-table"

export default function TeamLeadTeamPage() {
    return (
        <div className="space-y-6">
            <div className="flex justify-between items-center">
                <h1 className="text-3xl font-bold tracking-tight">My Team</h1>
            </div>
            <EmployeeTable role="team_lead" />
        </div>
    )
}
