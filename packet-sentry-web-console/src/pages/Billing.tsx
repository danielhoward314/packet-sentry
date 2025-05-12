"use client";

import { BillingDetails } from "@/components/BillingDetails";
import { useAdminUser } from "@/contexts/AdminUserContext";
import MainContentCardLayout from "@/layouts/MainContentCardLayout";
import { getOrganization, updateOrganization } from "@/lib/api";
import { GetOrganizationResponse } from "@/types/api";
import { useEffect, useState } from "react";
import { toast } from "sonner";
import { AlertCircle, Loader2 } from "lucide-react";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";

export default function BillingPage() {
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [existingOrg, setExistingOrg] =
    useState<GetOrganizationResponse | null>(null);
  const { adminUser, refreshAdminUser } = useAdminUser();

  useEffect(() => {
    if (!adminUser?.id) {
      refreshAdminUser();
      if (!adminUser?.id) {
        console.error("admin user not in context after refresh");
        return;
      }
    }

    const fetchData = async () => {
      try {
        const responseData = await getOrganization(adminUser.organizationId);
        setExistingOrg(responseData);
      } catch (err) {
        console.error(err);
        setError("failed to fetch organization data");
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  }, [adminUser?.id]);

  const handleBillingPageSave = async (
    formName: string,
    formData: FormData,
  ) => {
    if (!existingOrg?.id) return;

    const data = Object.fromEntries(formData.entries());

    let toastSuccessMsg = "";

    if (formName === "billingPlanForm") {
      try {
        const response = await updateOrganization(existingOrg.id, {
          billingPlan: data.billingPlan as string,
        });

        if (response.status < 200 || response.status >= 400) {
          throw new Error("non-200 status for update");
        }
      } catch (err) {
        console.error(err);
        setError("Failed to save update your billing plan.");
      }
      toastSuccessMsg = "Your billing plan has been updated.";
    } else if (formName === "creditCardForm") {
      console.log(data)
      try {
        const response = await updateOrganization(existingOrg.id, {
          paymentDetails: {
            cardName: data.cardHolderName as string,
            cardNumber: data.cardNumber as string,
            addressLineOne: data.billingAddressLineOne as string,
            addressLineTwo: data.billingAddressLineTwo as string,
            expirationMonth: data.expirationMonth as string,
            expirationYear: data.expirationYear as string,
            cvc: data.cvc as string,
          }
        });

        if (response.status < 200 || response.status >= 400) {
          throw new Error("non-200 status for update");
        }
      } catch (err) {
        console.error(err);
        setError("Failed to save update your payment details.");
      }
      toastSuccessMsg = "Your payment details has been updated.";
    }

    try {
      const responseData = await getOrganization(existingOrg.id);
      setExistingOrg(responseData);
    } catch (err) {
      console.error(err);
      setError("failed to reload organization data after update");
    } finally {
      setLoading(false);
    }

    toast.success(toastSuccessMsg);
  };

  if (error) {
    return (
      <MainContentCardLayout
        cardDescription="Details about current plan, payment method and billing history."
        cardTitle="Billing Information"
      >
        <Alert variant="destructive">
          <AlertCircle className="h-4 w-4" />
          <AlertTitle>Error</AlertTitle>
          <AlertDescription>Failed to load organization data.</AlertDescription>
        </Alert>
      </MainContentCardLayout>
    );
  }

  return (
    <MainContentCardLayout
      cardDescription="Details about current plan, payment method and billing history."
      cardTitle="Billing Information"
    >
      {loading ? (
        <div className="flex justify-center items-center h-screen">
          <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
        </div>
      ) : (
        <BillingDetails
          existingOrganization={existingOrg ?? ({} as GetOrganizationResponse)}
          onSubmit={handleBillingPageSave}
        />
      )}
    </MainContentCardLayout>
  );
}
