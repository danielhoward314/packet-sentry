import { useState } from "react";
import { Label } from "@/components/ui/label";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
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
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from "@/components/ui/tooltip";
import { zodResolver } from "@hookform/resolvers/zod";
import { useForm } from "react-hook-form";
import { z, ZodTypeAny } from "zod";
import { CircleHelp } from "lucide-react";
import { GetOrganizationResponse } from "@/types/api";

type BillingPlanField = {
  type: "select";
  label: string;
  id: string;
  options: { value: string; label: string }[];
  default: string;
};

type CreditCardField =
  | { type: "text"; label: string; id: string; options: [] }
  | { type: "creditCardNumber"; label: string; id: string; options: [] }
  | { type: "cvc"; label: string; id: string; options: [] }
  | {
      type: "select";
      label: string;
      id: string;
      options: { value: string; label: string }[];
    };
interface BillingDetailsProps
  extends Omit<React.ComponentProps<"div">, "onSubmit"> {
  existingOrganization: GetOrganizationResponse;
  onSubmit?: (formName: string, formData: FormData) => void | Promise<void>;
}

function buildBillingPlanSchema() {
  const shape: Record<string, ZodTypeAny> = {};
  shape["billingPlan"] = z.enum(
    ["10_DEVICES_99_MONTH", "50_DEVICES_399_MONTH", "100_DEVICES_799_MONTH"],
    {
      required_error: "Please select a billing plan.",
    },
  );

  return z.object(shape);
}

function buildCreditCardSchema(fields: CreditCardField[]) {
  const shape: Record<string, ZodTypeAny> = {};

  for (const field of fields) {
    let schema = z.string().min(1, { message: `${field.label} is required.` });

    if (field.type === "creditCardNumber") {
      schema = schema.regex(/^\d{13,19}$/, {
        message: "Invalid credit card number.",
      });
    }

    if (field.type === "cvc") {
      schema = schema.regex(/^\d{3,4}$/, { message: "Invalid CVC code." });
    }

    shape[field.id] = schema;
  }

  // Return object schema with joint expiration date validation
  return z.object(shape).superRefine((data, ctx) => {
    const month = parseInt(data.expirationMonth, 10);
    const year = parseInt(data.expirationYear, 10);
    const now = new Date();
    const currentMonth = now.getMonth() + 1;
    const currentYear = now.getFullYear();

    if (
      isNaN(month) ||
      isNaN(year) ||
      year < currentYear ||
      (year === currentYear && month < currentMonth)
    ) {
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        message: "Expiration date must be in the future.",
        path: ["expirationMonth"],
      });
      ctx.addIssue({
        code: z.ZodIssueCode.custom,
        message: "Expiration date must be in the future.",
        path: ["expirationYear"],
      });
    }
  });
}

export function BillingDetails({
  existingOrganization,
  onSubmit,
}: BillingDetailsProps) {
  const billingPlanField: BillingPlanField = {
    type: "select",
    label: "Billing Plan",
    id: "billingPlan",
    options: [
      { value: "10_DEVICES_99_MONTH", label: "10 Devices at $99/month" },
      { value: "50_DEVICES_399_MONTH", label: "50 Devices at $399/month" },
      { value: "100_DEVICES_799_MONTH", label: "100 Devices at $799/month" },
    ],
    default: "10_DEVICES_99_MONTH",
  };
  const creditCardFields: CreditCardField[] = [
    { type: "text", label: "Name on Card", id: "cardHolderName", options: [] },
    {
      type: "creditCardNumber",
      label: "Card Number",
      id: "cardNumber",
      options: [],
    },
    { type: "cvc", label: "CVC", id: "cvc", options: [] },
    {
      type: "select",
      label: "Expiration Month",
      id: "expirationMonth",
      options: Array.from({ length: 12 }, (_, i) => ({
        value: String(i + 1).padStart(2, "0"),
        label: String(i + 1).padStart(2, "0"),
      })),
    },
    {
      type: "select",
      label: "Expiration Year",
      id: "expirationYear",
      options: Array.from({ length: 10 }, (_, i) => {
        const year = new Date().getFullYear() + i;
        return { value: String(year), label: String(year) };
      }),
    },
  ];

  const [error, setError] = useState<string | null>(null);
  const [openBillingPlanDialog, setOpenBillingPlanDialog] = useState(false);
  const [openPaymentMethodDialog, setOpenPaymentMethodDialog] = useState(false);
  const billingFormSchema = buildBillingPlanSchema();
  const creditCardFormSchema = buildCreditCardSchema(creditCardFields);

  const billingForm = useForm<z.infer<typeof billingFormSchema>>({
    resolver: zodResolver(billingFormSchema),
    defaultValues: {
      billingPlan: "10_DEVICES_99_MONTH",
    },
  });

  const creditCardForm = useForm<z.infer<typeof creditCardFormSchema>>({
    resolver: zodResolver(creditCardFormSchema),
    defaultValues: {
      ...Object.fromEntries(creditCardFields.map((f) => [f.id, ""])),
    },
  });

  const handleBillingFormSave = (values: z.infer<typeof billingFormSchema>) => {
    const formData = new FormData();
    for (const key in values) {
      console.log(key, values[key]);
      formData.append(key, values[key]);
    }

    onSubmit?.("billingPlanForm", formData);
    setOpenBillingPlanDialog(false);
    billingForm.reset();
  };

  const handleCreditCardFormSave = (
    values: z.infer<typeof creditCardFormSchema>,
  ) => {
    const formData = new FormData();
    for (const key in values) {
      console.log(key, values[key]);
      formData.append(key, values[key]);
    }

    onSubmit?.("creditCardForm", formData);
    setOpenPaymentMethodDialog(false);
    creditCardForm.reset();
  };

  const clearError = () => {
    setError(null);
  };

  const mapBillingPlan = (dbPlan: string) => {
    if (dbPlan === "10_DEVICES_99_MONTH") return "10 Devices / $99 Monthly";
    if (dbPlan === "50_DEVICES_399_MONTH") return "50 Devices / $399 Monthly";
    if (dbPlan === "100_DEVICES_799_MONTH") return "100 Devices / $799 Monthly";
    return "Unknown Plan";
  };

  return (
    <>
      <Label className="text-lg font-bold">Account Number</Label>
      <p className="text-muted-foreground text-balance">
        {existingOrganization.id}
      </p>
      <Label className="text-lg font-bold">Billing Plan</Label>
      <div className="flex justify-between">
        <p className="text-muted-foreground text-balance">
          {mapBillingPlan(existingOrganization.billingPlan)}
        </p>
        <Dialog
          open={openBillingPlanDialog}
          onOpenChange={setOpenBillingPlanDialog}
        >
          <DialogTrigger asChild>
            <Button variant="outline">Edit</Button>
          </DialogTrigger>
          <DialogContent className="sm:max-w-[500px]">
            <DialogHeader>
              <DialogTitle>Edit Billing Plan</DialogTitle>
              <DialogDescription>
                Upgrade billing plan (contact support@packetsentry.com for
                downgrades).
              </DialogDescription>
            </DialogHeader>
            <Form {...billingForm}>
              <form
                className="w-full my-4"
                onSubmit={billingForm.handleSubmit(handleBillingFormSave)}
              >
                <FormField
                  key={billingPlanField.id}
                  control={billingForm.control}
                  name={billingPlanField.id}
                  render={({ field: rhfField }) => (
                    <FormItem>
                      <FormLabel className="text-medium m-2">
                        {billingPlanField.label}
                      </FormLabel>
                      <FormControl>
                        <Select
                          value={rhfField.value}
                          onValueChange={rhfField.onChange}
                          defaultValue={billingPlanField.default}
                        >
                          <SelectTrigger
                            className="min-w-48"
                            id={billingPlanField.id}
                            aria-label={billingPlanField.label}
                          >
                            <SelectValue placeholder={billingPlanField.label} />
                          </SelectTrigger>
                          <SelectContent>
                            {billingPlanField.options.map((plan) => (
                              <SelectItem key={plan.value} value={plan.value}>
                                {plan.label}
                              </SelectItem>
                            ))}
                          </SelectContent>
                        </Select>
                      </FormControl>
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
                  <Button type="submit">Save Changes</Button>
                </DialogFooter>
              </form>
            </Form>
          </DialogContent>
        </Dialog>
      </div>
      <Label className="text-lg font-bold">Payment Method</Label>
      <div className="flex justify-between">
        <p className="text-muted-foreground text-balance">
          TODO last 4 digits of card
        </p>
        <Dialog
          open={openPaymentMethodDialog}
          onOpenChange={setOpenPaymentMethodDialog}
        >
          <DialogTrigger asChild>
            <Button variant="outline">Edit</Button>
          </DialogTrigger>
          <DialogContent className="sm:max-w-[500px]">
            <DialogHeader>
              <DialogTitle>Edit Payment Method</DialogTitle>
              <DialogDescription>
                Make changes to the payment method used for this account.
              </DialogDescription>
            </DialogHeader>
            <Form {...creditCardForm}>
              <form
                className="w-full my-4"
                onSubmit={creditCardForm.handleSubmit(handleCreditCardFormSave)}
              >
                <FormField
                  key={creditCardFields[0].id}
                  control={creditCardForm.control}
                  name={creditCardFields[0].id}
                  render={({ field: rhfField }) => (
                    <FormItem>
                      <FormLabel className="text-medium m-2">
                        {creditCardFields[0].label}
                      </FormLabel>
                      <FormControl>
                        <Input
                          className="max-w-[300px] m-0"
                          type={creditCardFields[0].type}
                          {...rhfField}
                          onChange={(e) => {
                            rhfField.onChange(e);
                            clearError();
                          }}
                        />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />

                <FormField
                  key={creditCardFields[1].id}
                  control={creditCardForm.control}
                  name={creditCardFields[1].id}
                  render={({ field: rhfField }) => (
                    <FormItem>
                      <div className="flex items-center">
                        <FormLabel className="text-medium m-2">
                          {creditCardFields[1].label}
                        </FormLabel>
                      </div>
                      <FormControl>
                        <Input
                          className="max-w-[300px] m-0"
                          type={creditCardFields[1].type}
                          {...rhfField}
                          onChange={(e) => {
                            rhfField.onChange(e);
                            clearError();
                          }}
                        />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />

                <div className="flex gap-4 my-4">
                  <FormField
                    key={creditCardFields[2].id}
                    control={creditCardForm.control}
                    name={creditCardFields[2].id}
                    render={({ field: rhfField }) => (
                      <FormItem>
                        <div className="flex items-center gap-2">
                          <FormLabel className="text-medium m-2">
                            {creditCardFields[2].label}
                          </FormLabel>
                          <TooltipProvider>
                            <Tooltip>
                              <TooltipTrigger>
                                <CircleHelp
                                  size={16}
                                  className="text-muted-foreground mb-[1px]"
                                />
                              </TooltipTrigger>
                              <TooltipContent className="max-w-48">
                                <p className="text-wrap">
                                  A security code printed on credit or debit
                                  cards designed to verify online transactions.
                                </p>
                              </TooltipContent>
                            </Tooltip>
                          </TooltipProvider>
                        </div>

                        <FormControl>
                          <Input
                            className="w-24 m-0"
                            type={creditCardFields[2].type}
                            {...rhfField}
                            onChange={(e) => {
                              rhfField.onChange(e);
                              clearError();
                            }}
                          />
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                  <FormField
                    key={creditCardFields[3].id}
                    control={creditCardForm.control}
                    name={creditCardFields[3].id}
                    render={({ field: rhfField }) => (
                      <FormItem>
                        <FormLabel className="text-medium m-2">
                          {creditCardFields[3].label}
                        </FormLabel>
                        <FormControl>
                          <Select
                            value={rhfField.value}
                            onValueChange={rhfField.onChange}
                          >
                            <SelectTrigger
                              className="min-w-24"
                              id={creditCardFields[3].id}
                              aria-label={creditCardFields[3].label}
                            >
                              <SelectValue placeholder="01" />
                            </SelectTrigger>
                            <SelectContent>
                              {creditCardFields[3].options.map((option) => (
                                <SelectItem
                                  key={option.value}
                                  value={option.value}
                                >
                                  {option.label}
                                </SelectItem>
                              ))}
                            </SelectContent>
                          </Select>
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                  <FormField
                    key={creditCardFields[4].id}
                    control={creditCardForm.control}
                    name={creditCardFields[4].id}
                    render={({ field: rhfField }) => (
                      <FormItem>
                        <FormLabel className="text-medium m-2">
                          {creditCardFields[4].label}
                        </FormLabel>
                        <FormControl>
                          <Select
                            value={rhfField.value}
                            onValueChange={rhfField.onChange}
                          >
                            <SelectTrigger
                              className="min-w-24"
                              id={creditCardFields[4].id}
                              aria-label={creditCardFields[4].label}
                            >
                              <SelectValue
                                placeholder={
                                  creditCardFields[4].options[1].value
                                }
                              />
                            </SelectTrigger>
                            <SelectContent>
                              {creditCardFields[4].options.map((option) => (
                                <SelectItem
                                  key={option.value}
                                  value={option.value}
                                >
                                  {option.label}
                                </SelectItem>
                              ))}
                            </SelectContent>
                          </Select>
                        </FormControl>
                        <FormMessage />
                      </FormItem>
                    )}
                  />
                </div>

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
  );
}
