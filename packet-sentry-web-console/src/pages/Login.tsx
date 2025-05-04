import { AuthForm } from '@/components/AuthForm'
import { ModeToggle } from '@/components/ModeToggle'
import { useAuth } from '@/contexts/AuthContext'
import { Link, useNavigate } from 'react-router-dom'

export default function LoginPage() {
  const navigate = useNavigate()
  const { recheckAuth } = useAuth()

  const handleLogin = async (formData: FormData) => {
    const email = formData.get('email') as string
    const password = formData.get('password') as string
    if (email === 'a@b.com' && password === 'abc') {
      localStorage.setItem('jwt', 'valid-token')
      await recheckAuth()
      navigate('/home')
      return
    }
  }

  return (
    <div className="relative flex min-h-svh flex-col items-center justify-center bg-muted p-6 md:p-10">
      <div className="absolute top-4 right-4">
        <ModeToggle />
      </div>

      <div className="w-full max-w-sm md:max-w-3xl">
        <AuthForm
          bottomText={
            <>
              Don&apos;t have an account?{' '}
              <Link to="/signup" className="underline underline-offset-4">
                Sign up
              </Link>
            </>
          }
          buttonText="Login"
          fields={[
            { type: 'email', label: 'Email', id: 'email' },
            { type: 'password', label: 'Password', id: 'password' },
          ]}
          onSubmit={async formData => {
            handleLogin(formData)
          }}
          subtitle="Login to your Packet Sentry account"
          showForgotPassword
          title="Welcome Back"
        />
      </div>
    </div>
  )
}
