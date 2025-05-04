import { AdministratorsTable } from '@/components/AdministratorsTable'
import MainContentCardLayout from '@/layouts/MainContentCardLayout'

export default function AdministratorsListPage() {
  return (
    <MainContentCardLayout
      cardDescription="Details below of administators in your organization."
      cardTitle="Administrators"
    >
      <AdministratorsTable />
    </MainContentCardLayout>
  )
}
