import { BrowserRouter as Router, Routes, Route } from 'react-router-dom'
import { AuthProvider } from './contexts/AuthContext'
import { ThemeProvider } from './contexts/ThemeProvider'
import { AuthenticatedRoute } from '@/components/AuthenticatedRoute'
import { UnauthenticatedRoute } from './components/UnauthenticatedRoute'
import { Toaster } from '@/components/ui/sonner'

// Pages
import Home from './pages/Home'
import Login from './pages/Login'
import ForgotPassword from './pages/ForgotPassword'
import RootRedirect from './pages/RootRedirect'
import Signup from './pages/Signup'
import AuthenticatedLayout from './layouts/AuthenticatedLayout'
import DevicesList from './pages/DevicesList'
import NewDevice from './pages/NewDevice'
import Logout from './pages/Logout'
import Settings from './pages/Settings'
import Billing from './pages/Billing'
import ResetPassword from './pages/ResetPassword'
import AdministratorsList from './pages/AdministratorsList'

function App() {
  return (
    <ThemeProvider
      defaultTheme="dark"
      storageKey="packet-sentry-web-console-theme"
    >
      {/* Make `toast` function callable in entire app */}
      <Toaster />
      {/* Makes AuthContext and its `useAuth` hook available in entire app */}
      <AuthProvider>
        <Router>
          <Routes>
            {/* Handles redirecting to `/login` or `/home` for unauth'd and auth'd, respectively */}
            <Route path="/" element={<RootRedirect />} />

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

            {/* Wrap all authenticated routes with the sidebar + top nav layout */}
            <Route path="/" element={<AuthenticatedLayout />}>
              {/* Authenticated Routes (depend on AuthProvider) */}
              <Route
                path="administrators"
                element={
                  <AuthenticatedRoute>
                    <AdministratorsList />
                  </AuthenticatedRoute>
                }
              />
              <Route
                path="billing"
                element={
                  <AuthenticatedRoute>
                    <Billing />
                  </AuthenticatedRoute>
                }
              />
              <Route
                path="devices/list"
                element={
                  <AuthenticatedRoute>
                    <DevicesList />
                  </AuthenticatedRoute>
                }
              />
              <Route
                path="devices/new"
                element={
                  <AuthenticatedRoute>
                    <NewDevice />
                  </AuthenticatedRoute>
                }
              />
              <Route
                path="home"
                element={
                  <AuthenticatedRoute>
                    <Home />
                  </AuthenticatedRoute>
                }
              />
              <Route
                path="logout"
                element={
                  <AuthenticatedRoute>
                    <Logout />
                  </AuthenticatedRoute>
                }
              />
              <Route
                path="reset-password"
                element={
                  <AuthenticatedRoute>
                    <ResetPassword />
                  </AuthenticatedRoute>
                }
              />
              <Route
                path="settings"
                element={
                  <AuthenticatedRoute>
                    <Settings />
                  </AuthenticatedRoute>
                }
              />
            </Route>
          </Routes>
        </Router>
      </AuthProvider>
    </ThemeProvider>
  )
}

export default App
