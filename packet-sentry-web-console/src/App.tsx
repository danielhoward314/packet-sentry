import {
  BrowserRouter as Router,
  Routes,
  Route,
  Navigate,
} from "react-router-dom";
import { AuthProvider } from "./contexts/AuthContext";
import { EnvProvider } from "@/contexts/EnvContext";
import { ThemeProvider } from "./contexts/ThemeProvider";
import { AuthenticatedRoute } from "@/components/AuthenticatedRoute";
import { UnauthenticatedRoute } from "./components/UnauthenticatedRoute";
import { Toaster } from "@/components/ui/sonner";

// Pages
import Home from "./pages/Home";
import Login from "./pages/Login";
import ForgotPassword from "./pages/ForgotPassword";
import Signup from "./pages/Signup";
import AuthenticatedLayout from "./layouts/AuthenticatedLayout";
import DevicesList from "./pages/DevicesList";
import NewDevice from "./pages/NewDevice";
import Logout from "./pages/Logout";
import Settings from "./pages/Settings";
import Billing from "./pages/Billing";
import ResetPassword from "./pages/ResetPassword";
import AdministratorsList from "./pages/AdministratorsList";
import { AdminUserProvider } from "./contexts/AdminUserContext";
import NewAdministrator from "./pages/NewAdministrator";
import { AxiosInterceptorProvider } from "@/contexts/AxiosProvider";
import UpdateAdministrator from "./pages/UpdateAdministrator";
import ActivateAdministrator from "./pages/ActivateAdministrator";
import UpdateDevice from "./pages/UpdateDevice";

function App() {
  return (
    <EnvProvider>
      <ThemeProvider
        defaultTheme="dark"
        storageKey="packet-sentry-web-console-theme"
      >
        {/* Make `toast` function callable in entire app */}
        <Toaster />
        <AuthProvider>
          <Router>
            <AxiosInterceptorProvider />
            <Routes>
              <Route path="/" element={<Navigate to="/home" />} />

              {/* Unauthenticated Routes (depend on AuthProvider) */}
              <Route
                path="/forgot-password"
                element={
                  <UnauthenticatedRoute>
                    <ForgotPassword />
                  </UnauthenticatedRoute>
                }
              />
              <Route
                path="/login"
                element={
                  <UnauthenticatedRoute>
                    <Login />
                  </UnauthenticatedRoute>
                }
              />
              <Route
                path="/signup"
                element={
                  <UnauthenticatedRoute>
                    <Signup />
                  </UnauthenticatedRoute>
                }
              />

              <Route
                path="/activate"
                element={
                  <UnauthenticatedRoute>
                    <ActivateAdministrator />
                  </UnauthenticatedRoute>
                }
              />

              <Route
                path="/"
                element={
                  <AuthenticatedRoute>
                    <AdminUserProvider>
                      <AuthenticatedLayout />
                    </AdminUserProvider>
                  </AuthenticatedRoute>
                }
              >
                <Route path="home" element={<Home />} />
                <Route path="devices/list" element={<DevicesList />} />
                <Route path="devices/new" element={<NewDevice />} />
                <Route
                  path="devices/update/:id"
                  element={<UpdateDevice />}
                />
                <Route path="administrators" element={<AdministratorsList />} />
                <Route
                  path="administrators/new"
                  element={<NewAdministrator />}
                />
                <Route
                  path="administrators/update/:id"
                  element={<UpdateAdministrator />}
                />
                <Route path="billing" element={<Billing />} />
                <Route path="settings" element={<Settings />} />
                <Route path="reset-password" element={<ResetPassword />} />
                <Route path="logout" element={<Logout />} />
              </Route>
            </Routes>
          </Router>
        </AuthProvider>
      </ThemeProvider>
    </EnvProvider>
  );
}

export default App;
