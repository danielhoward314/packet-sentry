import { AuthForm } from "@/components/AuthForm";
import { ModeToggle } from "@/components/ModeToggle";
import { useAuth } from "@/contexts/AuthContext";
import { Link, useNavigate } from "react-router-dom";
import { LOCALSTORAGE } from "@/lib/consts";
import { useEnv } from "@/contexts/EnvContext";
import { toast } from "sonner";
import { LoginRequest } from "@/types/api";
import { isLoginResponse } from "@/lib/apiTypeGuards";

export default function LoginPage() {
  const {
    ADMIN_UI_ACCESS_TOKEN,
    ADMIN_UI_REFRESH_TOKEN,
    API_ACCESS_TOKEN,
    API_REFRESH_TOKEN,
  } = LOCALSTORAGE;
  const { API_BASE_URL } = useEnv();
  const navigate = useNavigate();
  const { recheckAuth } = useAuth();

  const handleLogin = async (formData: FormData) => {
    const email = formData.get("email") as string;
    const password = formData.get("password") as string;

    const body: LoginRequest = { email, password };

    try {
      const response = await fetch(`${API_BASE_URL}/v1/login`, {
        headers: { "Content-Type": "application/json" },
        method: "POST",
        mode: "cors",
        body: JSON.stringify(body),
      });

      const data: unknown = await response.json();

      if (!isLoginResponse(data)) {
        throw new Error("login response is invalid");
      }

      localStorage.setItem(ADMIN_UI_ACCESS_TOKEN, data.adminUiAccessToken);
      localStorage.setItem(ADMIN_UI_REFRESH_TOKEN, data.adminUiRefreshToken);
      localStorage.setItem(API_ACCESS_TOKEN, data.apiAccessToken);
      localStorage.setItem(API_REFRESH_TOKEN, data.apiRefreshToken);

      await recheckAuth();
      navigate("/home");
    } catch (e) {
      console.error("failed to log in: ", e);
      toast.error("Failed to log in.");
    }
  };

  return (
    <div className="relative flex min-h-svh flex-col items-center justify-center bg-muted p-6 md:p-10">
      <div className="absolute top-4 right-4">
        <ModeToggle />
      </div>

      <div className="w-full max-w-sm md:max-w-3xl">
        <AuthForm
          bottomText={
            <>
              Don&apos;t have an account?{" "}
              <Link to="/signup" className="underline underline-offset-4">
                Sign up
              </Link>
            </>
          }
          buttonText="Login"
          fields={[
            { type: "email", label: "Email", id: "email" },
            { type: "password", label: "Password", id: "password" },
          ]}
          onSubmit={async (formData) => {
            handleLogin(formData);
          }}
          subtitle="Login to your Packet Sentry account"
          showForgotPassword
          title="Welcome Back"
        />
      </div>
    </div>
  );
}
