import { createContext, useContext, useEffect, useState } from "react";
import { LOCALSTORAGE } from "@/lib/consts";
import { useEnv } from "./EnvContext";
import { parseJwt } from "@/lib/jwt";
import {
  ClaimsType,
  RefreshTokenRequest,
  ValidateSessionRequest,
} from "@/types/api";
import {
  isRefreshTokenResponse,
  isValidateSessionResponse,
} from "@/lib/apiTypeGuards";

interface AuthContextType {
  isAuthenticated: boolean;
  loading: boolean;
  setIsAuthenticated: (val: boolean) => void;
  recheckAuth: () => Promise<void>;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const { ADMIN_UI_ACCESS_TOKEN, ADMIN_UI_REFRESH_TOKEN } = LOCALSTORAGE;
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [loading, setLoading] = useState(true);
  const { API_BASE_URL } = useEnv();

  const checkAuth = async () => {
    const accessToken = localStorage.getItem(ADMIN_UI_ACCESS_TOKEN);
    const refreshToken = localStorage.getItem(ADMIN_UI_REFRESH_TOKEN);

    if (!accessToken) {
      setIsAuthenticated(false);
      setLoading(false);
      return;
    }

    try {
      const sessionResponse = await fetch(`${API_BASE_URL}/v1/session`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ jwt: accessToken } as ValidateSessionRequest),
      });

      if (sessionResponse.status === 200) {
        const data: unknown = await sessionResponse.json();

        if (!isValidateSessionResponse(data) || data.jwt !== accessToken) {
          throw new Error(
            "admin UI access token from localStorage does not match API response",
          );
        }

        const parsedTokenPayload = parseJwt(accessToken);
        if (!parsedTokenPayload?.authorization_role) {
          throw new Error("no authorization role in login jwt");
        }
        setIsAuthenticated(true);
      } else if (sessionResponse.status === 401 && refreshToken) {
        const refreshResponse = await fetch(`${API_BASE_URL}/v1/refresh`, {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({
            jwt: refreshToken,
            claimsType: ClaimsType.ADMIN_UI_SESSION,
          } as RefreshTokenRequest),
        });

        const refreshData: unknown = await refreshResponse.json();
        if (!isRefreshTokenResponse(refreshData)) {
          throw new Error("invalid refresh token response");
        }

        localStorage.setItem(ADMIN_UI_ACCESS_TOKEN, refreshData.jwt);
        setIsAuthenticated(true);
      }
    } catch (e) {
      console.error("session or refresh failed: ", e);
      localStorage.removeItem(ADMIN_UI_ACCESS_TOKEN);
      localStorage.removeItem(ADMIN_UI_REFRESH_TOKEN);
      setIsAuthenticated(false);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    checkAuth();
  }, []);

  return (
    <AuthContext.Provider
      value={{
        isAuthenticated,
        loading,
        setIsAuthenticated,
        recheckAuth: checkAuth,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (!context) throw new Error("useAuth must be used within AuthProvider");
  return context;
}
