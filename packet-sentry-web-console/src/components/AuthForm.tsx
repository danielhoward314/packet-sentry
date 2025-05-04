'use client'

import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { zodResolver } from '@hookform/resolvers/zod'
import { useForm } from 'react-hook-form'
import { z, ZodTypeAny } from 'zod'
import { cn } from '@/lib/utils'
import { Link } from 'react-router-dom'
import reactLogo from '@/assets/react.svg'
import shadcnLogo from '@/assets/shadcn.svg'

type Field =
  | { type: 'email'; label: string; id: string }
  | { type: 'password'; label: string; id: string }
  | { type: 'text'; label: string; id: string }

interface AuthFormProps extends Omit<React.ComponentProps<'div'>, 'onSubmit'> {
  buttonText: string
  bottomText?: React.ReactNode
  error?: string | null
  fields: Field[]
  inMainLayout?: boolean
  onSubmit?: (formData: FormData) => void | Promise<void>
  onChange?: () => void
  showForgotPassword?: boolean
  subtitle: string
  title: string
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

  return z.object(shape)
}

export function AuthForm({
  buttonText,
  bottomText,
  error,
  fields,
  inMainLayout,
  onSubmit,
  onChange,
  showForgotPassword = false,
  subtitle,
  title,
}: AuthFormProps) {
  const formSchema = buildZodSchema(fields)

  const form = useForm<z.infer<typeof formSchema>>({
    resolver: zodResolver(formSchema),
    defaultValues: Object.fromEntries(fields.map(f => [f.id, ''])),
  })

  const handleSubmit = (values: z.infer<typeof formSchema>) => {
    // Convert to FormData-like object for compatibility with your onSubmit API
    const formData = new FormData()
    for (const key in values) {
      formData.append(key, values[key])
    }

    onSubmit?.(formData)
  }

  return (
    <Card
      className={cn(
        'overflow-hidden p-0',
        inMainLayout && 'w-3/4 h-full overflow-y-auto'
      )}
    >
      <CardContent
        className={cn('grid p-0 md:grid-cols-2', inMainLayout && 'w-full')}
      >
        <Form {...form}>
          <form
            onSubmit={form.handleSubmit(handleSubmit)}
            className="p-6 md:p-8 w-full"
          >
            <div className="flex flex-col gap-6">
              <div className="flex flex-col items-center text-center">
                <h1 className="text-2xl font-bold">{title}</h1>
                <p className="text-muted-foreground text-balance">{subtitle}</p>
              </div>

              {fields.map(field => (
                <FormField
                  key={field.id}
                  control={form.control}
                  name={field.id}
                  render={({ field: rhfField }) => (
                    <FormItem>
                      <div className="flex items-center">
                        <FormLabel>{field.label}</FormLabel>
                        {field.type === 'password' && showForgotPassword && (
                          <Link
                            to="/forgot-password"
                            className="ml-auto text-sm underline-offset-2 hover:underline"
                          >
                            Forgot your password?
                          </Link>
                        )}
                      </div>
                      <FormControl>
                        <Input
                          type={field.type}
                          {...rhfField}
                          onChange={e => {
                            rhfField.onChange(e)
                            onChange?.()
                          }}
                        />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
              ))}

              {error && (
                <FormMessage className="text-destructive text-sm text-center">
                  {error}
                </FormMessage>
              )}

              <Button type="submit" className="w-full">
                {buttonText}
              </Button>

              {bottomText && (
                <div className="text-center text-sm">{bottomText}</div>
              )}
            </div>
          </form>
        </Form>
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
  )
}
