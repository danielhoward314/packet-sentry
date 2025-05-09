import { Button } from "@/components/ui/button";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import { Input } from "@/components/ui/input";
import { z, ZodTypeAny } from "zod";
import { useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { CheckCircle } from "lucide-react";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";

export type Field =
  | { type: "text"; label: string; id: string }
  | { type: "email"; label: string; id: string }
  | {
      type: "select";
      label: string;
      id: string;
      options: { value: string; label: string }[];
      default: string;
    };

function buildZodSchema(fields: Field[]) {
  const shape: Record<string, ZodTypeAny> = {};

  for (const field of fields) {
    let schema = z.string().min(1, { message: `${field.label} is required.` });

    if (field.type === "email") {
      schema = schema.email({ message: "Invalid email address." });
    }

    if (field.type === "select") {
      shape["authorizationRole"] = z.enum(
        ["PRIMARY_ADMIN", "SECONDARY_ADMIN"],
        {
          required_error:
            "Please select an authorization role for the new administrator.",
        },
      );
    }

    shape[field.id] = schema;
  }

  return z.object(shape);
}

interface NewAdministratorProps {
  step: 1 | 2;
  fields: Field[];
  onSubmit?: (formData: FormData) => void | Promise<void>;
}

export function NewAdministratorForm({
  step,
  fields,
  onSubmit,
}: NewAdministratorProps) {
  const [error, setError] = useState<string | null>(null);

  const formSchema = buildZodSchema(fields);

  const form = useForm<z.infer<typeof formSchema>>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      ...Object.fromEntries(fields.map((f) => [f.id, ""])),
    },
  });

  const handleSubmit = (values: z.infer<typeof formSchema>) => {
    const formData = new FormData();
    for (const key in values) {
      formData.append(key, values[key]);
    }
    onSubmit?.(formData);
  };

  const clearError = () => {
    setError(null);
  };

  if (step === 1) {
    return (
      <Form {...form}>
        <form
          className="w-full my-4"
          onSubmit={form.handleSubmit(handleSubmit)}
        >
          {fields.map((field) => {
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
                      {field.type === "select" ? (
                        <Select
                          value={rhfField.value}
                          onValueChange={rhfField.onChange}
                          defaultValue={field.default}
                        >
                          <SelectTrigger
                            className="min-w-48"
                            id={field.id}
                            aria-label={field.label}
                          >
                            <SelectValue placeholder={field.label} />
                          </SelectTrigger>
                          <SelectContent>
                            {field.options.map((role) => (
                              <SelectItem key={role.value} value={role.value}>
                                {role.label}
                              </SelectItem>
                            ))}
                          </SelectContent>
                        </Select>
                      ) : (
                        <Input
                          className="w-full m-0"
                          type={field.type}
                          {...rhfField}
                          onChange={(e) => {
                            rhfField.onChange(e);
                            clearError();
                          }}
                        />
                      )}
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
            );
          })}

          {error && (
            <FormMessage className="text-destructive text-sm text-center">
              {error}
            </FormMessage>
          )}
          <Button className="my-2" type="submit">
            Save Changes
          </Button>
        </form>
      </Form>
    );
  }

  return (
    <Alert>
      <CheckCircle className="h-4 w-4" />
      <AlertTitle>Success</AlertTitle>
      <AlertDescription>
        Your administrator has been created and an account activation email has
        been sent.
      </AlertDescription>
    </Alert>
  );
}
