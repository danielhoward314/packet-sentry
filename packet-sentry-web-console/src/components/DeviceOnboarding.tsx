import { useState, useEffect } from 'react'
import { Tabs, TabsList, TabsTrigger, TabsContent } from '@/components/ui/tabs'
import { Card, CardContent } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import {
  Dialog,
  DialogTrigger,
  DialogContent,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Skeleton } from '@/components/ui/skeleton'
import { ClipboardCopyIcon } from 'lucide-react'

export function DeviceOnboarding() {
  const [step, setStep] = useState('os')
  const [os, setOs] = useState<string | null>(null)
  const [deviceName, setDeviceName] = useState('')
  const [installKey, setInstallKey] = useState<string | null>(null)
  const [backendReady, setBackendReady] = useState(false)
  const [copied, setCopied] = useState(false)

  useEffect(() => {
    if (step === 'waiting') {
      const timer = setTimeout(() => setBackendReady(true), 3000)
      return () => clearTimeout(timer)
    }
  }, [step])

  const generateInstallKey = () => {
    const key = crypto.randomUUID()
    setInstallKey(key)
  }

  const copyToClipboard = async () => {
    if (installKey) {
      await navigator.clipboard.writeText(installKey)
      setCopied(true)
      setTimeout(() => setCopied(false), 1500)
    }
  }

  return (
    <Tabs
      value={step}
      onValueChange={setStep}
      className="w-full max-w-3xl mx-auto mt-10"
    >
      <TabsList className="grid grid-cols-5 w-full">
        <TabsTrigger value="os">OS</TabsTrigger>
        <TabsTrigger value="device" disabled={!os}>
          Name
        </TabsTrigger>
        <TabsTrigger value="key" disabled={!deviceName}>
          Key
        </TabsTrigger>
        <TabsTrigger value="summary" disabled={!installKey}>
          Summary
        </TabsTrigger>
        <TabsTrigger value="waiting" disabled={!installKey}>
          Finish
        </TabsTrigger>
      </TabsList>

      <TabsContent value="os">
        <Card>
          <CardContent className="p-6 grid gap-4">
            <h2 className="text-lg font-semibold">
              Download the Packet Sentry Agent
            </h2>
            <div className="flex gap-4">
              {['Windows', 'macOS', 'Linux'].map(platform => (
                <Button
                  key={platform}
                  onClick={() => {
                    setOs(platform)
                    setStep('device')
                  }}
                >
                  {platform}
                </Button>
              ))}
            </div>
          </CardContent>
        </Card>
      </TabsContent>

      <TabsContent value="device">
        <Card>
          <CardContent className="p-6 grid gap-4">
            <h2 className="text-lg font-semibold">Name your device</h2>
            <Input
              value={deviceName}
              onChange={e => setDeviceName(e.target.value)}
              placeholder="My-Laptop"
            />
            <Button
              onClick={() => setStep('key')}
              disabled={!deviceName.trim()}
            >
              Continue
            </Button>
          </CardContent>
        </Card>
      </TabsContent>

      <TabsContent value="key">
        <Card>
          <CardContent className="p-6 grid gap-4">
            <Dialog>
              <DialogTrigger asChild>
                <Button onClick={generateInstallKey}>
                  Generate Install Key
                </Button>
              </DialogTrigger>
              <DialogContent>
                <DialogHeader>
                  <DialogTitle>Install Key</DialogTitle>
                </DialogHeader>
                <div className="flex items-center gap-2">
                  <code className="bg-muted px-2 py-1 rounded text-sm">
                    {installKey || 'No key yet'}
                  </code>
                  <Button
                    variant="ghost"
                    size="icon"
                    onClick={copyToClipboard}
                    disabled={!installKey}
                  >
                    <ClipboardCopyIcon className="w-4 h-4" />
                  </Button>
                  {copied && (
                    <span className="text-sm text-green-600">Copied!</span>
                  )}
                </div>
                <Button onClick={() => setStep('summary')}>Continue</Button>
              </DialogContent>
            </Dialog>
          </CardContent>
        </Card>
      </TabsContent>

      <TabsContent value="summary">
        <Card>
          <CardContent className="p-6 grid gap-4">
            <h2 className="text-lg font-semibold">Summary</h2>
            <ul className="list-disc pl-5">
              <li>Download the installer for {os}.</li>
              <li>Run it on the device named "{deviceName}".</li>
              <li>When prompted, enter the install key.</li>
            </ul>
            <Button onClick={() => setStep('waiting')}>Finish</Button>
          </CardContent>
        </Card>
      </TabsContent>

      <TabsContent value="waiting">
        <Card>
          <CardContent className="p-6 grid gap-4">
            <h2 className="text-lg font-semibold">Waiting for backend...</h2>
            {!backendReady ? (
              <>
                <Skeleton className="h-4 w-3/4" />
                <Skeleton className="h-4 w-1/2" />
                <Skeleton className="h-4 w-1/3" />
              </>
            ) : (
              <Button>Installation Complete!</Button>
            )}
          </CardContent>
        </Card>
      </TabsContent>
    </Tabs>
  )
}
