import { useState } from 'react'
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import { ModeRadioForm } from '@/components/ModeRadioForm'
import { Button } from '@/components/ui/button'
import { Skeleton } from '@/components/ui/skeleton'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog'
import { Label } from '@/components/ui/label'
import { zodResolver } from '@hookform/resolvers/zod'
import { useForm } from 'react-hook-form'
import { useTheme } from '@/contexts/ThemeProvider'
import { z, ZodTypeAny } from 'zod'

type Field =
  | { type: 'text'; label: string; id: string }
  | { type: 'email'; label: string; id: string }
  | {
      type: 'select'
      label: string
      id: string
      options: { value: string; label: string }[]
    }

interface SettingsDetailsProps
  extends Omit<React.ComponentProps<'div'>, 'onSubmit'> {
  fields: Field[]
  onSubmit?: (formData: FormData) => void | Promise<void>
}

function buildZodSchema(fields: Field[]) {
  const shape: Record<string, ZodTypeAny> = {}

  for (const field of fields) {
    let schema = z.string().min(1, { message: `${field.label} is required.` })

    if (field.type === 'email') {
      schema = schema.email({ message: 'Invalid email address.' })
    }

    shape[field.id] = schema
  }

  shape['theme'] = z.enum(['light', 'dark'], {
    required_error: 'Please select a theme.',
  })

  return z.object(shape)
}

export function SettingsDetails({ fields, onSubmit }: SettingsDetailsProps) {
  const [error, setError] = useState<string | null>(null)
  const [openSettingsFormDialog, setOpenSettingsFormDialog] = useState(false)
  const { theme } = useTheme()
  const formSchema = buildZodSchema(fields)

  const form = useForm<z.infer<typeof formSchema>>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      ...Object.fromEntries(fields.map(f => [f.id, ''])),
      theme: 'light',
    },
  })

  const handleSubmit = (values: z.infer<typeof formSchema>) => {
    const formData = new FormData()
    for (const key in values) {
      formData.append(key, values[key])
    }
    onSubmit?.(formData)
    setOpenSettingsFormDialog(false)
    form.reset()
  }

  const themeLabel = theme === 'light' ? 'Light' : 'Dark'

  const clearError = () => {
    setError(null)
  }

  return (
    <>
      <Label className="text-lg font-bold">Full Name</Label>
      <p className="text-muted-foreground text-balance">TODO full name here</p>
      <Label className="text-lg font-bold">Email</Label>
      <p className="text-muted-foreground text-balance">TODO email here</p>
      <Label className="text-lg font-bold">Appearance</Label>
      <div className="flex justify-between">
        <div
          className={`${theme === 'light' ? 'force-light' : 'dark'} w-[300px] h-[100px] rounded border p-4 transition-all bg-background text-foreground [&_.skeleton]:bg-muted`}
        >
          <Label className="pb-2">{themeLabel}</Label>
          <div className="flex items-center gap-4">
            <Skeleton className="h-10 w-10 rounded-full skeleton" />
            <div className="flex flex-col gap-2 flex-1">
              <Skeleton className="h-4 w-3/4 skeleton" />
              <Skeleton className="h-4 w-1/2 skeleton" />
            </div>
          </div>
        </div>
        <Dialog
          open={openSettingsFormDialog}
          onOpenChange={setOpenSettingsFormDialog}
        >
          <DialogTrigger asChild>
            <Button variant="outline" className="self-end">
              Edit
            </Button>
          </DialogTrigger>
          <DialogContent className="sm:max-w-[500px]">
            <DialogHeader>
              <DialogTitle>Settings</DialogTitle>
              <DialogDescription>
                Make changes to the settings for your administrator user.
              </DialogDescription>
            </DialogHeader>
            <Form {...form}>
              <form
                className="w-full my-4"
                onSubmit={form.handleSubmit(handleSubmit)}
              >
                {fields.map(field => {
                  if (field.id === 'theme') return
                  return (
                    <FormField
                      key={field.id}
                      control={form.control}
                      name={field.id}
                      render={({ field: rhfField }) => (
                        <FormItem>
                          <div className="flex items-center">
                            <FormLabel className="text-medium m-2">
                              {field.label}
                            </FormLabel>
                          </div>
                          <FormControl>
                            <Input
                              className="w-full m-0"
                              type={field.type}
                              {...rhfField}
                              onChange={e => {
                                rhfField.onChange(e)
                                clearError()
                              }}
                            />
                          </FormControl>
                          <FormMessage />
                        </FormItem>
                      )}
                    />
                  )
                })}

                <FormField
                  control={form.control}
                  name="theme"
                  render={({ field }) => (
                    <FormItem className="my-4">
                      <FormLabel>Theme Preference</FormLabel>
                      <ModeRadioForm
                        field={{
                          value: field.value,
                          onChange: field.onChange,
                        }}
                        clearError={clearError}
                      />
                      <FormMessage />
                    </FormItem>
                  )}
                />

                {error && (
                  <FormMessage className="text-destructive text-sm text-center">
                    {error}
                  </FormMessage>
                )}

                <DialogFooter className="my-2">
                  <Button className="my-2" type="submit">
                    Save Changes
                  </Button>
                </DialogFooter>
              </form>
            </Form>
          </DialogContent>
        </Dialog>
      </div>
    </>
  )
}
