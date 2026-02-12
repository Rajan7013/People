"use client"

import { useState, useEffect } from "react"
import { Mail, Phone, Briefcase, Search } from "lucide-react"

import {
    Card,
    CardContent,
    CardDescription,
    CardHeader,
    CardTitle,
} from "@/components/ui/card"
import { Badge } from "@/components/ui/badge"
import { Input } from "@/components/ui/input"
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar"
import api from "@/lib/api"
import { toast } from "sonner"

interface TeamMember {
    id: string
    first_name: string
    last_name: string
    email: string
    phone?: string
    position?: string
    role: string
    status: string
    department?: string
}

export default function ManagerTeamPage() {
    const [team, setTeam] = useState<TeamMember[]>([])
    const [loading, setLoading] = useState(true)
    const [search, setSearch] = useState("")

    useEffect(() => {
        const fetchTeam = async () => {
            try {
                const response = await api.get("/company/manager/team")
                // Check if response data is array or object with data property
                const data = Array.isArray(response.data) ? response.data : (response.data.data || [])
                setTeam(data)
            } catch (error) {
                console.error("Failed to fetch team:", error)
                toast.error("Failed to load team members")
            } finally {
                setLoading(false)
            }
        }

        fetchTeam()
    }, [])

    const filteredTeam = team.filter(member =>
        member.first_name.toLowerCase().includes(search.toLowerCase()) ||
        member.last_name.toLowerCase().includes(search.toLowerCase()) ||
        member.email.toLowerCase().includes(search.toLowerCase()) ||
        (member.position && member.position.toLowerCase().includes(search.toLowerCase()))
    )

    if (loading) {
        return <div className="p-8 text-center">Loading team members...</div>
    }

    return (
        <div className="container mx-auto py-8">
            <div className="flex flex-col md:flex-row justify-between items-start md:items-center gap-4 mb-8">
                <div>
                    <h1 className="text-3xl font-bold tracking-tight">My Team</h1>
                    <p className="text-muted-foreground">Overview of your team members and their status.</p>
                </div>
                <div className="relative w-full md:w-[300px]">
                    <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
                    <Input
                        type="search"
                        placeholder="Search team..."
                        className="pl-8"
                        value={search}
                        onChange={(e) => setSearch(e.target.value)}
                    />
                </div>
            </div>

            {filteredTeam.length === 0 ? (
                <div className="text-center py-12 border rounded-lg bg-muted/10">
                    <h3 className="text-lg font-medium">No team members found</h3>
                    <p className="text-muted-foreground mt-1">You don't have any team members assigned to you yet.</p>
                </div>
            ) : (
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                    {filteredTeam.map((member) => (
                        <Card key={member.id} className="hover:shadow-md transition-shadow">
                            <CardHeader className="flex flex-row items-center gap-4 pb-2">
                                <Avatar className="h-12 w-12">
                                    <AvatarImage src="" /> {/* Placeholder for now */}
                                    <AvatarFallback className="bg-primary/10 text-primary">
                                        {member.first_name[0]}{member.last_name[0]}
                                    </AvatarFallback>
                                </Avatar>
                                <div className="flex-1 overflow-hidden">
                                    <CardTitle className="text-lg truncate">
                                        {member.first_name} {member.last_name}
                                    </CardTitle>
                                    <CardDescription className="truncate">
                                        {member.position || member.role.replace("_", " ")}
                                    </CardDescription>
                                </div>
                                <Badge variant={member.status === 'active' ? 'default' : 'secondary'} className="capitalize">
                                    {member.status}
                                </Badge>
                            </CardHeader>
                            <CardContent>
                                <div className="grid gap-2 text-sm mt-2">
                                    <div className="flex items-center gap-2 text-muted-foreground">
                                        <Mail className="h-4 w-4" />
                                        <span className="truncate">{member.email}</span>
                                    </div>
                                    {member.phone && (
                                        <div className="flex items-center gap-2 text-muted-foreground">
                                            <Phone className="h-4 w-4" />
                                            <span>{member.phone}</span>
                                        </div>
                                    )}
                                    <div className="flex items-center gap-2 text-muted-foreground">
                                        <Briefcase className="h-4 w-4" />
                                        <span>{member.department || "No Department"}</span>
                                    </div>
                                </div>
                            </CardContent>
                        </Card>
                    ))}
                </div>
            )}
        </div>
    )
}
