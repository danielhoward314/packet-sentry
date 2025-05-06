import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { AuthForm } from "@/components/AuthForm";
import { Link } from "react-router-dom";
import { Card, CardContent } from "@/components/ui/card";
import reactLogo from "@/assets/react.svg";
import shadcnLogo from "@/assets/shadcn.svg";
import { Button } from "@/components/ui/button";
import {
  ResetPasswordRequest,
  CredentialType,
  IdentifierType,
} from "@/types/api";
import { useAdminUser } from "@/contexts/AdminUserContext";
import { useEnv } from "@/contexts/EnvContext";
import { toast } from "sonner";

export default function ResetPasswordPage() {
  const navigate = useNavigate();
  const [step, setStep] = useState<1 | 2>(1);
  const [error, setError] = useState<string | null>(null);
  const { adminUser, refreshAdminUser } = useAdminUser();
  const { API_BASE_URL } = useEnv();

  const handleResetPassword = async (formData: FormData) => {
    const oldPassword = formData.get("oldPassword") as string;
    const password = formData.get("password") as string;
    const confirm = formData.get("confirm") as string;

    if (password !== confirm) {
      setError("Passwords do not match.");
      return;
    }
    if (!adminUser?.id) {
      console.log('admin user not in context, refreshing')
      refreshAdminUser()
      if (!adminUser?.id) {
        console.error('admin user not in context after refresh')
        return
      }
    }

    const body: ResetPasswordRequest = {
      credential: oldPassword,
      credentialType: CredentialType.PASSWORD,
      identifier: adminUser.id,
      identifierType: IdentifierType.ID,
      newPassword: password,
      confirmNewPassword: confirm,
    };

    try {
      const response = await fetch(`${API_BASE_URL}/v1/passwords`, {
        headers: { "Content-Type": "application/json" },
        method: "PUT",
        mode: "cors",
        body: JSON.stringify(body),
      });

      if (response.status !== 200) {
        throw new Error("non-200 status for password reset");
      }
      setStep(2);
    } catch (err) {
      toast.error("Failed to reset your password.");
      return;
    }
  };

  const clearError = () => {
    setError(null);
  };

  return (
    <main className="w-full flex flex-col items-center gap-x-4 px-4 gap-t-4 pt-4 overflow-y-auto">
      {step === 1 && (
        <AuthForm
          bottomText={
            <>
              Go back?{" "}
              <Link to="/home" className="underline underline-offset-4">
                Home
              </Link>
            </>
          }
          buttonText="Reset Password"
          error={error}
          fields={[
            { type: "password", label: "Old Password", id: "oldPassword" },
            { type: "password", label: "New Password", id: "password" },
            { type: "password", label: "Confirm New Password", id: "confirm" },
          ]}
          inMainLayout
          onChange={clearError}
          onSubmit={async (formData) => {
            handleResetPassword(formData);
          }}
          subtitle="Confirm password details below."
          title="Reset Password"
        />
      )}

      {step === 2 && (
        <Card className="overflow-hidden p-0 w-3/4 h-full overflow-y-auto">
          <CardContent className="grid p-0 md:grid-cols-2 w-full">
            <div className="flex flex-col gap-6 p-6 md:p-8 w-full">
              <div className="flex flex-col items-center text-center">
                <h1 className="text-2xl font-bold">Success!</h1>
                <p className="text-muted-foreground text-balance">
                  Your password is reset.
                </p>
                <Button
                  onClick={() => navigate("/home")}
                  className="m-8 w-full"
                >
                  Home
                </Button>
              </div>
            </div>
            <div className="bg-muted hidden md:flex justify-center items-center">
              <img
                src={shadcnLogo}
                alt="Logo Light"
                className="block dark:hidden w-md h-md max-w-[200px] max-h-[200px] object-cover"
              />
              <img
                src={reactLogo}
                alt="Logo Dark"
                className="hidden dark:block w-md h-md max-w-[200px] max-h-[200px] object-cover"
              />
            </div>
          </CardContent>
        </Card>
      )}
    </main>
  );
}
