import { useAuth } from '@/contexts/AuthContext'
import { JSX } from 'react'
import { Navigate } from 'react-router-dom'

export const UnauthenticatedRoute = ({
  children,
}: {
  children: JSX.Element
}) => {
  const { isAuthenticated, loading } = useAuth()

  if (loading) {
    return <div className="text-center mt-10">Loading...</div> // or a spinner
  }

  return !isAuthenticated ? children : <Navigate to="/home" replace />
}
