"use client";

import { SettingsDetails } from "@/components/SettingsDetails";
import { Theme, useTheme } from "@/contexts/ThemeProvider";
import MainContentCardLayout from "@/layouts/MainContentCardLayout";
import { toast } from "sonner";

export default function SettingsPage() {
  const { setTheme } = useTheme();

  const handleSettingsSave = async (formData: FormData) => {
    const fullName = formData.get("fullName") as string;
    const theme = formData.get("theme") as string;
    const email = formData.get("email") as string;

    console.log("Full Name:", fullName);
    console.log("Theme:", theme);
    console.log("Email:", email);

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
