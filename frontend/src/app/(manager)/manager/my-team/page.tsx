import { EmployeeTable } from "@/components/dashboard/employee-table"

export default function ManagerTeamPage() {
    return (
        <div className="space-y-6">
            <div className="flex justify-between items-center">
                <h1 className="text-3xl font-bold tracking-tight">My Team</h1>
            </div>
            {/* EmployeeTable will need 'role' prop refactor to filter correctly, 
          but backend GetEmployees now handles strict filtering regardless of frontend query */}
            <EmployeeTable role="manager" />
        </div>
    )
}
