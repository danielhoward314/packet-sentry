import { DevicesTable } from "@/components/DevicesTable";
import MainContentCardLayout from "@/layouts/MainContentCardLayout";

export default function DevicesListPage() {
  return (
    <MainContentCardLayout
      cardDescription="The Packet Sentry Agent is installed on and reporting telemetry from these devices."
      cardTitle="Devices"
      cardFullWidth
    >
      <DevicesTable />
    </MainContentCardLayout>
  );
}
