import { Button } from "@/components/ui/button";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Label } from "@/components/ui/label";
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
import { GetAdministratorResponse } from "@/types/api";

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

interface UpdateAdministratorProps {
  existingAdmin: GetAdministratorResponse;
  fields: Field[];
  onSubmit?: (formData: FormData) => void | Promise<void>;
}

export function UpdateAdministratorForm({
  existingAdmin,
  fields,
  onSubmit,
}: UpdateAdministratorProps) {
  const [error, setError] = useState<string | null>(null);
  const [openUpdateAdminFormDialog, setOpenUpdateAdminFormDialog] =
    useState(false);

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
    setOpenUpdateAdminFormDialog(false);
  };

  const clearError = () => {
    setError(null);
  };

  const mapAuthorizationRole = (dbRole: string) => {
    if (dbRole === "PRIMARY_ADMIN") return "Primary Administrator";
    if (dbRole === "SECONDARY_ADMIN") return "Secondary Administrator";
    return "Unknown Role";
  };

  return (
    <>
      <Label className="text-lg font-bold">Full Name</Label>
      <p className="text-muted-foreground text-balance">
        {existingAdmin?.displayName ?? ""}
      </p>
      <Label className="text-lg font-bold">Email</Label>
      <p className="text-muted-foreground text-balance">
        {existingAdmin?.email ?? ""}
      </p>
      <Label className="text-lg font-bold">Authorization Role</Label>
      <p className="text-muted-foreground text-balance">
        {mapAuthorizationRole(existingAdmin?.authorizationRole)}
      </p>
      <Dialog
        open={openUpdateAdminFormDialog}
        onOpenChange={setOpenUpdateAdminFormDialog}
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
                                  <SelectItem
                                    key={role.value}
                                    value={role.value}
                                  >
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
              <DialogFooter className="my-2">
                <Button className="my-2" type="submit">
                  Save Changes
                </Button>
              </DialogFooter>
            </form>
          </Form>
        </DialogContent>
      </Dialog>
    </>
  );
}
