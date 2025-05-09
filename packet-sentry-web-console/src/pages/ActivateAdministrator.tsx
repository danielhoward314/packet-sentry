import { useState } from "react";
import { useNavigate, useSearchParams } from "react-router-dom";
import { AuthForm } from "@/components/AuthForm";
import { Card, CardContent } from "@/components/ui/card";
import reactLogo from "@/assets/react.svg";
import shadcnLogo from "@/assets/shadcn.svg";
import { Button } from "@/components/ui/button";
import { activateAdministrator } from "@/lib/api";
import { toast } from "sonner";

export default function ActivateAdministratorPage() {
  const [searchParams] = useSearchParams();
  const token = searchParams.get("token");
  const navigate = useNavigate();
  const [step, setStep] = useState<1 | 2>(1);
  const [error, setError] = useState<string | null>(null);

  const handleActivation = async (formData: FormData) => {
    if (!token || token === "") return;
    const verificationCode = formData.get("verificationCode") as string;
    const password = formData.get("password") as string;
    const confirm = formData.get("confirm") as string;

    if (password !== confirm) {
      setError("Passwords do not match.");
      return;
    }

    try {
      const response = await activateAdministrator({
        token,
        verificationCode,
        password,
      });

      if (response.status < 200 || response.status >= 400) {
        throw new Error("non-200 status for admin activation");
      }
      setStep(2);
    } catch (err) {
      console.error(err);
      toast.error("Failed to activate your account");
      setStep(1);
    }
  };

  const clearError = () => {
    setError(null);
  };

  return (
    <main className="w-full flex flex-col items-center gap-x-4 px-4 gap-t-4 pt-4 overflow-y-auto">
      {step === 1 && (
        <AuthForm
          buttonText="Save"
          error={error}
          fields={[
            {
              type: "text",
              label: "Verification Code",
              id: "verificationCode",
            },
            { type: "password", label: "New Password", id: "password" },
            { type: "password", label: "Confirm Password", id: "confirm" },
          ]}
          onChange={clearError}
          onSubmit={async (formData) => {
            handleActivation(formData);
          }}
          subtitle="Check your email for an activation email."
          title="Activate Account"
        />
      )}

      {step === 2 && (
        <Card className="overflow-hidden p-0 w-3/4 h-full overflow-y-auto">
          <CardContent className="grid p-0 md:grid-cols-2 w-full">
            <div className="flex flex-col gap-6 p-6 md:p-8 w-full">
              <div className="flex flex-col items-center text-center">
                <h1 className="text-2xl font-bold">Success!</h1>
                <p className="text-muted-foreground text-balance">
                  Your account is activated.
                </p>
                <Button
                  onClick={() => navigate("/login")}
                  className="m-8 w-full"
                >
                  Login
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
