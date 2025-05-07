import { LOCALSTORAGE } from "@/lib/consts";
import { parseJwt } from "@/lib/jwt";
import { createContext, useContext, useEffect, useState } from "react";

interface AdminUser {
  id: string;
  authorizationRole: string;
  organizationId: string;
}

interface AdminUserContextType {
  adminUser: AdminUser | null;
  refreshAdminUser: () => void;
  setAdminUser: (adminUser: AdminUser | null) => void;
}

const AdminUserContext = createContext<AdminUserContextType | undefined>(
  undefined,
);

export function AdminUserProvider({ children }: { children: React.ReactNode }) {
  const [adminUser, setAdminUser] = useState<AdminUser | null>(null);
  const { ADMIN_UI_ACCESS_TOKEN } = LOCALSTORAGE;

  const refreshAdminUser = () => {
    const accessToken = localStorage.getItem(ADMIN_UI_ACCESS_TOKEN);
    if (!accessToken) return;

    try {
      const parsedTokenPayload = parseJwt(accessToken);
      const id = parsedTokenPayload?.sub ?? "";
      const authorizationRole = parsedTokenPayload?.authorization_role ?? "";
      const organizationId = parsedTokenPayload?.organization_id ?? "";
      setAdminUser({
        id,
        authorizationRole,
        organizationId,
      });
    } catch (e) {
      console.error("Failed to parse JWT:", e);
    }
  };

  useEffect(() => {
    refreshAdminUser();
  }, []);

  return (
    <AdminUserContext.Provider
      value={{ adminUser, refreshAdminUser, setAdminUser }}
    >
      {children}
    </AdminUserContext.Provider>
  );
}

export function useAdminUser() {
  const context = useContext(AdminUserContext);
  if (!context)
    throw new Error("useAdminUser must be used within AdminUserProvider");
  return context;
}
