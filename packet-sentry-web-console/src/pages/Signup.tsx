import { useState } from 'react'
import { AuthForm } from '@/components/AuthForm'
import { ModeToggle } from '@/components/ModeToggle'
import { Link } from 'react-router-dom'
import { useNavigate } from 'react-router-dom'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import reactLogo from '@/assets/react.svg'
import shadcnLogo from '@/assets/shadcn.svg'

export default function SignupPage() {
  const navigate = useNavigate()
  const [step, setStep] = useState<1 | 2 | 3>(1)

  const handleSignup = async (formData: FormData) => {
    const name = formData.get('name') as string
    const organization = formData.get('organization') as string
    const email = formData.get('email') as string
    const password = formData.get('password') as string
    console.log(name, organization, email, password)
    if (email === 'a@b.com' && password === 'abc') {
      setStep(2)
    }
  }

  const handleVerify = async (formData: FormData) => {
    const verificationCode = formData.get('verificationCode') as string
    if (verificationCode === '123456') {
      setStep(3)
    }
  }

  return (
    <div className="relative flex min-h-svh flex-col items-center justify-center bg-muted p-6 md:p-10">
      <div className="absolute top-4 right-4">
        <ModeToggle />
      </div>

      <div className="w-full max-w-sm md:max-w-3xl">
        {step === 1 && (
          <AuthForm
            bottomText={
              <>
                Already have an account?{' '}
                <Link to="/login" className="underline underline-offset-4">
                  Log in
                </Link>
              </>
            }
            buttonText="Sign up"
            fields={[
              { type: 'text', label: 'Full Name', id: 'name' },
              { type: 'text', label: 'Organization', id: 'organization' },
              { type: 'email', label: 'Email', id: 'email' },
              { type: 'password', label: 'Password', id: 'password' },
            ]}
            onSubmit={async formData => {
              handleSignup(formData)
            }}
            subtitle="Create your Packet Sentry account"
            title="Welcome"
          />
        )}
        {step === 2 && (
          <AuthForm
            bottomText={
              <>
                Start over?{' '}
                <Link
                  to="/signup"
                  className="underline underline-offset-4"
                  onClick={() => {
                    setStep(1)
                  }}
                >
                  Sign up
                </Link>
              </>
            }
            buttonText="Submit"
            fields={[
              {
                type: 'text',
                label: 'Email Verification Code',
                id: 'verificationCode',
              },
            ]}
            onSubmit={async formData => {
              handleVerify(formData)
            }}
            subtitle="Enter the verification code sent to your email."
            title="Verify"
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
                      Your account has been created.
                    </p>
                    <Button
                      onClick={() => navigate('/login')}
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
  )
}
