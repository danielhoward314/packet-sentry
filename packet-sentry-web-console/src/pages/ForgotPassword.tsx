import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { AuthForm } from "@/components/AuthForm";
import { ModeToggle } from "@/components/ModeToggle";
import { Link } from "react-router-dom";
import { Card, CardContent } from "@/components/ui/card";
import reactLogo from "@/assets/react.svg";
import shadcnLogo from "@/assets/shadcn.svg";
import { Button } from "@/components/ui/button";
import {
  CredentialType,
  IdentifierType,
  ResetPasswordRequest,
  ResetVerifyRequest,
} from "@/types/api";
import { useEnv } from "@/contexts/EnvContext";
import { toast } from "sonner";

export default function ForgotPasswordPage() {
  const navigate = useNavigate();
  const [step, setStep] = useState<1 | 2 | 3>(1);
  const [email, setEmail] = useState("");
  const [error, setError] = useState<string | null>(null);
  const { API_BASE_URL } = useEnv();

  const handleSendCode = async (formData: FormData) => {
    const email = formData.get("email") as string;

    const body: ResetVerifyRequest = {
      email: email,
    };

    try {
      const response = await fetch(`${API_BASE_URL}/v1/reset-verify`, {
        headers: { "Content-Type": "application/json" },
        method: "POST",
        mode: "cors",
        body: JSON.stringify(body),
      });

      if (response.status !== 200) {
        throw new Error(
          "non-200 status for forgot password email verify request",
        );
      }
      setEmail(email);
      setStep(2);
    } catch (err) {
      toast.error("Failed to send email for password reset.");
      return;
    }
  };

  const handleResetPassword = async (formData: FormData) => {
    const verificationCode = formData.get("verificationCode") as string;
    const password = formData.get("password") as string;
    const confirm = formData.get("confirm") as string;

    if (password !== confirm) {
      setError("Passwords do not match.");
      return;
    }

    const body: ResetPasswordRequest = {
      credential: verificationCode,
      credentialType: CredentialType.EMAIL_VERIFICATION_CODE,
      identifier: email,
      identifierType: IdentifierType.EMAIL,
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

      setStep(3);
    } catch (err) {
      toast.error("Failed to reset password.");
      setStep(1);
      return;
    }
  };

  const clearError = () => {
    setError(null);
  };

  return (
    <div className="relative flex h-full min-h-svh flex-col items-center justify-center bg-muted p-6 md:p-10">
      <div className="absolute top-4 right-4">
        <ModeToggle />
      </div>
      <div className="w-full max-w-sm md:max-w-3xl">
        {step === 1 && (
          <AuthForm
            bottomText={
              <>
                Go back?{" "}
                <Link to="/login" className="underline underline-offset-4">
                  Log in
                </Link>
              </>
            }
            buttonText="Send Code"
            error={error}
            fields={[{ type: "email", label: "Email", id: "email" }]}
            onChange={clearError}
            onSubmit={async (formData) => {
              handleSendCode(formData);
            }}
            subtitle="Enter your email to receive a reset code"
            title="Reset"
          />
        )}

        {step === 2 && (
          <AuthForm
            bottomText={
              <>
                Go back?{" "}
                <Link to="/login" className="underline underline-offset-4">
                  Log in
                </Link>
              </>
            }
            buttonText="Reset Password"
            error={error}
            fields={[
              { type: "text", label: "Reset Code", id: "verificationCode" },
              { type: "password", label: "New Password", id: "password" },
              { type: "password", label: "Confirm Password", id: "confirm" },
            ]}
            onChange={clearError}
            onSubmit={async (formData) => {
              handleResetPassword(formData);
            }}
            subtitle="Check your email for the code"
            title="Enter Reset Code"
          />
        )}

        {step === 3 && (
          <div className="flex flex-col gap-6">
            <Card className="overflow-hidden p-0">
              <CardContent className="grid p-0 md:grid-cols-2">
                <div className="flex flex-col gap-6 p-6 md:p-8 w-full">
                  <div className="flex flex-col items-center text-center">
                    <h1 className="text-2xl font-bold">Success!</h1>
                    <p className="text-muted-foreground text-balance">
                      Your password is reset.
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
          </div>
        )}
      </div>
    </div>
  );
}
