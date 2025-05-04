import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Separator } from '@/components/ui/separator'
import { ReactNode } from 'react'

interface MainContentCardLayoutProps {
  cardDescription: string
  cardTitle: string
  children: ReactNode
}

export default function MainContentCardLayout({
  cardDescription,
  cardTitle,
  children,
}: MainContentCardLayoutProps) {
  return (
    <main className="w-full h-full flex flex-col items-center gap-x-4 px-4 gap-t-4 pt-4 overflow-y-auto">
      <Card className="w-3/4 h-full pt-6 pb-0 overflow-y-auto">
        <CardHeader className="w-full">
          <CardTitle className="text-2xl font-bold">{cardTitle}</CardTitle>
          <CardDescription>{cardDescription}</CardDescription>
          <Separator />
        </CardHeader>
        <CardContent className="w-full flex flex-col gap-4">
          {children}
        </CardContent>
      </Card>
    </main>
  )
}
