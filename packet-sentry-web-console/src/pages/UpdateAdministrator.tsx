import { UpdateAdministratorForm } from "@/components/UpdateAdministratorForm";
import MainContentCardLayout from "@/layouts/MainContentCardLayout";
import { getAdministrator, updateAdministrator } from "@/lib/api";
import { GetAdministratorResponse } from "@/types/api";
import { useEffect, useState } from "react";
import { useParams } from "react-router-dom";
import { toast } from "sonner";
import { AlertCircle, Loader2 } from "lucide-react";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";

export default function UpdateAdministratorPage() {
  const { id } = useParams();
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [existingAdmin, setExistingAdmin] =
    useState<GetAdministratorResponse | null>(null);

  useEffect(() => {
    if (!id) return;
    const fetchData = async () => {
      try {
        const responseData = await getAdministrator(id);
        setExistingAdmin(responseData);
      } catch (err) {
        console.error(err);
        setError("failed to fetch existing admin");
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  }, [id]);

  const handleUpdateAdministrator = async (formData: FormData) => {
    if (!id) return;

    const displayName = formData.get("displayName") as string;
    const email = formData.get("email") as string;
    const authorizationRole = formData.get("authorizationRole") as string;

    const response = await updateAdministrator(id, {
      email,
      displayName,
      authorizationRole,
    });

    if (response.status < 200 || response.status >= 400) {
      toast.error(
        `Failed to make updates to ${existingAdmin?.displayName ?? "the administrator"}.`,
      );
      return;
    }

    try {
      const responseData = await getAdministrator(id);
      setExistingAdmin(responseData);
    } catch (err) {
      console.error(err);
      setError("failed to fetch existing admin");
    }

    toast.success(
      `Successfully updated ${existingAdmin?.displayName ?? "the administrator"}.`,
    );
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
            Failed to load administrators in your organization.
          </AlertDescription>
        </Alert>
      </MainContentCardLayout>
    );
  }

  return (
    <MainContentCardLayout
      cardDescription="Make changes to an administrator in your organization."
      cardTitle="Update Administrator"
    >
      {loading ? (
        <div className="flex justify-center items-center h-screen">
          <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
        </div>
      ) : (
        <UpdateAdministratorForm
          existingAdmin={existingAdmin ?? ({} as GetAdministratorResponse)}
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
          onSubmit={handleUpdateAdministrator}
        />
      )}
    </MainContentCardLayout>
  );
}
