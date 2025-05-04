'use client'

import { BillingDetails } from '@/components/BillingDetails'
import MainContentCardLayout from '@/layouts/MainContentCardLayout'
import { toast } from 'sonner'

export default function BillingPage() {
  const handleBillingPageSave = async (
    formName: string,
    formData: FormData
  ) => {
    const data = Object.fromEntries(formData.entries())

    console.log(data) // For debugging

    let toastSuccessMsg = ''

    if (formName === 'primaryAdministratorForm') {
      console.log(data.primaryAdminEmail)
      toastSuccessMsg = 'Your primary administrator has been updated.'
    } else if (formName === 'billingPlanForm') {
      console.log(data.billingPlan)
      toastSuccessMsg = 'Your billing plan has been updated.'
    } else if (formName === 'creditCardForm') {
      toastSuccessMsg = 'Your payment method has been updated.'
    }

    toast.success(toastSuccessMsg)
  }

  return (
    <MainContentCardLayout
      cardDescription="Details about current plan, payment method and billing history."
      cardTitle="Billing Information"
    >
      <BillingDetails onSubmit={handleBillingPageSave} />
    </MainContentCardLayout>
  )
}
