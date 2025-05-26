"use client";

import { SettingsDetails } from "@/components/SettingsDetails";
import { useAdminUser } from "@/contexts/AdminUserContext";
import { Theme, useTheme } from "@/contexts/ThemeProvider";
import MainContentCardLayout from "@/layouts/MainContentCardLayout";
import { getAdministrator, updateAdministrator } from "@/lib/api";
import { GetAdministratorResponse } from "@/types/api";
import { useEffect, useState } from "react";
import { toast } from "sonner";
import { AlertCircle, Loader2 } from "lucide-react";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";

export default function SettingsPage() {
  const { setTheme } = useTheme();
  const { adminUser, refreshAdminUser } = useAdminUser();
  const [existingAdmin, setExistingAdmin] =
    useState<GetAdministratorResponse | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (!adminUser?.id) {
      refreshAdminUser();
      if (!adminUser?.id) {
        console.error("admin user not in context after refresh");
        return;
      }
    }

    const fetchData = async () => {
      try {
        const responseData = await getAdministrator(adminUser.id);
        setExistingAdmin(responseData);
      } catch (err) {
        console.error(err);
        setError("failed to fetch existing admin");
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  }, [adminUser?.id]);

  const handleSettingsSave = async (formData: FormData) => {
    const fullName = formData.get("fullName") as string;
    const theme = formData.get("theme") as string;
    const email = formData.get("email") as string;

    if (!adminUser?.id) {
      refreshAdminUser();
      if (!adminUser?.id) {
        console.error("admin user not in context after refresh");
        return;
      }
    }

    const response = await updateAdministrator(adminUser.id, {
      email,
      authorizationRole: "",
      displayName: fullName,
    });

    if (response.status < 200 || response.status >= 400) {
      toast.error("Failed to save your settings.");
      return;
    }

    if (theme === "light" || theme === "dark" || theme === "system") {
      setTheme(theme as Theme);
    } else {
      console.warn("Invalid theme selected:", theme);
    }

    try {
      const responseData = await getAdministrator(adminUser.id);
      setExistingAdmin(responseData);
    } catch (err) {
      console.error(err);
      setError("failed to fetch existing admin");
    }

    toast.success("Your settings have been saved.");
  };

  if (error) {
    return (
      <MainContentCardLayout
        cardDescription="Make changes to an administrator in your organization."
        cardTitle="Update Administrator"
      >
        <Alert variant="destructive">
          <AlertCircle className="h-4 w-4" />
          <AlertTitle>Error</AlertTitle>
          <AlertDescription>
            Failed to load your existing settings.
          </AlertDescription>
        </Alert>
      </MainContentCardLayout>
    );
  }

  return (
    <MainContentCardLayout
      cardDescription="Manage the settings for your administrator profile."
      cardTitle="Settings"
    >
      {loading ? (
        <div className="flex justify-center items-center h-screen">
          <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
        </div>
      ) : (
        <SettingsDetails
          existingDisplayName={existingAdmin?.displayName ?? ""}
          existingEmail={existingAdmin?.email ?? ""}
          fields={[
            { type: "text", label: "Full Name", id: "fullName" },
            { type: "email", label: "Email", id: "email" },
            {
              type: "select",
              label: "Theme",
              id: "theme",
              options: [
                { value: "system", label: "System" },
                { value: "light", label: "Light" },
                { value: "dark", label: "Dark" },
              ],
            },
          ]}
          onSubmit={handleSettingsSave}
        />
      )}
    </MainContentCardLayout>
  );
}
