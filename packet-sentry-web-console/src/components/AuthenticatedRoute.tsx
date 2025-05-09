import { useAuth } from "@/contexts/AuthContext";
import { JSX } from "react";
import { Navigate } from "react-router-dom";
import { Loader2 } from "lucide-react";

export const AuthenticatedRoute = ({ children }: { children: JSX.Element }) => {
  const { isAuthenticated, loading } = useAuth();

  if (loading) {
    return (
      <div className="flex justify-center items-center h-screen">
        <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
      </div>
    );
  }

  return isAuthenticated ? children : <Navigate to="/login" replace />;
};
