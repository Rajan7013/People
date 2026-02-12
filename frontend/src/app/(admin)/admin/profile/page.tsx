"use client"

import { useState, useEffect } from "react"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Button } from "@/components/ui/button"
import { toast } from "sonner"
import api from "@/lib/api"
import { Building2, Mail, Phone, MapPin, Calendar, CreditCard } from "lucide-react"

interface OrganizationProfile {
    id: string
    name: string
    email: string // Read-only
    phone: string
    address: string
    industry: string
    size: string
    logo_url: string
    created_at: string // Read-only
    subscription_id: string | null
    plan_name: string
    plan_status: string
    billing_cycle: string
}

export default function OrganizationProfilePage() {
    const [profile, setProfile] = useState<OrganizationProfile | null>(null)
    const [loading, setLoading] = useState(true)
    const [saving, setSaving] = useState(false)
    const [editMode, setEditMode] = useState(false)
    const [formData, setFormData] = useState({
        name: "",
        phone: "",
        address: "",
        logo_url: "",
    })

    useEffect(() => {
        loadProfile()
    }, [])

    const loadProfile = async () => {
        try {
            setLoading(true)
            const response = await api.get("/company/admin/organization")
            setProfile(response.data)
            setFormData({
                name: response.data.name || "",
                phone: response.data.phone || "",
                address: response.data.address || "",
                logo_url: response.data.logo_url || "",
            })
        } catch (error) {
            console.error("Failed to load organization profile:", error)
            toast.error("Failed to load organization profile")
        } finally {
            setLoading(false)
        }
    }

    const handleSave = async () => {
        try {
            setSaving(true)
            await api.put("/company/admin/organization", formData)
            toast.success("Organization profile updated successfully")
            setEditMode(false)
            loadProfile()
        } catch (error: any) {
            console.error("Failed to update profile:", error)
            toast.error(error.response?.data?.message || "Failed to update profile")
        } finally {
            setSaving(false)
        }
    }

    const handleCancel = () => {
        if (profile) {
            setFormData({
                name: profile.name || "",
                phone: profile.phone || "",
                address: profile.address || "",
                logo_url: profile.logo_url || "",
            })
        }
        setEditMode(false)
    }

    if (loading) {
        return (
            <div className="container mx-auto py-10">
                <div className="flex items-center justify-center h-64">
                    <p className="text-muted-foreground">Loading organization profile...</p>
                </div>
            </div>
        )
    }

    if (!profile) {
        return (
            <div className="container mx-auto py-10">
                <div className="flex items-center justify-center h-64">
                    <p className="text-muted-foreground">Organization profile not found</p>
                </div>
            </div>
        )
    }

    return (
        <div className="container mx-auto py-10">
            <div className="mb-6">
                <h1 className="text-3xl font-bold">Organization Profile</h1>
                <p className="text-muted-foreground">
                    View and manage your organization details
                </p>
            </div>

            <div className="grid gap-6 md:grid-cols-2">
                {/* Organization Information - Editable */}
                <Card>
                    <CardHeader>
                        <CardTitle className="flex items-center gap-2">
                            <Building2 className="h-5 w-5" />
                            Organization Information
                        </CardTitle>
                        <CardDescription>
                            Update your organization details
                        </CardDescription>
                    </CardHeader>
                    <CardContent className="space-y-4">
                        <div className="space-y-2">
                            <Label htmlFor="name">Organization Name</Label>
                            <Input
                                id="name"
                                value={formData.name}
                                onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                                disabled={!editMode}
                            />
                        </div>

                        <div className="space-y-2">
                            <Label htmlFor="phone" className="flex items-center gap-2">
                                <Phone className="h-4 w-4" />
                                Phone Number
                            </Label>
                            <Input
                                id="phone"
                                value={formData.phone}
                                onChange={(e) => setFormData({ ...formData, phone: e.target.value })}
                                disabled={!editMode}
                                placeholder="+1 (555) 000-0000"
                            />
                        </div>

                        <div className="space-y-2">
                            <Label htmlFor="address" className="flex items-center gap-2">
                                <MapPin className="h-4 w-4" />
                                Address
                            </Label>
                            <Input
                                id="address"
                                value={formData.address}
                                onChange={(e) => setFormData({ ...formData, address: e.target.value })}
                                disabled={!editMode}
                                placeholder="123 Main St, City, State, ZIP"
                            />
                        </div>

                        <div className="flex gap-2 pt-4">
                            {!editMode ? (
                                <Button onClick={() => setEditMode(true)}>
                                    Edit Information
                                </Button>
                            ) : (
                                <>
                                    <Button onClick={handleSave} disabled={saving}>
                                        {saving ? "Saving..." : "Save Changes"}
                                    </Button>
                                    <Button variant="outline" onClick={handleCancel} disabled={saving}>
                                        Cancel
                                    </Button>
                                </>
                            )}
                        </div>
                    </CardContent>
                </Card>

                {/* Admin Account - Read Only */}
                <Card>
                    <CardHeader>
                        <CardTitle className="flex items-center gap-2">
                            <Mail className="h-5 w-5" />
                            Admin Account
                        </CardTitle>
                        <CardDescription>
                            Account information (managed by super admin)
                        </CardDescription>
                    </CardHeader>
                    <CardContent className="space-y-4">
                        <div className="space-y-2">
                            <Label>Admin Email</Label>
                            <Input value={profile.email} disabled className="bg-muted" />
                            <p className="text-xs text-muted-foreground">
                                Email cannot be changed. Contact super admin if needed.
                            </p>
                        </div>

                        <div className="space-y-2">
                            <Label className="flex items-center gap-2">
                                <Calendar className="h-4 w-4" />
                                Account Created
                            </Label>
                            <Input
                                value={new Date(profile.created_at).toLocaleDateString()}
                                disabled
                                className="bg-muted"
                            />
                        </div>
                    </CardContent>
                </Card>

                {/* Subscription Details - Read Only */}
                <Card className="md:col-span-2">
                    <CardHeader>
                        <CardTitle className="flex items-center gap-2">
                            <CreditCard className="h-5 w-5" />
                            Subscription Details
                        </CardTitle>
                        <CardDescription>
                            Your current subscription plan
                        </CardDescription>
                    </CardHeader>
                    <CardContent>
                        <div className="grid gap-4 md:grid-cols-3">
                            <div className="space-y-2">
                                <Label>Current Plan</Label>
                                <Input
                                    value={profile.plan_name || "No active plan"}
                                    disabled
                                    className="bg-muted"
                                />
                            </div>

                            <div className="space-y-2">
                                <Label>Status</Label>
                                <Input
                                    value={profile.plan_status || "N/A"}
                                    disabled
                                    className="bg-muted capitalize"
                                />
                            </div>

                            <div className="space-y-2">
                                <Label>Billing Cycle</Label>
                                <Input
                                    value={profile.billing_cycle || "N/A"}
                                    disabled
                                    className="bg-muted capitalize"
                                />
                            </div>
                        </div>
                        <p className="text-xs text-muted-foreground mt-4">
                            To upgrade or change your plan, please contact support.
                        </p>
                    </CardContent>
                </Card>
            </div>
        </div>
    )
}
