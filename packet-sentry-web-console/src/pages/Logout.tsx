import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { Link, useNavigate } from 'react-router-dom'
import { useAuth } from '@/contexts/AuthContext'
import reactLogo from '@/assets/react.svg'
import shadcnLogo from '@/assets/shadcn.svg'

export default function LogoutPage() {
  const navigate = useNavigate()
  const { recheckAuth } = useAuth()

  const handleLogout = async () => {
    localStorage.setItem('jwt', '')
    await recheckAuth()
    navigate('/')
  }
  return (
    <main className="w-full flex flex-col items-center gap-x-4 px-4 gap-t-4 pt-4 overflow-y-auto">
      <Card className="overflow-hidden p-0 w-3/4 h-full overflow-y-auto">
        <CardContent className="grid p-0 md:grid-cols-2">
          <div className="flex flex-col items-center justify-center gap-6 p-4 md:p-8">
            <div className="flex flex-col items-center text-center">
              <h1 className="text-2xl font-bold m-4">Logging Out</h1>
              <p className="text-muted-foreground text-balance">
                You are about to log out of your account.
              </p>
              <p className="text-muted-foreground text-balance">
                Are you sure you want to continue?
              </p>
            </div>
            <Button onClick={handleLogout} className="w-full">
              Logout
            </Button>

            <div className="text-center text-sm">
              Go back?{' '}
              <Link to="/home" className="underline underline-offset-4">
                Home
              </Link>
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
    </main>
  )
}
