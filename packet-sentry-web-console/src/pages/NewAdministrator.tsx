import { NewAdministratorForm } from "@/components/NewAdministratorForm";
import { useAdminUser } from "@/contexts/AdminUserContext";
import MainContentCardLayout from "@/layouts/MainContentCardLayout";
import { createAdministrator } from "@/lib/api";
import { useState } from "react";
import { toast } from "sonner";

export default function NewAdministratorPage() {
  const { adminUser, refreshAdminUser } = useAdminUser();
  const [step, setStep] = useState<1 | 2>(1);

  const handleCreateAdministrator = async (formData: FormData) => {
    const displayName = formData.get("displayName") as string;
    const email = formData.get("email") as string;
    const authorizationRole = formData.get("authorizationRole") as string;
    if (!adminUser?.id) {
      console.log("admin user not in context, refreshing");
      refreshAdminUser();
      if (!adminUser?.id) {
        console.error("admin user not in context after refresh");
        return;
      }
    }

    const response = await createAdministrator({
      organizationId: adminUser.organizationId,
      email,
      displayName,
      authorizationRole,
    });

    if (response.status < 200 || response.status >= 400) {
      toast.error("Failed to save your settings.");
      return;
    }
    setStep(2);
    toast.success("Your settings have been saved.");
  };

  return (
    <MainContentCardLayout
      cardDescription="Send a welcome email to a new administrator in your organization."
      cardTitle="New Administrator"
    >
      <NewAdministratorForm
        step={step}
        fields={[
          { type: "text", label: "Full Name", id: "displayName" },
          { type: "email", label: "Email", id: "email" },
          {
            type: "select",
            label: "Authorization Role",
            id: "authorizationRole",
            options: [
              { value: "PRIMARY_ADMIN", label: "Primary Administrator" },
              { value: "SECONDARY_ADMIN", label: "Secondary Administrator" },
            ],
            default: "PRIMARY_ADMIN",
          },
        ]}
        onSubmit={handleCreateAdministrator}
      />
    </MainContentCardLayout>
  );
}
