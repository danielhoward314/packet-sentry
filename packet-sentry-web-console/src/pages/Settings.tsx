"use client";

import { SettingsDetails } from "@/components/SettingsDetails";
import { useAdminUser } from "@/contexts/AdminUserContext";
import { Theme, useTheme } from "@/contexts/ThemeProvider";
import MainContentCardLayout from "@/layouts/MainContentCardLayout";
import { updateAdministrators } from "@/lib/api";
import { toast } from "sonner";

export default function SettingsPage() {
  const { setTheme } = useTheme();
  const { adminUser, refreshAdminUser } = useAdminUser();

  const handleSettingsSave = async (formData: FormData) => {
    const fullName = formData.get("fullName") as string;
    const theme = formData.get("theme") as string;
    const email = formData.get("email") as string;

    if (!adminUser?.id) {
      console.log("admin user not in context, refreshing");
      refreshAdminUser();
      if (!adminUser?.id) {
        console.error("admin user not in context after refresh");
        return;
      }
    }

    const response = await updateAdministrators(adminUser.id, {
      email,
      authorizationRole: '',
      displayName: fullName,
    })

    if (response.status < 200 || response.status >= 400) {
      toast.error('Failed to save your settings.')
      return
    }

    if (theme === "light" || theme === "dark" || theme === "system") {
      setTheme(theme as Theme);
    } else {
      console.warn("Invalid theme selected:", theme);
    }

    toast.success("Your settings have been saved.");
  };

  return (
    <MainContentCardLayout
      cardDescription="Manage the settings for your administrator profile."
      cardTitle="Settings"
    >
      <SettingsDetails
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
    </MainContentCardLayout>
  );
}
